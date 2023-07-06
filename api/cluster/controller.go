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
	"io"
	"time"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kservev1beta1client "github.com/kserve/kserve/pkg/client/clientset/versioned/typed/serving/v1beta1"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	policyv1client "k8s.io/client-go/kubernetes/typed/policy/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/caraml-dev/merlin/cluster/resource"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
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
	servingClient              kservev1beta1client.ServingV1beta1Interface
	clusterClient              corev1client.CoreV1Interface
	batchClient                batchv1client.BatchV1Interface
	policyClient               policyv1client.PolicyV1Interface
	namespaceCreator           NamespaceCreator
	deploymentConfig           *config.DeploymentConfig
	kfServingResourceTemplater *resource.InferenceServiceTemplater
	ContainerFetcher
}

func NewController(clusterConfig Config, deployConfig config.DeploymentConfig, standardTransformerConfig config.StandardTransformerConfig) (Controller, error) {
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
	servingClient, err := kservev1beta1client.NewForConfig(cfg)
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

	containerFetcher := NewContainerFetcher(coreV1Client, Metadata{
		ClusterName: clusterConfig.ClusterName,
		GcpProject:  clusterConfig.GcpProject,
	})

	kfServingResourceTemplater := resource.NewInferenceServiceTemplater(standardTransformerConfig)
	return newController(servingClient, coreV1Client, batchV1Client, policyV1Client, deployConfig, containerFetcher, kfServingResourceTemplater)
}

func newController(kfservingClient kservev1beta1client.ServingV1beta1Interface,
	coreV1Client corev1client.CoreV1Interface,
	batchV1Client batchv1client.BatchV1Interface,
	policyV1Client policyv1client.PolicyV1Interface,
	deploymentConfig config.DeploymentConfig,
	containerFetcher ContainerFetcher,
	templater *resource.InferenceServiceTemplater,
) (Controller, error) {
	return &controller{
		servingClient:              kfservingClient,
		clusterClient:              coreV1Client,
		batchClient:                batchV1Client,
		policyClient:               policyV1Client,
		namespaceCreator:           NewNamespaceCreator(coreV1Client, deploymentConfig.NamespaceTimeout),
		deploymentConfig:           &deploymentConfig,
		ContainerFetcher:           containerFetcher,
		kfServingResourceTemplater: templater,
	}, nil
}

func (k *controller) Deploy(ctx context.Context, modelService *models.Service) (*models.Service, error) {
	if modelService.ResourceRequest != nil {
		cpuRequest, _ := modelService.ResourceRequest.CPURequest.AsInt64()
		maxCPU, _ := k.deploymentConfig.MaxCPU.AsInt64()
		if cpuRequest > maxCPU {
			log.Errorf("insufficient available cpu resource to fulfil user request of %d", cpuRequest)
			return nil, ErrInsufficientCPU
		}
		memRequest, _ := modelService.ResourceRequest.MemoryRequest.AsInt64()
		maxMem, _ := k.deploymentConfig.MaxMemory.AsInt64()
		if memRequest > maxMem {
			log.Errorf("insufficient available memory resource to fulfil user request of %d", memRequest)
			return nil, ErrInsufficientMem
		}
	}

	_, err := k.namespaceCreator.CreateNamespace(ctx, modelService.Namespace)
	if err != nil {
		log.Errorf("unable to create namespace %s %v", modelService.Namespace, err)
		return nil, ErrUnableToCreateNamespace
	}

	isvcName := modelService.Name
	s, err := k.servingClient.InferenceServices(modelService.Namespace).Get(isvcName, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("unable to check inference service %s %v", isvcName, err)
			return nil, ErrUnableToGetInferenceServiceStatus
		}

		// create new resource
		spec, err := k.kfServingResourceTemplater.CreateInferenceServiceSpec(modelService, k.deploymentConfig)
		if err != nil {
			log.Errorf("unable to create inference service spec %s %v", isvcName, err)
			return nil, ErrUnableToCreateInferenceService
		}

		s, err = k.servingClient.InferenceServices(modelService.Namespace).Create(spec)
		if err != nil {
			log.Errorf("unable to create inference service %s %v", isvcName, err)
			return nil, ErrUnableToCreateInferenceService
		}
	} else {
		patchedSpec, err := k.kfServingResourceTemplater.PatchInferenceServiceSpec(s, modelService, k.deploymentConfig)
		if err != nil {
			log.Errorf("unable to update inference service %s %v", isvcName, err)
			return nil, ErrUnableToUpdateInferenceService
		}

		// existing resource found, do update
		s, err = k.servingClient.InferenceServices(modelService.Namespace).Update(patchedSpec)
		if err != nil {
			log.Errorf("unable to update inference service %s %v", isvcName, err)
			return nil, ErrUnableToUpdateInferenceService
		}
	}

	if k.deploymentConfig.PodDisruptionBudget.Enabled {
		pdbs := k.createPodDisruptionBudgets(modelService, k.deploymentConfig.PodDisruptionBudget)
		if err := k.deployPodDisruptionBudgets(ctx, pdbs); err != nil {
			log.Errorf("unable to create pdb %v", err)
			return nil, ErrUnableToCreatePDB
		}
	}

	s, err = k.waitInferenceServiceReady(s)
	if err != nil {
		// remove created inferenceservice when got error
		if err := k.deleteInferenceService(isvcName, modelService.Namespace); err != nil {
			log.Warnf("unable to delete inference service %s with error %v", isvcName, err)
		}

		return nil, err
	}

	inferenceURL := models.GetInferenceURL(s.Status.URL, isvcName, modelService.Protocol)
	return &models.Service{
		Name:        s.Name,
		Namespace:   s.Namespace,
		ServiceName: s.Status.URL.Host,
		URL:         inferenceURL,
		Metadata:    modelService.Metadata,
	}, nil
}

func (k *controller) Delete(ctx context.Context, modelService *models.Service) (*models.Service, error) {
	infSvc, err := k.servingClient.InferenceServices(modelService.Namespace).Get(modelService.Name, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "unable to check status of inference service: %s", infSvc.Name)
		}
		return modelService, nil
	}

	if err := k.deleteInferenceService(modelService.Name, modelService.Namespace); err != nil {
		return nil, err
	}

	if k.deploymentConfig.PodDisruptionBudget.Enabled {
		pdbs := k.createPodDisruptionBudgets(modelService, k.deploymentConfig.PodDisruptionBudget)
		if err := k.deletePodDisruptionBudgets(ctx, pdbs); err != nil {
			log.Errorf("unable to delete pdb %v", err)
			return nil, ErrUnableToDeletePDB
		}
	}

	return modelService, nil
}

func (k *controller) deleteInferenceService(serviceName string, namespace string) error {
	gracePeriod := int64(deletionGracePeriodSecond)
	err := k.servingClient.InferenceServices(namespace).Delete(serviceName, &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrapf(err, "unable to delete inference service: %s %v", serviceName, err)
	}
	return nil
}

func (k *controller) waitInferenceServiceReady(service *kservev1beta1.InferenceService) (*kservev1beta1.InferenceService, error) {
	timeout := time.After(k.deploymentConfig.DeploymentTimeout)
	ticker := time.NewTicker(time.Second * tickDurationSecond)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for inference service to be ready %s", service.Name)
			return nil, ErrTimeoutCreateInferenceService
		case <-ticker.C:
			s, err := k.servingClient.InferenceServices(service.Namespace).Get(service.Name, metav1.GetOptions{})
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
