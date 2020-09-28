// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package batch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/storage"
)

const (
	resyncPeriod = 5 * time.Second
	maxRetries   = 3
)

var statusMap = map[v1beta2.ApplicationStateType]models.State{
	v1beta2.NewState:              models.JobPending,
	v1beta2.SubmittedState:        models.JobPending,
	v1beta2.RunningState:          models.JobRunning,
	v1beta2.CompletedState:        models.JobCompleted,
	v1beta2.FailedState:           models.JobFailed,
	v1beta2.FailedSubmissionState: models.JobFailedSubmission,
	v1beta2.PendingRerunState:     models.JobRunning,
	v1beta2.InvalidatingState:     models.JobRunning,
	v1beta2.SucceedingState:       models.JobRunning,
	v1beta2.FailingState:          models.JobRunning,
	v1beta2.UnknownState:          models.JobRunning,
}

var BatchCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "batch_count",
		Namespace: "merlin_api",
		Help:      "Number of batch prediction execution",
	},
	[]string{"project", "model", "status"},
)

func init() {
	prometheus.MustRegister(BatchCounter)
}

// Controller abstract Kubernetes Controller -- https://kubernetes.io/docs/concepts/architecture/controller/ for
// submitting and tracking the status of spark application.
//
// Each environment/cluster has one controller and this controller got initialized and executed in cmd/main.go by calling Run().
// After execution, it will have the queue to process the next event (processNextItem()).
// This event is mainly the state change of the Spark application.
// Every time user deploys a new model prediction job in his targeted cluster, merlin-api calls Submit()
// then the controller for that cluster will create necessary resources and deploy Spark application.
// Every time that Spark application has a new status, it will be added to the queue to be processed by processNextItem().
// processNextItem() will do necessary jobs for updated Spark application (saving, cleaning)
type Controller interface {
	Submit(predictionJob *models.PredictionJob, namespace string) error
	Run(stopCh <-chan struct{})
	Stop(predictionJob *models.PredictionJob, namespace string) error
	cluster.ContainerFetcher
}

type controller struct {
	store            storage.PredictionJobStorage
	mlpApiClient     mlp.APIClient
	sparkClient      versioned.Interface
	kubeClient       kubernetes.Interface
	namespaceCreator cluster.NamespaceCreator
	manifestManager  ManifestManager
	informer         cache.SharedIndexInformer
	queue            workqueue.RateLimitingInterface

	cluster.ContainerFetcher
}

func NewController(store storage.PredictionJobStorage, mlpApiClient mlp.APIClient, sparkClient versioned.Interface, kubeClient kubernetes.Interface, manifestManager ManifestManager, envMetaData cluster.Metadata) Controller {
	informerFactory := externalversions.NewSharedInformerFactory(sparkClient, resyncPeriod)
	informer := informerFactory.Sparkoperator().V1beta2().SparkApplications().Informer()
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	controller := &controller{
		store:            store,
		mlpApiClient:     mlpApiClient,
		sparkClient:      sparkClient,
		kubeClient:       kubeClient,
		manifestManager:  manifestManager,
		namespaceCreator: cluster.NewNamespaceCreator(kubeClient.CoreV1(), time.Second*5),
		informer:         informer,
		queue:            queue,

		ContainerFetcher: cluster.NewContainerFetcher(kubeClient.CoreV1(), envMetaData),
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: controller.onUpdate,
	})
	return controller
}

func (c *controller) Submit(predictionJob *models.PredictionJob, namespace string) error {
	ctx := context.Background()

	var err error
	defer func() {
		if err != nil {
			// Directly cleanup if error happens during submission
			c.cleanup(predictionJob, namespace)
		}
	}()

	_, err = c.namespaceCreator.CreateNamespace(namespace)
	if err != nil {
		return fmt.Errorf("failed creating namespace %s: %v", namespace, err)
	}

	driverServiceAccount, err := c.manifestManager.CreateDriverAuthorization(namespace)
	if err != nil {
		return fmt.Errorf("failed creating spark driver authorization in namespace %s: %v", namespace, err)
	}

	secret, err := c.mlpApiClient.GetPlainSecretByNameAndProjectID(ctx, predictionJob.Config.ServiceAccountName, int32(predictionJob.ProjectId))
	if err != nil {
		return fmt.Errorf("service account %s is not found within %s project: %s", predictionJob.Config.ServiceAccountName, namespace, err)
	}

	_, err = c.manifestManager.CreateSecret(predictionJob.Name, namespace, secret.Data)
	if err != nil {
		return fmt.Errorf("failed creating secret for job %s in namespace %s: %v", predictionJob.Name, namespace, err)
	}

	_, err = c.manifestManager.CreateJobSpec(predictionJob.Name, namespace, predictionJob.Config.JobConfig)
	if err != nil {
		return fmt.Errorf("failed creating job specification configmap for job %s in namespace %s: %v", predictionJob.Name, namespace, err)
	}

	sparkResource, err := CreateSparkApplicationResource(predictionJob)
	if err != nil {
		return fmt.Errorf("failed creating spark application resource for job %s in namespace %s: %v", predictionJob.Name, namespace, err)
	}

	sparkResource.Spec.Driver.ServiceAccount = &driverServiceAccount
	_, err = c.sparkClient.SparkoperatorV1beta2().SparkApplications(namespace).Create(sparkResource)
	if err != nil {
		return fmt.Errorf("failed submitting spark application to spark controller for job %s in namespace %s: %v", predictionJob.Name, namespace, err)
	}

	return c.store.Save(predictionJob)
}

func (c *controller) Run(stopCh <-chan struct{}) {
	defer c.queue.ShutDown()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.hasSynced) {
		log.Errorf("timed out while waiting for cache to sync")
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *controller) Stop(predictionJob *models.PredictionJob, namespace string) error {
	sparkResources, _ := c.sparkClient.SparkoperatorV1beta2().SparkApplications(namespace).List(v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelPredictionJobId, predictionJob.Id.String()),
	})

	if len(sparkResources.Items) == 0 {
		return fmt.Errorf("unable to retrieve spark application of prediction job id %s from spark client", predictionJob.Id.String())
	}

	for _, resource := range sparkResources.Items {
		err := c.sparkClient.SparkoperatorV1beta2().SparkApplications(namespace).Delete(resource.Name, v1.NewDeleteOptions(0))
		if err != nil {
			return fmt.Errorf("failed to delete spark application resource %s for job %s in namespace %s: %v", resource.Name, predictionJob.Name, namespace, err)
		}
	}

	c.cleanup(predictionJob, namespace)
	predictionJob.Status = models.JobTerminated

	return c.store.Save(predictionJob)
}

func (c *controller) cleanup(job *models.PredictionJob, namespace string) {
	err := c.manifestManager.DeleteSecret(job.Name, namespace)
	if err != nil {
		log.Warnf("failed deleting secret %s in namespace %s: %v", job.Name, namespace, err)
	}

	err = c.manifestManager.DeleteJobSpec(job.Name, namespace)
	if err != nil {
		log.Warnf("failed deleting job spec %s in namespace %s: %v", job.Name, namespace, err)
	}
}

// hasSynced is required for the cache.Controller interface.
func (c *controller) hasSynced() bool {
	return c.informer.HasSynced()
}

func (c *controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *controller) processNextItem() bool {
	// Get will block until next event is received
	key, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(key)
	err := c.syncStatus(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		// Retry
		log.Warnf("error processing event, retry attempt : %d", c.queue.NumRequeues(key))
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		log.Warnf("error processing event, discarding event: %v", key)
		c.queue.Forget(key)
	}

	return true
}

func (c *controller) syncStatus(key string) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}

	sparkApp, _ := obj.(*v1beta2.SparkApplication)
	predictionJobId, err := models.ParseId(sparkApp.Labels[labelPredictionJobId])
	if err != nil {
		return errors.Wrapf(err, "unable to parse prediction job id")
	}
	predictionJob, err := c.store.Get(predictionJobId)
	if err != nil {
		return errors.Wrapf(err, "unable to find prediction job with id: %s", predictionJobId)
	}

	if predictionJob.Status != models.JobTerminated {
		predictionJob.Status = statusMap[sparkApp.Status.AppState.State]
		predictionJob.Error = sparkApp.Status.AppState.ErrorMessage
	}

	if predictionJob.Status.IsTerminal() {
		c.cleanup(predictionJob, sparkApp.Namespace)
		modelName := getModelName(predictionJob.Name)
		BatchCounter.WithLabelValues(sparkApp.Namespace, modelName, string(predictionJob.Status)).Inc()
	}

	return c.store.Save(predictionJob)
}

func (c *controller) onUpdate(old, new interface{}) {
	oldApp, _ := old.(*v1beta2.SparkApplication)
	newApp, _ := new.(*v1beta2.SparkApplication)
	if oldApp.Status.AppState.State != newApp.Status.AppState.State {
		key, _ := cache.MetaNamespaceKeyFunc(newApp)
		c.queue.AddRateLimited(key)
	}
}

func getModelName(projectionJobName string) string {
	s := strings.Split(projectionJobName, "-")
	return strings.Join(s[:len(s)-2], "-")
}
