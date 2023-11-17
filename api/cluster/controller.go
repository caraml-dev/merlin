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

package cluster

import (
	"context"
	"fmt"
	"io"
	"time"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kservev1beta1client "github.com/kserve/kserve/pkg/client/clientset/versioned/typed/serving/v1beta1"
	"github.com/pkg/errors"
	networkingv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	policyv1client "k8s.io/client-go/kubernetes/typed/policy/v1"
	"k8s.io/client-go/rest"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	knservingclient "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/caraml-dev/merlin/cluster/resource"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/deployment"
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
)

type Controller interface {
	Deploy(ctx context.Context, modelService *models.Service) (*models.Service, error)
	Delete(ctx context.Context, modelService *models.Service) (*models.Service, error)

	ListPods(ctx context.Context, namespace, labelSelector string) (*corev1.PodList, error)
	StreamPodLogs(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error)

	ListJobs(ctx context.Context, namespace, labelSelector string) (*batchv1.JobList, error)
	DeleteJob(ctx context.Context, namespace, jobName string, deleteOptions metav1.DeleteOptions) error
	DeleteJobs(ctx context.Context, namespace string, deleteOptions metav1.DeleteOptions, listOptions metav1.ListOptions) error

	GetCurrentDeploymentScale(ctx context.Context, namespace string,
		components map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec) resource.DeploymentScale

	ContainerFetcher
}

// Config Model cluster authentication settings
type Config struct {
	// Cluster Name
	ClusterName string
	// GCP project where the cluster resides
	GcpProject string
	// Use Kubernetes service account in cluster config
	InClusterConfig bool

	// Alternative to CACert, ClientCert info
	mlpcluster.Credentials
}

const (
	tickDurationSecond        = 1
	deletionGracePeriodSecond = 30
)

type controller struct {
	knServingClient            knservingclient.ServingV1Interface
	kserveClient               kservev1beta1client.ServingV1beta1Interface
	clusterClient              corev1client.CoreV1Interface
	batchClient                batchv1client.BatchV1Interface
	policyClient               policyv1client.PolicyV1Interface
	istioClient                networkingv1beta1.NetworkingV1beta1Interface
	namespaceCreator           NamespaceCreator
	deploymentConfig           *config.DeploymentConfig
	kfServingResourceTemplater *resource.InferenceServiceTemplater
	ContainerFetcher
}

func NewController(clusterConfig Config, deployConfig config.DeploymentConfig) (Controller, error) {
	var cfg *rest.Config
	var err error
	if clusterConfig.InClusterConfig {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clusterConfig.Credentials.ToRestConfig()
	}
	if err != nil {
		return nil, err
	}

	knsClientSet, err := knservingclientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	kserveClient, err := kservev1beta1client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	coreV1Client, err := corev1client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	batchV1Client, err := batchv1client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	policyV1Client, err := policyv1client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	istioClient, err := networkingv1beta1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	containerFetcher := NewContainerFetcher(coreV1Client, Metadata{
		ClusterName: clusterConfig.ClusterName,
		GcpProject:  clusterConfig.GcpProject,
	})

	kfServingResourceTemplater := resource.NewInferenceServiceTemplater(deployConfig)
	return newController(
		knsClientSet.ServingV1(),
		kserveClient,
		coreV1Client,
		batchV1Client,
		policyV1Client,
		istioClient,
		deployConfig,
		containerFetcher,
		kfServingResourceTemplater,
	)
}

func newController(
	knServingClient knservingclient.ServingV1Interface,
	kfkserveClient kservev1beta1client.ServingV1beta1Interface,
	coreV1Client corev1client.CoreV1Interface,
	batchV1Client batchv1client.BatchV1Interface,
	policyV1Client policyv1client.PolicyV1Interface,
	istioClient networkingv1beta1.NetworkingV1beta1Interface,
	deploymentConfig config.DeploymentConfig,
	containerFetcher ContainerFetcher,
	templater *resource.InferenceServiceTemplater,
) (Controller, error) {
	return &controller{
		knServingClient:            knServingClient,
		kserveClient:               kfkserveClient,
		clusterClient:              coreV1Client,
		batchClient:                batchV1Client,
		policyClient:               policyV1Client,
		istioClient:                istioClient,
		namespaceCreator:           NewNamespaceCreator(coreV1Client, deploymentConfig.NamespaceTimeout),
		deploymentConfig:           &deploymentConfig,
		ContainerFetcher:           containerFetcher,
		kfServingResourceTemplater: templater,
	}, nil
}

func (c *controller) Deploy(ctx context.Context, modelService *models.Service) (*models.Service, error) {
	if modelService.ResourceRequest != nil {
		cpuRequest, _ := modelService.ResourceRequest.CPURequest.AsInt64()
		maxCPU, _ := c.deploymentConfig.MaxCPU.AsInt64()
		if cpuRequest > maxCPU {
			log.Errorf("insufficient available cpu resource to fulfil user request of %d", cpuRequest)
			return nil, ErrInsufficientCPU
		}
		memRequest, _ := modelService.ResourceRequest.MemoryRequest.AsInt64()
		maxMem, _ := c.deploymentConfig.MaxMemory.AsInt64()
		if memRequest > maxMem {
			log.Errorf("insufficient available memory resource to fulfil user request of %d", memRequest)
			return nil, ErrInsufficientMem
		}
	}

	_, err := c.namespaceCreator.CreateNamespace(ctx, modelService.Namespace)
	if err != nil {
		log.Errorf("unable to create namespace %s %v", modelService.Namespace, err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateNamespace, modelService.Namespace))
	}

	isvcName := modelService.Name

	// Get current scale of the existing deployment
	deploymentScale := resource.DeploymentScale{}
	if modelService.CurrentIsvcName != "" {
		if modelService.DeploymentMode == deployment.ServerlessDeploymentMode ||
			modelService.DeploymentMode == deployment.EmptyDeploymentMode {
			currentIsvc, err := c.kserveClient.InferenceServices(modelService.Namespace).Get(modelService.CurrentIsvcName, metav1.GetOptions{})
			if err != nil && !kerrors.IsNotFound(err) {
				return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToGetInferenceServiceStatus, isvcName))
			}

			deploymentScale = c.GetCurrentDeploymentScale(ctx, modelService.Namespace, currentIsvc.Status.Components)
		}
	}

	// create new resource
	spec, err := c.kfServingResourceTemplater.CreateInferenceServiceSpec(modelService, deploymentScale)
	if err != nil {
		log.Errorf("unable to create inference service spec %s: %v", isvcName, err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateInferenceService, isvcName))
	}

	s, err := c.kserveClient.InferenceServices(modelService.Namespace).Create(spec)
	if err != nil {
		log.Errorf("unable to create inference service %s: %v", isvcName, err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateInferenceService, isvcName))
	}

	if c.deploymentConfig.PodDisruptionBudget.Enabled {
		pdbs := createPodDisruptionBudgets(modelService, c.deploymentConfig.PodDisruptionBudget)
		if err := c.deployPodDisruptionBudgets(ctx, pdbs); err != nil {
			log.Errorf("unable to create pdb: %v", err)
			return nil, errors.Wrapf(err, fmt.Sprintf("%v", ErrUnableToCreatePDB))
		}
	}

	s, err = c.waitInferenceServiceReady(s)
	if err != nil {
		// remove created inferenceservice when got error
		if err := c.deleteInferenceService(isvcName, modelService.Namespace); err != nil {
			log.Errorf("unable to delete inference service %s with error %v", isvcName, err)
		}

		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToGetInferenceServiceStatus, isvcName))
	}

	inferenceURL := models.GetInferenceURL(s.Status.URL, isvcName, modelService.Protocol)

	// Create / update virtual service
	vsCfg, err := NewVirtualService(modelService, inferenceURL)
	if err != nil {
		log.Errorf("unable to initialize virtual service builder: %v", err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v", ErrUnableToCreateVirtualService))
	}

	vs, err := c.deployVirtualService(ctx, vsCfg)
	if err != nil {
		log.Errorf("unable to create virtual service: %v", err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateVirtualService, vsCfg.Name))
	}

	if vs != nil && len(vs.Spec.Hosts) > 0 {
		inferenceURL = vsCfg.getInferenceURL(vs)
	}

	// Delete previous inference service
	if modelService.CurrentIsvcName != "" {
		if err := c.deleteInferenceService(modelService.CurrentIsvcName, modelService.Namespace); err != nil {
			log.Errorf("unable to delete prevision revision %s with error %v", modelService.CurrentIsvcName, err)
			return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToDeletePreviousInferenceService, modelService.CurrentIsvcName))
		}
	}

	return &models.Service{
		Name:            s.Name,
		Namespace:       s.Namespace,
		ServiceName:     s.Status.URL.Host,
		URL:             inferenceURL,
		Metadata:        modelService.Metadata,
		CurrentIsvcName: s.Name,
	}, nil
}

func (c *controller) Delete(ctx context.Context, modelService *models.Service) (*models.Service, error) {
	infSvc, err := c.kserveClient.InferenceServices(modelService.Namespace).Get(modelService.Name, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "unable to check status of inference service: %s", infSvc.Name)
		}
		return modelService, nil
	}

	if err := c.deleteInferenceService(modelService.Name, modelService.Namespace); err != nil {
		return nil, err
	}

	if c.deploymentConfig.PodDisruptionBudget.Enabled {
		pdbs := createPodDisruptionBudgets(modelService, c.deploymentConfig.PodDisruptionBudget)
		if err := c.deletePodDisruptionBudgets(ctx, pdbs); err != nil {
			log.Errorf("unable to delete pdb %v", err)
			return nil, ErrUnableToDeletePDB
		}
	}

	if modelService.RevisionID > 1 {
		vsName := fmt.Sprintf("%s-%s-%s", modelService.ModelName, modelService.ModelVersion, models.VirtualServiceComponentType)
		if err := c.deleteVirtualService(ctx, vsName, modelService.Namespace); err != nil {
			log.Errorf("unable to delete virtual service %v", err)
			return nil, ErrUnableToDeleteVirtualService
		}
	}

	return modelService, nil
}

func (c *controller) deleteInferenceService(serviceName string, namespace string) error {
	gracePeriod := int64(deletionGracePeriodSecond)
	err := c.kserveClient.InferenceServices(namespace).Delete(serviceName, &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrapf(err, "unable to delete inference service: %s %v", serviceName, err)
	}
	return nil
}

func (c *controller) waitInferenceServiceReady(service *kservev1beta1.InferenceService) (*kservev1beta1.InferenceService, error) {
	timeout := time.After(c.deploymentConfig.DeploymentTimeout)
	ticker := time.NewTicker(time.Second * tickDurationSecond)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for inference service to be ready %s", service.Name)
			return nil, ErrTimeoutCreateInferenceService
		case <-ticker.C:
			s, err := c.kserveClient.InferenceServices(service.Namespace).Get(service.Name, metav1.GetOptions{})
			if err != nil {
				log.Errorf("unable to get inference service status %s %v", service.Name, err)
				return nil, ErrUnableToGetInferenceServiceStatus
			}

			if s.Status.IsReady() {
				// Inference service is completely ready
				return s, nil
			}
		}
	}
}

func (c *controller) ListPods(ctx context.Context, namespace, labelSelector string) (*corev1.PodList, error) {
	return c.clusterClient.Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
}

func (c *controller) StreamPodLogs(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	return c.clusterClient.Pods(namespace).GetLogs(podName, opts).Stream(ctx)
}

func (c *controller) GetCurrentDeploymentScale(
	ctx context.Context,
	namespace string,
	components map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec,
) resource.DeploymentScale {
	deploymentScale := resource.DeploymentScale{}

	// Init knative revisions getter
	revisions := c.knServingClient.Revisions(namespace)

	// Get predictor scale
	if rev, err := revisions.Get(ctx, components[kservev1beta1.PredictorComponent].LatestCreatedRevision, metav1.GetOptions{}); err == nil {
		if rev.Status.DesiredReplicas != nil {
			predictorScale := int(*rev.Status.DesiredReplicas)
			deploymentScale.Predictor = &predictorScale
		}
	}

	// Get transformer scale, if enabled
	if _, ok := components[kservev1beta1.TransformerComponent]; ok {
		if rev, err := revisions.Get(ctx, components[kservev1beta1.TransformerComponent].LatestCreatedRevision, metav1.GetOptions{}); err == nil {
			if rev.Status.DesiredReplicas != nil {
				transformerScale := int(*rev.Status.DesiredReplicas)
				deploymentScale.Transformer = &transformerScale
			}
		}
	}

	return deploymentScale
}
