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
	"errors"
	"fmt"
	"testing"
	"time"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	fakekserve "github.com/kserve/kserve/pkg/client/clientset/versioned/fake"
	fakekservev1beta1 "github.com/kserve/kserve/pkg/client/clientset/versioned/typed/serving/v1beta1/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	fakepolicyv1 "k8s.io/client-go/kubernetes/typed/policy/v1/fake"
	k8stesting "k8s.io/client-go/testing"
	ktesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/network"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingfake "knative.dev/serving/pkg/client/clientset/versioned/fake"

	clusterresource "github.com/caraml-dev/merlin/cluster/resource"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
)

const (
	listMethod             = "list"
	getMethod              = "get"
	createMethod           = "create"
	updateMethod           = "update"
	deleteMethod           = "delete"
	deleteCollectionMethod = "delete-collection"

	kfservingGroup           = "kubeflow.com/kfserving"
	knativeGroup             = "serving.knative.dev"
	knativeVersion           = "v1"
	inferenceServiceResource = "inferenceservices"
	revisionResource         = "revisions"

	coreGroup         = ""
	namespaceResource = "namespaces"
	podResource       = "pods"
	jobResource       = "jobs"
	pdbResource       = "poddisruptionbudgets"

	baseUrl = "example.com"
)

type namespaceReactor struct {
	namespace *corev1.Namespace
	err       error
}

type inferenceServiceReactor struct {
	isvc *kservev1beta1.InferenceService
	err  error
}

type knativeRevisionReactor struct {
	rev *knservingv1.Revision
	err error
}

var clusterMetadata = Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}

// TestDeployInferenceServiceNamespaceCreation test namespaceResource creation when deploying inference service
func TestController_DeployInferenceService_NamespaceCreation(t *testing.T) {
	nsTimeout := 2 * tickDurationSecond * time.Second
	model := &models.Model{
		Name: "my-model",
	}
	project := mlp.Project{
		Name: "my-project",
	}
	version := &models.Version{
		ID: 1,
	}
	modelOpt := &models.ModelOption{}
	isvc := fakeInferenceService(model.Name, version.ID.String(), project.Name)

	modelSvc := &models.Service{
		Name:      isvc.Name,
		Namespace: project.Name,
		Options:   modelOpt,
	}

	tests := []struct {
		name         string
		getResult    *namespaceReactor
		createResult *namespaceReactor
		checkResult  *namespaceReactor
		nsTimeout    time.Duration
		wantError    bool
	}{
		{
			"success: create namespaceResource",
			&namespaceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: coreGroup, Resource: namespaceResource}, project.Name),
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceActive,
					},
				},
				nil,
			},
			nsTimeout,
			false,
		},
		{
			"success: namespaceResource exists",
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceActive,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceActive,
					},
				},
				nil,
			},
			nsTimeout,
			false,
		},
		{
			"error: namespaceResource terminating",
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceTerminating,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceTerminating,
					},
				},
				nil,
			},
			nsTimeout,
			true,
		},
		{
			"error: namespaceResource terminating",
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceTerminating,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
				},
				nil,
			},
			&namespaceReactor{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: project.Name,
					},
					Status: corev1.NamespaceStatus{
						Phase: corev1.NamespaceTerminating,
					},
				},
				nil,
			},
			nsTimeout,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knClient := knservingfake.NewSimpleClientset().ServingV1()
			kfClient := fakekserve.NewSimpleClientset().ServingV1beta1().(*fakekservev1beta1.FakeServingV1beta1)
			kfClient.PrependReactor(getMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				kfClient.PrependReactor(getMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, isvc, nil
				})
				return true, nil, kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvc.Name)
			})
			kfClient.PrependReactor(createMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, isvc, nil
			})

			v1Client := fake.NewSimpleClientset().CoreV1()
			nsClient := v1Client.Namespaces().(*fakecorev1.FakeNamespaces)
			nsClient.Fake.PrependReactor(getMethod, namespaceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				nsClient.Fake.PrependReactor(getMethod, namespaceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, tt.checkResult.namespace, tt.checkResult.err
				})
				return true, tt.getResult.namespace, tt.getResult.err
			})
			nsClient.Fake.PrependReactor(createMethod, namespaceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.createResult.namespace, tt.createResult.err
			})

			policyV1Client := fake.NewSimpleClientset().PolicyV1()

			deployConfig := config.DeploymentConfig{
				NamespaceTimeout:             tt.nsTimeout,
				DeploymentTimeout:            2 * tickDurationSecond * time.Second,
				DefaultModelResourceRequests: &config.ResourceRequests{},
			}

			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
			ctl, _ := newController(knClient, kfClient, v1Client, nil, policyV1Client, deployConfig, containerFetcher, nil)
			iSvc, err := ctl.Deploy(context.Background(), modelSvc)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, iSvc)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, iSvc)
		})
	}
}

func TestController_DeployInferenceService(t *testing.T) {
	defaultMaxUnavailablePDB := 20

	deployTimeout := 2 * tickDurationSecond * time.Second
	model := &models.Model{
		Name: "my-model",
	}
	project := mlp.Project{
		Name: "my-project",
	}
	version := &models.Version{
		ID: 1,
	}
	modelOpt := &models.ModelOption{}
	isvcName := models.CreateInferenceServiceName(model.Name, version.ID.String())
	statusReady := createServiceReadyStatus(isvcName, project.Name, baseUrl)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: project.Name},
		Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	}
	pdb := &policyv1.PodDisruptionBudget{}

	modelSvc := &models.Service{
		Name:      isvcName,
		Namespace: project.Name,
		Options:   modelOpt,
	}

	tests := []struct {
		name          string
		modelService  *models.Service
		getRevResult  *knativeRevisionReactor
		getResult     *inferenceServiceReactor
		createResult  *inferenceServiceReactor
		updateResult  *inferenceServiceReactor
		checkResult   *inferenceServiceReactor
		deployTimeout time.Duration
		wantError     bool
	}{
		{
			"success: create inference service",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			false,
		},
		{
			"success: update inference service",
			modelSvc,
			&knativeRevisionReactor{err: kerrors.NewNotFound(schema.GroupResource{}, "test service")},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			false,
		},
		{
			"success: deploying service",
			&models.Service{
				Name:      isvcName,
				Namespace: project.Name,
				Options:   modelOpt,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("100m"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			false,
		},
		{
			"success: create inference service with transformer",
			&models.Service{
				Name:      isvcName,
				Namespace: project.Name,
				Options:   modelOpt,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "ghcr.io/caraml-dev/merlin-transformer-test",
				},
			},
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			false,
		},
		{
			"error: failed get",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				errors.New("error"),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			true,
		},
		{
			"error: failed create",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: "kubeflow.com/kfserving", Resource: "inferenceservices"}, isvcName),
			},
			&inferenceServiceReactor{
				nil,
				errors.New("error creating inference service"),
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			true,
		},
		{
			"error: failed update",
			modelSvc,
			&knativeRevisionReactor{err: kerrors.NewNotFound(schema.GroupResource{}, "test service")},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				nil,
				errors.New("error updating inference service"),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			true,
		},
		{
			"error: failed check",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: "kubeflow.com/kfserving", Resource: "inferenceservices"}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				nil,
				errors.New("error check"),
			},
			deployTimeout,
			true,
		},
		{
			"error: predictor error",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: "kubeflow.com/kfserving", Resource: "inferenceservices"}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     createPredErrorCond(),
				},
				nil,
			},
			deployTimeout,
			true,
		},
		{
			"error: routes error",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: "kubeflow.com/kfserving", Resource: "inferenceservices"}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     createRoutesErrorCond(),
				},
				nil,
			},
			deployTimeout,
			true,
		},
		{
			"error: timeout",
			modelSvc,
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			1 * time.Millisecond,
			true,
		},
		{
			"error: deploying service due to insufficient CPU",
			&models.Service{
				Name:      isvcName,
				Namespace: project.Name,
				Options:   modelOpt,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("10"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			true,
		},
		{
			"error: deploying service due to insufficient memory",
			&models.Service{
				Name:      isvcName,
				Namespace: project.Name,
				Options:   modelOpt,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1000m"),
					MemoryRequest: resource.MustParse("10Gi"),
				},
			},
			&knativeRevisionReactor{},
			&inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			nil,
			&inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knClient := knservingfake.NewSimpleClientset()
			knClient.PrependReactor(getMethod, revisionResource, func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, tt.getRevResult.rev, tt.getRevResult.err
			})

			kfClient := fakekserve.NewSimpleClientset().ServingV1beta1().(*fakekservev1beta1.FakeServingV1beta1)
			kfClient.PrependReactor(getMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				kfClient.PrependReactor(getMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, tt.checkResult.isvc, tt.checkResult.err
				})
				return true, tt.getResult.isvc, tt.getResult.err
			})
			kfClient.PrependReactor(createMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.createResult.isvc, tt.createResult.err
			})
			kfClient.PrependReactor(updateMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.updateResult.isvc, tt.updateResult.err
			})

			kfClient.PrependReactor(deleteMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			v1Client := fake.NewSimpleClientset().CoreV1()
			nsClient := v1Client.Namespaces().(*fakecorev1.FakeNamespaces)
			nsClient.Fake.PrependReactor(getMethod, namespaceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, namespace, nil
			})

			policyV1Client := fake.NewSimpleClientset().PolicyV1().(*fakepolicyv1.FakePolicyV1)
			policyV1Client.Fake.PrependReactor("patch", pdbResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, pdb, nil
			})

			deployConfig := config.DeploymentConfig{
				DeploymentTimeout:                  tt.deployTimeout,
				NamespaceTimeout:                   2 * tickDurationSecond * time.Second,
				MaxCPU:                             resource.MustParse("8"),
				MaxMemory:                          resource.MustParse("8Gi"),
				DefaultModelResourceRequests:       &config.ResourceRequests{},
				DefaultTransformerResourceRequests: &config.ResourceRequests{},
				PodDisruptionBudget: config.PodDisruptionBudgetConfig{
					Enabled:                  true,
					MaxUnavailablePercentage: &defaultMaxUnavailablePDB,
				},
			}

			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
			templater := clusterresource.NewInferenceServiceTemplater(config.StandardTransformerConfig{
				ImageName:             "ghcr.io/caraml-dev/merlin-transformer-test",
				FeastServingKeepAlive: &config.FeastServingKeepAliveConfig{},
			})

			ctl, _ := newController(knClient.ServingV1(), kfClient, v1Client, nil, policyV1Client, deployConfig, containerFetcher, templater)
			iSvc, err := ctl.Deploy(context.Background(), tt.modelService)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, iSvc)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, iSvc)
		})
	}
}

func TestGetCurrentDeploymentScale(t *testing.T) {
	testNamespace := "test-namespace"
	var testDesiredReplicas int32 = 5
	var testDesiredReplicasInt int = 5

	resourceItem := schema.GroupVersionResource{
		Group:    knativeGroup,
		Version:  knativeVersion,
		Resource: revisionResource,
	}

	// Define tests
	tests := map[string]struct {
		components    map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec
		rFunc         func(action k8stesting.Action) (bool, runtime.Object, error)
		expectedScale clusterresource.DeploymentScale
	}{
		"failure | revision not found": {
			components: map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec{
				kservev1beta1.PredictorComponent: {
					LatestCreatedRevision: "test-predictor-0",
				},
			},
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(resourceItem, testNamespace, "test-predictor-0")
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Return nil object and error to indicate non existent object
				return true, nil, kerrors.NewNotFound(schema.GroupResource{}, "test-predictor-0")
			},
			expectedScale: clusterresource.DeploymentScale{},
		},
		"failure | desired replicas not set": {
			components: map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec{
				kservev1beta1.PredictorComponent: {
					LatestCreatedRevision: "test-predictor-0",
				},
			},
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(resourceItem, testNamespace, "test-predictor-0")
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Return test response
				return true, &knservingv1.Revision{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-predictor-0",
					},
				}, nil
			},
			expectedScale: clusterresource.DeploymentScale{},
		},
		"success | predictor only": {
			components: map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec{
				kservev1beta1.PredictorComponent: {
					LatestCreatedRevision: "test-predictor-0",
				},
			},
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(resourceItem, testNamespace, "test-predictor-0")
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Return test response
				return true, &knservingv1.Revision{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-predictor-0",
					},
					Status: knservingv1.RevisionStatus{
						DesiredReplicas: &testDesiredReplicas,
					},
				}, nil
			},
			expectedScale: clusterresource.DeploymentScale{Predictor: &testDesiredReplicasInt},
		},
		"success | predictor and transformer": {
			components: map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec{
				kservev1beta1.PredictorComponent: {
					LatestCreatedRevision: "test-svc-0",
				},
				kservev1beta1.TransformerComponent: {
					LatestCreatedRevision: "test-svc-0",
				},
			},
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(resourceItem, testNamespace, "test-svc-0")
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Return test response
				return true, &knservingv1.Revision{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-svc-0",
					},
					Status: knservingv1.RevisionStatus{
						DesiredReplicas: &testDesiredReplicas,
					},
				}, nil
			},
			expectedScale: clusterresource.DeploymentScale{
				Predictor:   &testDesiredReplicasInt,
				Transformer: &testDesiredReplicasInt,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			knClient := knservingfake.NewSimpleClientset()
			knClient.PrependReactor(getMethod, revisionResource, tt.rFunc)

			kfClient := fakekserve.NewSimpleClientset().ServingV1beta1().(*fakekservev1beta1.FakeServingV1beta1)
			v1Client := fake.NewSimpleClientset().CoreV1()
			policyV1Client := fake.NewSimpleClientset().PolicyV1().(*fakepolicyv1.FakePolicyV1)

			deployConfig := config.DeploymentConfig{}
			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
			templater := clusterresource.NewInferenceServiceTemplater(config.StandardTransformerConfig{
				ImageName:             "ghcr.io/caraml-dev/merlin-transformer-test",
				FeastServingKeepAlive: &config.FeastServingKeepAliveConfig{},
			})

			// Create test controller
			ctl, _ := newController(knClient.ServingV1(), kfClient, v1Client, nil, policyV1Client, deployConfig, containerFetcher, templater)

			desiredReplicas := ctl.GetCurrentDeploymentScale(context.TODO(), testNamespace, tt.components)
			assert.Equal(t, tt.expectedScale, desiredReplicas)
		})
	}
}

func fakeInferenceService(model, version, project string) *kservev1beta1.InferenceService {
	svcName := models.CreateInferenceServiceName(model, version)
	status := createServiceReadyStatus(svcName, project, baseUrl)
	return &kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: project}, Status: status}
}

func createServiceReadyStatus(iSvcName, namespace, baseUrl string) kservev1beta1.InferenceServiceStatus {
	status := kservev1beta1.InferenceServiceStatus{}
	status.InitializeConditions()

	url, err := apis.ParseURL(fmt.Sprintf("%s.%s.%s", iSvcName, namespace, baseUrl))
	if err != nil {
		panic(err)
	}
	status.URL = url
	status.Address = &duckv1.Addressable{
		URL: &apis.URL{
			Host:   network.GetServiceHostname(iSvcName, namespace),
			Scheme: "http",
		},
	}
	status.SetConditions(apis.Conditions{
		{
			Type:   apis.ConditionReady,
			Status: corev1.ConditionTrue,
		},
	})
	return status
}

func createPredErrorCond() kservev1beta1.InferenceServiceStatus {
	status := kservev1beta1.InferenceServiceStatus{}
	status.InitializeConditions()
	status.SetConditions(apis.Conditions{
		{
			Type:   kservev1beta1.IngressReady,
			Status: corev1.ConditionTrue,
		},
		{
			Type:    kservev1beta1.PredictorReady,
			Status:  corev1.ConditionFalse,
			Message: "predictor error",
		},
		{
			Type:   apis.ConditionReady,
			Status: corev1.ConditionFalse,
		},
	})
	return status
}

func createRoutesErrorCond() kservev1beta1.InferenceServiceStatus {
	status := kservev1beta1.InferenceServiceStatus{}
	status.InitializeConditions()
	status.SetConditions(apis.Conditions{
		{
			Type:    kservev1beta1.IngressReady,
			Status:  corev1.ConditionFalse,
			Message: "routes error",
		},
		{
			Type:   kservev1beta1.PredictorReady,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   apis.ConditionReady,
			Status: corev1.ConditionFalse,
		},
	})
	return status
}

func isIn(container corev1.Container, containers []*models.Container, podName string) bool {
	for _, c := range containers {
		if container.Name == c.Name && podName == c.PodName {
			return true
		}
	}
	return false
}

func Test_controller_ListPods(t *testing.T) {
	namespace := "test-namespace"

	v1Client := fake.NewSimpleClientset()
	v1Client.PrependReactor(listMethod, podResource, func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, &corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-model-1-predictor-default-a",
						Labels: map[string]string{
							"serving.knative.dev/service": "test-model-1-predictor-default",
						},
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							{Name: "storage-initializer"},
						},
						Containers: []corev1.Container{
							{Name: "kfserving-container"},
							{Name: "queue-proxy"},
							{Name: "inferenceservice-logger"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-model-1-predictor-default-b",
						Labels: map[string]string{
							"serving.knative.dev/service": "test-model-1-predictor-default",
						},
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							{Name: "storage-initializer"},
						},
						Containers: []corev1.Container{
							{Name: "kfserving-container"},
							{Name: "queue-proxy"},
							{Name: "inferenceservice-logger"},
						},
					},
				},
			},
		}, nil
	})

	ctl := &controller{
		clusterClient: v1Client.CoreV1(),
	}

	podList, err := ctl.ListPods(context.Background(), namespace, "serving.knative.dev/service=test-model-1-predictor-default")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(podList.Items))
	assert.Equal(t, "test-model-1-predictor-default-a", podList.Items[0].ObjectMeta.Name)
	assert.Equal(t, "test-model-1-predictor-default-b", podList.Items[1].ObjectMeta.Name)
}
