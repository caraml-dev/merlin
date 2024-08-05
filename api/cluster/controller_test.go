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
	istiov1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	fakeistionetworking "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1/fake"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	fakepolicyv1 "k8s.io/client-go/kubernetes/typed/policy/v1/fake"
	ktesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/network"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingfake "knative.dev/serving/pkg/client/clientset/versioned/fake"

	clusterresource "github.com/caraml-dev/merlin/cluster/resource"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
)

const (
	listMethod             = "list"
	getMethod              = "get"
	createMethod           = "create"
	patchMethod            = "patch"
	updateMethod           = "update"
	deleteMethod           = "delete"
	deleteCollectionMethod = "delete-collection"

	kfservingGroup           = "kubeflow.com/kfserving"
	knativeGroup             = "serving.knative.dev"
	knativeVersion           = "v1"
	inferenceServiceResource = "inferenceservices"
	revisionResource         = "revisions"
	virtualServiceResource   = "virtualservices"

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

type pdbReactor struct {
	pdb *policyv1.PodDisruptionBudget
	err error
}

type vsReactor struct {
	vs  *istiov1beta1.VirtualService
	err error
}

type isvcWatchReactor struct {
	result chan watch.Event
}

func newIsvcWatchReactor(isvc *kservev1beta1.InferenceService) *isvcWatchReactor {
	w := &isvcWatchReactor{result: make(chan watch.Event, 1)}
	w.result <- watch.Event{Type: watch.Added, Object: isvc}
	return w
}

func (w *isvcWatchReactor) Handles(action ktesting.Action) bool {
	return action.GetResource().Resource == inferenceServiceResource
}

func (w *isvcWatchReactor) React(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
	return true, watch.NewProxyWatcher(w.result), nil
}

var (
	clusterMetadata = Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}

	userContainerCPUDefaultLimit          = "8"
	userContainerCPULimitRequestFactor    = float64(0)
	userContainerMemoryLimitRequestFactor = float64(2)
)

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
	revisionID := models.ID(1)
	modelOpt := &models.ModelOption{}

	isvc := fakeInferenceService(model.Name, version.ID.String(), revisionID.String(), project.Name)
	vs := fakeVirtualService(model.Name, version.ID.String())

	modelSvc := &models.Service{
		Name:         isvc.Name,
		ModelName:    model.Name,
		ModelVersion: version.ID.String(),
		RevisionID:   revisionID,
		Namespace:    project.Name,
		Options:      modelOpt,
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
			kfClient.PrependReactor(createMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, isvc, nil
			})
			isvcWatchReactor := newIsvcWatchReactor(isvc)
			kfClient.WatchReactionChain = []ktesting.WatchReactor{isvcWatchReactor}

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

			istioClient := fakeistio.NewSimpleClientset().NetworkingV1beta1().(*fakeistionetworking.FakeNetworkingV1beta1)
			istioClient.PrependReactor(patchMethod, virtualServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, vs, nil
			})

			deployConfig := config.DeploymentConfig{
				NamespaceTimeout:                      tt.nsTimeout,
				DeploymentTimeout:                     2 * tickDurationSecond * time.Second,
				DefaultModelResourceRequests:          &config.ResourceRequests{},
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
			}

			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)

			ctl, _ := newController(knClient, kfClient, v1Client, nil, policyV1Client, istioClient, deployConfig, containerFetcher, clusterresource.NewInferenceServiceTemplater(deployConfig))
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
	minAvailablePercentage := 80
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
	revisionID := models.ID(1)
	modelOpt := &models.ModelOption{}

	isvcName := models.CreateInferenceServiceName(model.Name, version.ID.String(), revisionID.String())
	statusReady := createServiceReadyStatus(isvcName, project.Name, baseUrl)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: project.Name},
		Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	}
	pdb := &policyv1.PodDisruptionBudget{}
	vs := fakeVirtualService(model.Name, version.ID.String())

	modelSvc := &models.Service{
		Name:         isvcName,
		ModelName:    model.Name,
		ModelVersion: version.ID.String(),
		RevisionID:   revisionID,
		Namespace:    project.Name,
		Options:      modelOpt,
	}

	tests := []struct {
		name            string
		modelService    *models.Service
		createResult    *inferenceServiceReactor
		checkResult     *inferenceServiceReactor
		createPdbResult *pdbReactor
		createVsResult  *vsReactor
		deployTimeout   time.Duration
		wantError       bool
	}{
		{
			name:         "success: create inference service",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       false,
		},
		{
			name: "success: deploying service",
			modelService: &models.Service{
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
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       false,
		},
		{
			name: "success: create inference service with transformer",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: project.Name,
				Options:   modelOpt,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "ghcr.io/caraml-dev/merlin-transformer-test",
				},
			},
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       false,
		},
		{
			name:         "error: failed create",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				nil,
				errors.New("error creating inference service"),
			},
			checkResult: &inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{}, ""),
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name:         "error: failed check",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				nil,
				errors.New("error check"),
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name:         "error: predictor error",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     createPredErrorCond(),
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name:         "error: routes error",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     createRoutesErrorCond(),
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name:         "error: pdb error",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     createRoutesErrorCond(),
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{nil, ErrUnableToCreatePDB},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name:         "error: vs error",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     createRoutesErrorCond(),
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{nil, ErrUnableToCreateVirtualService},
			wantError:       true,
		},
		{
			name:         "error: isvc timeout",
			modelService: modelSvc,
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout: 1 * time.Microsecond,
			wantError:     true,
		},
		{
			name: "error: deploying service due to insufficient CPU",
			modelService: &models.Service{
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
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name: "error: deploying service due to insufficient memory",
			modelService: &models.Service{
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
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
		{
			name: "error: deploying service due to max replica requests greater than max value allowed",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: project.Name,
				Options:   modelOpt,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    5,
					CPURequest:    resource.MustParse("1000m"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			createResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name}},
				nil,
			},
			checkResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
					Status:     statusReady,
				},
				nil,
			},
			deployTimeout:   deployTimeout,
			createPdbResult: &pdbReactor{pdb, nil},
			createVsResult:  &vsReactor{vs, nil},
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knClient := knservingfake.NewSimpleClientset()

			kfClient := fakekserve.NewSimpleClientset().ServingV1beta1().(*fakekservev1beta1.FakeServingV1beta1)
			if tt.checkResult != nil && tt.checkResult.isvc != nil {
				isvcWatchReactor := newIsvcWatchReactor(tt.checkResult.isvc)
				kfClient.WatchReactionChain = []ktesting.WatchReactor{isvcWatchReactor}
			}
			kfClient.PrependReactor(createMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.createResult.isvc, tt.createResult.err
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
			policyV1Client.Fake.PrependReactor(patchMethod, pdbResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.createPdbResult.pdb, tt.createPdbResult.err
			})

			istioClient := fakeistio.NewSimpleClientset().NetworkingV1beta1().(*fakeistionetworking.FakeNetworkingV1beta1)
			if tt.createVsResult != nil && tt.createVsResult.vs != nil {
				istioClient.PrependReactor(patchMethod, virtualServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, tt.createVsResult.vs, tt.createVsResult.err
				})
			}

			deployConfig := config.DeploymentConfig{
				DeploymentTimeout:                  tt.deployTimeout,
				NamespaceTimeout:                   2 * tickDurationSecond * time.Second,
				MaxCPU:                             resource.MustParse("8"),
				MaxMemory:                          resource.MustParse("8Gi"),
				MaxAllowedReplica:                  4,
				DefaultModelResourceRequests:       &config.ResourceRequests{},
				DefaultTransformerResourceRequests: &config.ResourceRequests{},
				PodDisruptionBudget: config.PodDisruptionBudgetConfig{
					Enabled:                true,
					MinAvailablePercentage: &minAvailablePercentage,
				},
				StandardTransformer: config.StandardTransformerConfig{
					ImageName:             "ghcr.io/caraml-dev/merlin-transformer-test",
					FeastServingKeepAlive: &config.FeastServingKeepAliveConfig{},
				},
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
			}

			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
			templater := clusterresource.NewInferenceServiceTemplater(deployConfig)

			ctl, _ := newController(knClient.ServingV1(), kfClient, v1Client, nil, policyV1Client, istioClient, deployConfig, containerFetcher, templater)
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

func TestController_DeployInferenceService_PodDisruptionBudgetsRemoval(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", "dev")
	assert.Nil(t, err)

	minAvailablePercentage := 80

	model := &models.Model{
		Name: "my-model",
	}
	project := mlp.Project{
		Name: "my-project",
	}
	version := &models.Version{
		ID: 1,
	}
	revisionID := models.ID(5)

	isvcName := models.CreateInferenceServiceName(model.Name, version.ID.String(), revisionID.String())
	statusReady := createServiceReadyStatus(isvcName, project.Name, baseUrl)
	statusReady.Components = make(map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec)
	pdb := &policyv1.PodDisruptionBudget{}
	vs := fakeVirtualService(model.Name, version.ID.String())

	metadata := models.Metadata{
		App:       "mymodel",
		Component: models.ComponentModelVersion,
		Stream:    "mystream",
		Team:      "myteam",
	}
	modelSvc := &models.Service{
		Name:            isvcName,
		ModelName:       model.Name,
		ModelVersion:    version.ID.String(),
		RevisionID:      revisionID,
		Namespace:       project.Name,
		Metadata:        metadata,
		CurrentIsvcName: isvcName,
	}

	defaultMatchLabel := map[string]string{
		"gojek.com/app":                      "mymodel",
		"gojek.com/component":                "model-version",
		"gojek.com/environment":              "dev",
		"gojek.com/orchestrator":             "merlin",
		"gojek.com/stream":                   "mystream",
		"gojek.com/team":                     "myteam",
		"serving.kserve.io/inferenceservice": "mymodel-1-r4",
		"model-version-id":                   "1",
	}

	tests := []struct {
		name             string
		modelService     *models.Service
		existingPdbs     *policyv1.PodDisruptionBudgetList
		createPdbResult  *pdbReactor
		deletedPdbResult *pdbReactor
		wantPdbs         []string
	}{
		{
			name:         "success: remove pdbs",
			modelService: modelSvc,
			existingPdbs: &policyv1.PodDisruptionBudgetList{
				Items: []policyv1.PodDisruptionBudget{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-model-1-r4-predictor-pdb",
							Labels:    defaultMatchLabel,
							Namespace: modelSvc.Namespace,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-model-1-r4-transformer-pdb",
							Labels:    defaultMatchLabel,
							Namespace: modelSvc.Namespace,
						},
					},
				},
			},
			createPdbResult:  &pdbReactor{pdb, nil},
			deletedPdbResult: &pdbReactor{err: nil},
			wantPdbs: []string{
				"my-model-1-r4-predictor-pdb",
				"my-model-1-r4-transformer-pdb",
			},
		},
		{
			name:         "success: only remove pdb with same labels",
			modelService: modelSvc,
			existingPdbs: &policyv1.PodDisruptionBudgetList{
				Items: []policyv1.PodDisruptionBudget{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-model-1-r4-predictor-pdb",
							Labels:    defaultMatchLabel,
							Namespace: modelSvc.Namespace,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-model-2-r4-predictor-pdb",
							Labels: map[string]string{
								"gojek.com/app":                      "mymodel",
								"gojek.com/component":                "model-version",
								"gojek.com/environment":              "dev",
								"gojek.com/orchestrator":             "merlin",
								"gojek.com/stream":                   "mystream",
								"gojek.com/team":                     "myteam",
								"component":                          "predictor",
								"serving.kserve.io/inferenceservice": "mymodel-2-r4",
								"model-version-id":                   "2",
							},
							Namespace: modelSvc.Namespace,
						},
					},
				},
			},
			createPdbResult:  &pdbReactor{pdb, nil},
			deletedPdbResult: &pdbReactor{err: nil},
			wantPdbs: []string{
				"my-model-1-r4-predictor-pdb",
			},
		},
		{
			name:         "success: no pdb found or to delete",
			modelService: modelSvc,
			existingPdbs: &policyv1.PodDisruptionBudgetList{
				Items: []policyv1.PodDisruptionBudget{},
			},
			createPdbResult:  &pdbReactor{pdb, nil},
			deletedPdbResult: &pdbReactor{err: nil},
			wantPdbs:         []string{},
		},
		{
			name:         "fail deleting pdb should not return error",
			modelService: modelSvc,
			existingPdbs: &policyv1.PodDisruptionBudgetList{
				Items: []policyv1.PodDisruptionBudget{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-model-1-r4-predictor-pdb",
							Labels:    defaultMatchLabel,
							Namespace: modelSvc.Namespace,
						},
					},
				},
			},
			createPdbResult:  &pdbReactor{pdb, nil},
			deletedPdbResult: &pdbReactor{err: ErrUnableToDeletePDB},
			wantPdbs:         []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knClient := knservingfake.NewSimpleClientset()

			kfClient := fakekserve.NewSimpleClientset().ServingV1beta1().(*fakekservev1beta1.FakeServingV1beta1)
			isvcWatchReactor := newIsvcWatchReactor(&kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: isvcName, Namespace: project.Name},
				Status:     statusReady,
			})
			kfClient.WatchReactionChain = []ktesting.WatchReactor{isvcWatchReactor}

			kfClient.PrependReactor(getMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kservev1beta1.InferenceService{Status: statusReady}, nil
			})

			v1Client := fake.NewSimpleClientset().CoreV1()

			policyV1Client := fake.NewSimpleClientset(tt.existingPdbs).PolicyV1().(*fakepolicyv1.FakePolicyV1)
			policyV1Client.Fake.PrependReactor(patchMethod, pdbResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.createPdbResult.pdb, tt.createPdbResult.err
			})

			deletedPdbNames := []string{}
			policyV1Client.Fake.PrependReactor(deleteMethod, pdbResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				deletedPdbNames = append(deletedPdbNames, action.(ktesting.DeleteAction).GetName())
				return true, nil, tt.deletedPdbResult.err
			})

			istioClient := fakeistio.NewSimpleClientset().NetworkingV1beta1().(*fakeistionetworking.FakeNetworkingV1beta1)
			istioClient.PrependReactor(patchMethod, virtualServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, vs, nil
			})

			deployConfig := config.DeploymentConfig{
				DeploymentTimeout:                  2 * tickDurationSecond * time.Second,
				NamespaceTimeout:                   2 * tickDurationSecond * time.Second,
				DefaultModelResourceRequests:       &config.ResourceRequests{},
				DefaultTransformerResourceRequests: &config.ResourceRequests{},
				MaxAllowedReplica:                  4,
				PodDisruptionBudget: config.PodDisruptionBudgetConfig{
					Enabled:                true,
					MinAvailablePercentage: &minAvailablePercentage,
				},
				StandardTransformer:                   config.StandardTransformerConfig{},
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
			}

			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
			templater := clusterresource.NewInferenceServiceTemplater(deployConfig)

			ctl, _ := newController(knClient.ServingV1(), kfClient, v1Client, nil, policyV1Client, istioClient, deployConfig, containerFetcher, templater)
			iSvc, err := ctl.Deploy(context.Background(), tt.modelService)

			if tt.deletedPdbResult.err == nil {
				assert.ElementsMatch(t, deletedPdbNames, tt.wantPdbs)
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

	tests := map[string]struct {
		components    map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec
		rFunc         func(action ktesting.Action) (bool, runtime.Object, error)
		expectedScale clusterresource.DeploymentScale
	}{
		"failure | revision not found": {
			components: map[kservev1beta1.ComponentType]kservev1beta1.ComponentStatusSpec{
				kservev1beta1.PredictorComponent: {
					LatestCreatedRevision: "test-predictor-0",
				},
			},
			rFunc: func(action ktesting.Action) (bool, runtime.Object, error) {
				expAction := ktesting.NewGetAction(resourceItem, testNamespace, "test-predictor-0")
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
			rFunc: func(action ktesting.Action) (bool, runtime.Object, error) {
				expAction := ktesting.NewGetAction(resourceItem, testNamespace, "test-predictor-0")
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
			rFunc: func(action ktesting.Action) (bool, runtime.Object, error) {
				expAction := ktesting.NewGetAction(resourceItem, testNamespace, "test-predictor-0")
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
			rFunc: func(action ktesting.Action) (bool, runtime.Object, error) {
				expAction := ktesting.NewGetAction(resourceItem, testNamespace, "test-svc-0")
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

			deployConfig := config.DeploymentConfig{
				StandardTransformer: config.StandardTransformerConfig{
					ImageName:             "ghcr.io/caraml-dev/merlin-transformer-test",
					FeastServingKeepAlive: &config.FeastServingKeepAliveConfig{},
				},
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
			}
			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
			templater := clusterresource.NewInferenceServiceTemplater(deployConfig)

			// Create test controller
			ctl, _ := newController(knClient.ServingV1(), kfClient, v1Client, nil, policyV1Client, nil, deployConfig, containerFetcher, templater)

			desiredReplicas := ctl.GetCurrentDeploymentScale(context.TODO(), testNamespace, tt.components)
			assert.Equal(t, tt.expectedScale, desiredReplicas)
		})
	}
}

func fakeInferenceService(model, version, revisionID, project string) *kservev1beta1.InferenceService {
	svcName := models.CreateInferenceServiceName(model, version, revisionID)
	status := createServiceReadyStatus(svcName, project, baseUrl)
	return &kservev1beta1.InferenceService{ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: project}, Status: status}
}

func fakeVirtualService(model, version string) *istiov1beta1.VirtualService {
	return &istiov1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", model, version)}}
}

func createServiceReadyStatus(iSvcName, namespace, baseUrl string) kservev1beta1.InferenceServiceStatus {
	status := kservev1beta1.InferenceServiceStatus{}
	status.InitializeConditions()

	url, err := apis.ParseURL(fmt.Sprintf("%s.%s.%s", iSvcName, namespace, baseUrl))
	if err != nil {
		log.Panicf(err.Error())
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
						Name: "test-model-1-predictor-a",
						Labels: map[string]string{
							"serving.knative.dev/service": "test-model-1-predictor",
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
						Name: "test-model-1-predictor-b",
						Labels: map[string]string{
							"serving.knative.dev/service": "test-model-1-predictor",
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

	podList, err := ctl.ListPods(context.Background(), namespace, "serving.knative.dev/service=test-model-1-predictor")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(podList.Items))
	assert.Equal(t, "test-model-1-predictor-a", podList.Items[0].ObjectMeta.Name)
	assert.Equal(t, "test-model-1-predictor-b", podList.Items[1].ObjectMeta.Name)
}

func TestController_Delete(t *testing.T) {
	isvcName := models.CreateInferenceServiceName("my-model", "1", "1")
	projectName := "my-project"
	pdb := &policyv1.PodDisruptionBudget{}
	vs := fakeVirtualService("my-model", "1")

	tests := []struct {
		name         string
		modelService *models.Service
		getResult    *inferenceServiceReactor
		deleteResult *inferenceServiceReactor
		deployConfig config.DeploymentConfig
		wantError    bool
	}{
		{
			name: "success: delete predictor",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: projectName,
			},
			getResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deleteResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deployConfig: config.DeploymentConfig{},
			wantError:    false,
		},
		{
			name: "success: delete predictor and transformer",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: projectName,
				Transformer: &models.Transformer{
					Enabled: true,
				},
			},
			getResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deleteResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deployConfig: config.DeploymentConfig{},
			wantError:    false,
		},
		{
			name: "success: delete predictor and its pdb",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: projectName,
			},
			getResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deleteResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deployConfig: config.DeploymentConfig{
				PodDisruptionBudget: config.PodDisruptionBudgetConfig{
					Enabled: true,
				},
			},
			wantError: false,
		},
		{
			name: "success: delete predictor, transformer, and their pdb",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: projectName,
				Transformer: &models.Transformer{
					Enabled: true,
				},
			},
			getResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deleteResult: &inferenceServiceReactor{
				&kservev1beta1.InferenceService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      isvcName,
						Namespace: projectName,
					},
				},
				nil,
			},
			deployConfig: config.DeploymentConfig{
				PodDisruptionBudget: config.PodDisruptionBudgetConfig{
					Enabled: true,
				},
			},
			wantError: false,
		},
		{
			name: "skip: predictor not found",
			modelService: &models.Service{
				Name:      isvcName,
				Namespace: projectName,
			},
			getResult: &inferenceServiceReactor{
				nil,
				kerrors.NewNotFound(schema.GroupResource{Group: kfservingGroup, Resource: inferenceServiceResource}, isvcName),
			},
			deleteResult: &inferenceServiceReactor{
				nil,
				nil,
			},
			deployConfig: config.DeploymentConfig{},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knClient := knservingfake.NewSimpleClientset().ServingV1()
			kfClient := fakekserve.NewSimpleClientset().ServingV1beta1().(*fakekservev1beta1.FakeServingV1beta1)
			kfClient.PrependReactor(getMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.getResult.isvc, tt.getResult.err
			})
			kfClient.PrependReactor(deleteMethod, inferenceServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.deleteResult.isvc, tt.deleteResult.err
			})

			v1Client := fake.NewSimpleClientset().CoreV1()

			policyV1Client := fake.NewSimpleClientset().PolicyV1().(*fakepolicyv1.FakePolicyV1)
			policyV1Client.Fake.PrependReactor(deleteMethod, pdbResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, pdb, nil
			})

			istioClient := fakeistio.NewSimpleClientset().NetworkingV1beta1().(*fakeistionetworking.FakeNetworkingV1beta1)
			istioClient.PrependReactor(deleteMethod, virtualServiceResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, vs, nil
			})

			containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)

			templater := clusterresource.NewInferenceServiceTemplater(tt.deployConfig)

			ctl, _ := newController(knClient, kfClient, v1Client, nil, policyV1Client, istioClient, tt.deployConfig, containerFetcher, templater)
			mSvc, err := ctl.Delete(context.Background(), tt.modelService)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, mSvc)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, mSvc)
		})
	}
}
