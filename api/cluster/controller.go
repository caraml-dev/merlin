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
	"time"

	kfsv1alpha2 "github.com/kubeflow/kfserving/pkg/apis/serving/v1alpha2"
	kfservice "github.com/kubeflow/kfserving/pkg/client/clientset/versioned/typed/serving/v1alpha2"
	"github.com/kubeflow/kfserving/pkg/constants"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
)

type Controller interface {
	Deploy(modelService *models.Service, transformer models.Transformer) (*models.Service, error)
	Delete(modelService *models.Service) (*models.Service, error)
	ContainerFetcher
}

// ClusterConfig Model cluster authentication settings
type ClusterConfig struct {
	// Kubernetes API server endpoint
	Host string
	// CA Certificate to trust for TLS
	CACert string
	// Client Certificate for authenticating to cluster
	ClientCert string
	// Client Key for authenticating to cluster
	ClientKey string

	// Cluster Name
	ClusterName string
	// GCP project where the cluster resides
	GcpProject string
}

const (
	tickDurationSecond        = 1
	deletionGracePeriodSecond = 30
)

type controller struct {
	servingClient    kfservice.ServingV1alpha2Interface
	clusterClient    corev1.CoreV1Interface
	namespaceCreator NamespaceCreator
	config           *config.DeploymentConfig
	ContainerFetcher
}

func NewController(clusterConfig ClusterConfig, deployConfig config.DeploymentConfig) (Controller, error) {
	cfg := &rest.Config{
		Host: clusterConfig.Host,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   []byte(clusterConfig.CACert),
			CertData: []byte(clusterConfig.ClientCert),
			KeyData:  []byte(clusterConfig.ClientKey),
		},
	}

	servingClient, err := kfservice.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	coreV1Client, err := corev1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	containerFetcher := NewContainerFetcher(coreV1Client, Metadata{
		ClusterName: clusterConfig.ClusterName,
		GcpProject:  clusterConfig.GcpProject,
	})

	return newController(servingClient, coreV1Client, deployConfig, containerFetcher)
}

func newController(kfservingClient kfservice.ServingV1alpha2Interface, nsClient corev1.CoreV1Interface, deploymentConfig config.DeploymentConfig, containerFetcher ContainerFetcher) (Controller, error) {
	return &controller{
		servingClient:    kfservingClient,
		clusterClient:    nsClient,
		namespaceCreator: NewNamespaceCreator(nsClient, deploymentConfig.NamespaceTimeout),
		config:           &deploymentConfig,
		ContainerFetcher: containerFetcher,
	}, nil
}

func (k *controller) Deploy(modelService *models.Service, transformer models.Transformer) (*models.Service, error) {
	if modelService.ResourceRequest != nil {
		cpuRequest, _ := modelService.ResourceRequest.CpuRequest.AsInt64()
		maxCpu, _ := k.config.MaxCpu.AsInt64()
		if cpuRequest > maxCpu {
			log.Errorf("insufficient available cpu resource to fulfil user request of %d", cpuRequest)
			return nil, ErrInsufficientCpu
		}
		memRequest, _ := modelService.ResourceRequest.MemoryRequest.AsInt64()
		maxMem, _ := k.config.MaxMemory.AsInt64()
		if memRequest > maxMem {
			log.Errorf("insufficient available memory resource to fulfil user request of %d", memRequest)
			return nil, ErrInsufficientMem
		}
	}

	_, err := k.namespaceCreator.CreateNamespace(modelService.Namespace)
	if err != nil {
		log.Errorf("unable to create namespace %s %v", modelService.Namespace, err)
		return nil, ErrUnableToCreateNamespace
	}

	svcName := modelService.Name
	s, err := k.servingClient.InferenceServices(modelService.Namespace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("unable to check inference service %s %v", svcName, err)
			return nil, ErrUnableToGetInferenceServiceStatus
		}

		// create new resource
		s, err = k.servingClient.InferenceServices(modelService.Namespace).Create(createInferenceServiceSpec(modelService, transformer, k.config))
		if err != nil {
			log.Errorf("unable to create inference service %s %v", svcName, err)
			return nil, ErrUnableToCreateInferenceService
		}
	} else {
		// existing resource found, do update
		s, err = k.servingClient.InferenceServices(modelService.Namespace).Update(patchInferenceServiceSpec(s, modelService, k.config))
		if err != nil {
			log.Errorf("unable to update inference service %s %v", svcName, err)
			return nil, ErrUnableToUpdateInferenceService
		}
	}

	s, err = k.waitInferenceServiceReady(s)
	if err != nil {
		return nil, err
	}

	return &models.Service{
		Name:        s.Name,
		Namespace:   s.Namespace,
		ServiceName: (*s.Status.Default)[constants.Predictor].Hostname,
		Url:         s.Status.URL,
	}, nil
}

func (k *controller) Delete(modelService *models.Service) (*models.Service, error) {
	infSvc, err := k.servingClient.InferenceServices(modelService.Namespace).Get(modelService.Name, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "unable to check status of inference service: %s", infSvc.Name)
		}
		return modelService, nil
	}
	gracePeriod := int64(deletionGracePeriodSecond)
	err = k.servingClient.InferenceServices(modelService.Namespace).Delete(modelService.Name, &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to delete inference service: %s", infSvc.Name)
	}

	return modelService, nil
}

func (k *controller) waitInferenceServiceReady(service *kfsv1alpha2.InferenceService) (*kfsv1alpha2.InferenceService, error) {
	timeout := time.After(k.config.DeploymentTimeout)
	ticker := time.Tick(time.Second * tickDurationSecond)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for inference service to be ready %s", service.Name)
			return nil, ErrTimeoutCreateInferenceService
		case <-ticker:
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
