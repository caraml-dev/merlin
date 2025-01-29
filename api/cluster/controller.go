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
	"sort"
	"time"

	"github.com/caraml-dev/merlin/mlp"
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
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	knservingclient "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/caraml-dev/merlin/cluster/resource"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/caraml-dev/merlin/utils"
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
)

type Controller interface {
	Deploy(ctx context.Context, modelService *models.Service, projectID int) (*models.Service, error)
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
	mlpAPIClient               mlp.APIClient
	ContainerFetcher
}

func NewController(clusterConfig Config, deployConfig config.DeploymentConfig, mlpAPIClient mlp.APIClient) (Controller, error) {
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
		mlpAPIClient,
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
	mlpAPIClient mlp.APIClient,
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
		mlpAPIClient:               mlpAPIClient,
	}, nil
}

func (c *controller) Deploy(ctx context.Context, modelService *models.Service, projectID int) (*models.Service, error) {
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
		if modelService.ResourceRequest.MaxReplica > c.deploymentConfig.MaxAllowedReplica {
			log.Errorf("Requested Max Replica (%d) is more than max permissible (%d)",
				modelService.ResourceRequest.MaxReplica,
				c.deploymentConfig.MaxAllowedReplica,
			)
			return nil, ErrRequestedMaxReplicasNotAllowed
		}
	}

	_, err := c.namespaceCreator.CreateNamespace(ctx, modelService.Namespace)
	if err != nil {
		log.Errorf("unable to create namespace %s %v", modelService.Namespace, err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateNamespace, modelService.Namespace))
	}

	err = c.createSecrets(ctx, modelService, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed creating secret for deployment %s: %w", modelService.Name, err)
	}

	isvcName := modelService.Name

	// Get current scale of the existing deployment
	deploymentScale := resource.DeploymentScale{}
	if modelService.CurrentIsvcName != "" {
		if modelService.DeploymentMode == deployment.ServerlessDeploymentMode ||
			modelService.DeploymentMode == deployment.EmptyDeploymentMode {
			currentIsvc, err := c.kserveClient.InferenceServices(modelService.Namespace).Get(ctx,
				modelService.CurrentIsvcName, metav1.GetOptions{})
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

	// check the cluster to see if the inference service has already been deployed
	s, err := c.kserveClient.InferenceServices(modelService.Namespace).Get(ctx, modelService.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			s, err = c.kserveClient.InferenceServices(modelService.Namespace).Create(ctx, spec, metav1.CreateOptions{})
			if err != nil {
				log.Errorf("unable to create inference service %s: %v", isvcName, err)
				return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateInferenceService, isvcName))
			}
		} else {
			return nil, errors.Wrapf(err, "unable to check status of inference service: %s", modelService.Name)
		}
	} else {
		log.Infof("found existing inference service %s; skipping its creation", isvcName)
	}

	if c.deploymentConfig.PodDisruptionBudget.Enabled {
		// Create / update pdb
		pdbs := generatePDBSpecs(modelService, c.deploymentConfig.PodDisruptionBudget)
		if err := c.deployPodDisruptionBudgets(ctx, pdbs); err != nil {
			log.Errorf("unable to create pdb: %v", err)
			return nil, errors.Wrapf(err, fmt.Sprintf("%v", ErrUnableToCreatePDB))
		}
	}

	s, err = c.waitInferenceServiceReady(s)
	if err != nil {
		// remove created inferenceservice when got error
		if err := c.deleteInferenceService(ctx, isvcName, modelService.Namespace); err != nil {
			log.Errorf("unable to delete inference service %s with error %v", isvcName, err)
		}

		if err := c.deleteSecrets(ctx, modelService); err != nil {
			log.Warnf("failed deleting secret for deployment %s: %w", modelService.Name, err)
		}

		return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToCreateInferenceService, isvcName))
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

	// Delete previous inference service and pdb
	if modelService.CurrentIsvcName != "" {
		if err := c.deleteInferenceService(ctx, modelService.CurrentIsvcName, modelService.Namespace); err != nil {
			log.Errorf("unable to delete prevision revision %s with error %v", modelService.CurrentIsvcName, err)
			return nil, errors.Wrapf(err, fmt.Sprintf("%v (%s)", ErrUnableToDeletePreviousInferenceService, modelService.CurrentIsvcName))
		}

		unusedPdbs, err := c.getUnusedPodDisruptionBudgets(ctx, modelService)
		if err != nil {
			log.Warnf("unable to get model name %s, version %s unused pdb: %v", modelService.ModelName, modelService.ModelVersion, err)
		} else {
			if err := c.deletePodDisruptionBudgets(ctx, unusedPdbs); err != nil {
				log.Warnf("unable to delete model name %s, version %s unused pdb: %v", modelService.ModelName, modelService.ModelVersion, err)
			}
		}

		if err := c.deleteSecrets(ctx, modelService); err != nil {
			log.Warnf("failed deleting secret for deployment %s: %w", modelService.Name, err)
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
	infSvc, err := c.kserveClient.InferenceServices(modelService.Namespace).Get(ctx, modelService.Name,
		metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "unable to check status of inference service: %s", infSvc.Name)
		}
		return modelService, nil
	}

	if err := c.deleteInferenceService(ctx, modelService.Name, modelService.Namespace); err != nil {
		return nil, err
	}

	if c.deploymentConfig.PodDisruptionBudget.Enabled {
		pdbs := generatePDBSpecs(modelService, c.deploymentConfig.PodDisruptionBudget)
		if err := c.deletePodDisruptionBudgets(ctx, pdbs); err != nil {
			log.Errorf("unable to delete pdb %v", err)
			return nil, ErrUnableToDeletePDB
		}
	}

	err = c.deleteSecrets(ctx, modelService)
	if err != nil {
		return nil, fmt.Errorf("failed deleting secret for deployment %s: %w", modelService.Name, err)
	}

	vsCfg, err := NewVirtualService(modelService, "")
	if err != nil {
		log.Errorf("unable to initialize virtual service builder: %v", err)
		return nil, errors.Wrapf(err, fmt.Sprintf("%v", ErrUnableToDeleteVirtualService))
	}

	if err := c.deleteVirtualService(ctx, vsCfg); err != nil {
		log.Errorf("unable to delete virtual service %v", err)
		return nil, ErrUnableToDeleteVirtualService
	}

	return modelService, nil
}

func (c *controller) deleteInferenceService(ctx context.Context, serviceName string, namespace string) error {
	gracePeriod := int64(deletionGracePeriodSecond)
	err := c.kserveClient.InferenceServices(namespace).Delete(ctx, serviceName,
		metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrapf(err, "unable to delete inference service: %s %v", serviceName, err)
	}
	return nil
}

func (c *controller) waitInferenceServiceReady(service *kservev1beta1.InferenceService) (isvc *kservev1beta1.InferenceService, err error) {
	ctx := context.Background()

	timeout := time.After(c.deploymentConfig.DeploymentTimeout)

	isvcConditionTable := ""
	podContainerTable := ""
	podLastTerminationMessage := ""
	podLastTerminationReason := ""

	defer func() {
		if err == nil {
			return
		}

		if isvcConditionTable != "" {
			err = fmt.Errorf("%w\n\nModel service conditions:\n%s", err, isvcConditionTable)
		}

		if podContainerTable != "" {
			err = fmt.Errorf("%w\n\nPod container status:\n%s", err, podContainerTable)
		}

		if podLastTerminationMessage != "" {
			err = fmt.Errorf("%w\n\nPod last termination message:\n%s", err, podLastTerminationMessage)
		}
	}()

	isvcWatcher, err := c.kserveClient.InferenceServices(service.Namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", service.Name),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize isvc watcher: %s", service.Name)
	}

	podWatcher, err := c.clusterClient.Pods(service.Namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("serving.kserve.io/inferenceservice=%s", service.Name),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize isvc's pods watcher: %s", service.Name)
	}

	for {
		select {
		case <-timeout:
			return nil, ErrTimeoutCreateInferenceService

		case isvcEvent := <-isvcWatcher.ResultChan():
			isvc, ok := isvcEvent.Object.(*kservev1beta1.InferenceService)
			if !ok {
				return nil, errors.New("unable to cast isvc object")
			}
			log.Debugf("isvc event received [%s]", isvc.Name)

			if isvc.Status.Status.Conditions != nil {
				// Update isvc condition table with latest conditions
				isvcConditionTable, err = parseInferenceServiceConditions(isvc.Status.Status.Conditions)
			}

			if isvc.Status.IsReady() {
				return isvc, nil
			}

		case podEvent := <-podWatcher.ResultChan():
			pod, ok := podEvent.Object.(*corev1.Pod)
			if !ok {
				return nil, errors.New("unable to cast pod object")
			}
			log.Debugf("pod event received [%s]", pod.Name)

			if len(pod.Status.ContainerStatuses) > 0 {
				// Update pod container table with latest container statuses
				podContainerTable, podLastTerminationMessage, podLastTerminationReason = utils.ParsePodContainerStatuses(pod.Status.ContainerStatuses)
				err = errors.New(podLastTerminationReason)
			}
		}
	}
}

func parseInferenceServiceConditions(isvcConditions duckv1.Conditions) (string, error) {
	var err error

	isvcConditionHeaders := []string{"TYPE", "STATUS", "REASON", "MESSAGE"}
	isvcConditionRows := [][]string{}

	sort.Slice(isvcConditions, func(i, j int) bool {
		return isvcConditions[i].LastTransitionTime.Inner.Before(&isvcConditions[j].LastTransitionTime.Inner)
	})

	for _, condition := range isvcConditions {
		isvcConditionRows = append(isvcConditionRows, []string{
			string(condition.Type),
			string(condition.Status),
			condition.Reason,
			condition.Message,
		})

		err = errors.New(condition.Reason)
		if condition.Message != "" {
			err = fmt.Errorf("%w: %s", err, condition.Message)
		}
	}

	isvcTable := utils.LogTable(isvcConditionHeaders, isvcConditionRows)
	return isvcTable, err
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

func (c *controller) createSecrets(ctx context.Context, modelService *models.Service, projectID int) error {
	err := c.createSecretForComponent(ctx, modelService.Name, modelService.Secrets, modelService.Namespace, projectID)
	if err != nil {
		return fmt.Errorf("failed creating secret for model %s in namespace %s: %w", modelService.Name, modelService.Namespace, err)
	}

	if modelService.Transformer != nil {
		transformerSecretName := fmt.Sprintf("%s-transformer", modelService.Name)
		err = c.createSecretForComponent(ctx, transformerSecretName, modelService.Transformer.Secrets, modelService.Namespace, projectID)
		if err != nil {
			return fmt.Errorf("failed creating secret for transformer %s in namespace %s: %w", transformerSecretName, modelService.Namespace, err)
		}
	}
	return nil
}

func (c *controller) createSecretForComponent(ctx context.Context, componentName string, secrets models.Secrets, namespace string, projectID int) error {
	// Get MLP secrets for Google service account and user-specified MLP secrets
	secretMap, err := c.getMLPSecrets(ctx, secrets, namespace, projectID)
	if err != nil {
		return fmt.Errorf("error retrieving secrets: %w", err)
	}

	_, err = c.createSecret(ctx, componentName, namespace, secretMap)
	if err != nil {
		return fmt.Errorf("failed creating secret %s in namespace %s: %w", componentName, namespace, err)
	}
	return nil
}

func (c *controller) deleteSecrets(ctx context.Context, modelService *models.Service) error {
	err := c.deleteSecret(ctx, modelService.Name, modelService.Namespace)
	if err != nil && !kerrors.IsNotFound(err) {
		return fmt.Errorf("failed deleting secret for model %s in namespace %s: %w", modelService.Name, modelService.Namespace, err)
	}

	if modelService.Transformer != nil {
		transformerSecretName := fmt.Sprintf("%s-transformer", modelService.Name)
		err = c.deleteSecret(ctx, transformerSecretName, modelService.Namespace)
		if err != nil && !kerrors.IsNotFound(err) {
			return fmt.Errorf("failed deleting secret for transformer %s in namespace %s: %w", transformerSecretName, modelService.Namespace, err)
		}
	}
	return nil
}

func (c *controller) getMLPSecrets(ctx context.Context, secrets models.Secrets, namespace string, projectID int) (map[string]string, error) {
	secretMap := make(map[string]string)
	// Retrieve user-configured secrets from MLP
	for _, secret := range secrets {
		userConfiguredSecret, err := c.mlpAPIClient.GetSecretByName(ctx, secret.MLPSecretName, int32(projectID))
		if err != nil {
			return nil, fmt.Errorf("user-configured secret %s is not found within %s project: %w", secret.MLPSecretName, namespace, err)
		}
		secretMap[secret.MLPSecretName] = userConfiguredSecret.Data
	}

	return secretMap, nil
}
