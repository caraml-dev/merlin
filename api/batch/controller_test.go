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
	"strconv"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	sparkOpFake "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned/fake"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fake2 "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	jobspec "github.com/gojek/merlin-pyspark-app/pkg/spec"

	batchMock "github.com/gojek/merlin/batch/mocks"
	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/mlp"
	mlpMock "github.com/gojek/merlin/mlp/mocks"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/storage/mocks"
)

const (
	driverServiceAccountName = "driver-service-account"
	configName               = "prediction-job-config"
)

var (
	projectID, _ = models.ParseID("1")
	secret       = mlp.Secret{
		Id:   1,
		Name: "my-secret",
		Data: "secret-data",
	}

	predictionJob = &models.PredictionJob{
		Name:           jobName,
		ID:             jobID,
		VersionModelID: modelID,
		VersionID:      versionID,
		ProjectID:      projectID,
		Config: &models.Config{
			ServiceAccountName: secret.Name,
			JobConfig: &jobspec.PredictionJob{
				Version: "v1",
				Kind:    "PredictionJob",
				Name:    jobName,
				Source: &jobspec.PredictionJob_BigquerySource{
					BigquerySource: &jobspec.BigQuerySource{
						Table: "project.dataset.table_iris",
						Features: []string{
							"sepal_length",
							"sepal_width",
							"petal_length",
							"petal_width",
						},
					},
				},
				Model: &jobspec.Model{
					Type: jobspec.ModelType_PYFUNC_V2,
					Uri:  "gs://bucket-name/e2e/artifacts/model",
					Result: &jobspec.Model_ModelResult{
						Type: jobspec.ResultType_INTEGER,
					},
				},
				Sink: &jobspec.PredictionJob_BigquerySink{
					BigquerySink: &jobspec.BigQuerySink{
						Table:         "project.dataset.table_iris_result",
						StagingBucket: "bucket-name",
						ResultColumn:  "prediction",
						SaveMode:      jobspec.SaveMode_OVERWRITE,
						Options: map[string]string{
							"project": "project",
						},
					}},
			},
			ImageRef: imageRef,
			ResourceRequest: &models.PredictionJobResourceRequest{
				DriverCPURequest:      strconv.Itoa(int(driverCore)),
				DriverMemoryRequest:   driverMemory,
				ExecutorReplica:       executorReplica,
				ExecutorCPURequest:    strconv.Itoa(int(executorCore)),
				ExecutorMemoryRequest: executorMemory,
			},
		},
	}

	sparkApp, _ = CreateSparkApplicationResource(predictionJob)
	namespace   = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultNamespace}}
	errRaised           = errors.New("error")
)

func TestSubmit(t *testing.T) {
	type namespaceCreationResult struct {
		namespace *corev1.Namespace
		error     error
	}

	type driverAuthzCreationResult struct {
		serviceAccountName string
		error              error
	}

	type secretCreationResult struct {
		secretName string
		error      error
	}

	type jobConfigCreationResult struct {
		configName string
		error      error
	}

	type sparkResourceCreationResult struct {
		resource *v1beta2.SparkApplication
		error    error
	}

	type sparkResourceSubmissionResult struct {
		resource *v1beta2.SparkApplication
		error    error
	}

	type saveResult struct {
		error error
	}

	tests := []struct {
		name                          string
		namespace                     string
		existingServiceAccount        mlp.Secret
		namespaceCreationResult       namespaceCreationResult
		driverAuthzCreationResult     driverAuthzCreationResult
		jobConfigCreationResult       jobConfigCreationResult
		secretCreationResult          secretCreationResult
		sparkResourceCreationResult   sparkResourceCreationResult
		sparkResourceSubmissionResult sparkResourceSubmissionResult
		saveResult                    saveResult
		wantError                     bool
		wantErrorMsg                  string
	}{
		{
			name:                   "Nominal case",
			namespace:              defaultNamespace,
			existingServiceAccount: secret,
			namespaceCreationResult: namespaceCreationResult{
				namespace: namespace,
				error:     nil,
			},
			driverAuthzCreationResult: driverAuthzCreationResult{
				serviceAccountName: driverServiceAccountName,
				error:              nil,
			},
			secretCreationResult: secretCreationResult{
				secretName: secret.Name,
				error:      nil,
			},
			jobConfigCreationResult: jobConfigCreationResult{
				configName: configName,
				error:      nil,
			},
			sparkResourceCreationResult: sparkResourceCreationResult{
				resource: sparkApp,
				error:    nil,
			},
			sparkResourceSubmissionResult: sparkResourceSubmissionResult{
				resource: sparkApp,
				error:    nil,
			},
			saveResult: saveResult{
				error: nil,
			},
		},
		{
			name:                   "Failed namespace creation",
			namespace:              defaultNamespace,
			existingServiceAccount: secret,
			namespaceCreationResult: namespaceCreationResult{
				namespace: nil,
				error:     errRaised,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed creating namespace %s: %v", defaultNamespace, errRaised),
		},
		{
			name:                   "Failed spark driver authorization",
			namespace:              defaultNamespace,
			existingServiceAccount: secret,
			namespaceCreationResult: namespaceCreationResult{
				namespace: namespace,
				error:     nil,
			},
			driverAuthzCreationResult: driverAuthzCreationResult{
				serviceAccountName: "",
				error:              errRaised,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed creating spark driver authorization in namespace %s: %v", defaultNamespace, errRaised),
		},
		{
			name:      "Service account is not found",
			namespace: defaultNamespace,
			// existingServiceAccount: nil,
			namespaceCreationResult: namespaceCreationResult{
				namespace: namespace,
				error:     nil,
			},
			driverAuthzCreationResult: driverAuthzCreationResult{
				serviceAccountName: driverServiceAccountName,
				error:              nil,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("service account %s is not found within %s project: not found", predictionJob.Config.ServiceAccountName, defaultNamespace),
		},
		{
			name:                   "Failed secret creation",
			namespace:              defaultNamespace,
			existingServiceAccount: secret,
			namespaceCreationResult: namespaceCreationResult{
				namespace: namespace,
				error:     nil,
			},
			driverAuthzCreationResult: driverAuthzCreationResult{
				serviceAccountName: driverServiceAccountName,
				error:              nil,
			},
			secretCreationResult: secretCreationResult{
				secretName: "",
				error:      errRaised,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed creating secret for job %s in namespace %s: %v", predictionJob.Name, defaultNamespace, errRaised),
		},
		{
			name:                   "Failed job spec creation",
			namespace:              defaultNamespace,
			existingServiceAccount: secret,
			namespaceCreationResult: namespaceCreationResult{
				namespace: namespace,
				error:     nil,
			},
			driverAuthzCreationResult: driverAuthzCreationResult{
				serviceAccountName: driverServiceAccountName,
				error:              nil,
			},
			secretCreationResult: secretCreationResult{
				secretName: secret.Name,
				error:      nil,
			},
			jobConfigCreationResult: jobConfigCreationResult{
				configName: "",
				error:      errRaised,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed creating job specification configmap for job %s in namespace %s: %v", predictionJob.Name, defaultNamespace, errRaised),
		},
		{
			name:                   "Failed submitting spark application",
			namespace:              defaultNamespace,
			existingServiceAccount: secret,
			namespaceCreationResult: namespaceCreationResult{
				namespace: namespace,
				error:     nil,
			},
			driverAuthzCreationResult: driverAuthzCreationResult{
				serviceAccountName: driverServiceAccountName,
				error:              nil,
			},
			secretCreationResult: secretCreationResult{
				secretName: secret.Name,
				error:      nil,
			},
			jobConfigCreationResult: jobConfigCreationResult{
				configName: configName,
				error:      nil,
			},
			sparkResourceCreationResult: sparkResourceCreationResult{
				resource: sparkApp,
				error:    nil,
			},
			sparkResourceSubmissionResult: sparkResourceSubmissionResult{
				resource: nil,
				error:    errRaised,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed submitting spark application to spark controller for job %s in namespace %s: %v", predictionJob.Name, defaultNamespace, errRaised),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := &mocks.PredictionJobStorage{}
			mockMlpAPIClient := &mlpMock.APIClient{}
			mockSparkClient := &sparkOpFake.Clientset{}
			mockKubeClient := &fake2.Clientset{}
			mockManifestManager := &batchMock.ManifestManager{}
			clusterMetadata := cluster.Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}
			ctl := NewController(mockStorage, mockMlpAPIClient, mockSparkClient, mockKubeClient, mockManifestManager, clusterMetadata)

			mockKubeClient.PrependReactor("get", "namespaces", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, kerrors.NewNotFound(schema.GroupResource{}, action.(ktesting.GetAction).GetName())
			})
			mockKubeClient.PrependReactor("create", "namespaces", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, test.namespaceCreationResult.namespace, test.namespaceCreationResult.error
			})
			mockSparkClient.PrependReactor("create", "sparkapplications", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if test.sparkResourceSubmissionResult.error != nil {
					fmt.Println("test")
				}
				return true, test.sparkResourceSubmissionResult.resource, test.sparkResourceSubmissionResult.error
			})
			mockManifestManager.On("CreateDriverAuthorization", test.namespace).Return(test.driverAuthzCreationResult.serviceAccountName, test.driverAuthzCreationResult.error)
			mockManifestManager.On("CreateSecret", jobName, test.namespace, secret.Data).Return(test.secretCreationResult.secretName, test.secretCreationResult.error)
			mockManifestManager.On("CreateJobSpec", jobName, test.namespace, predictionJob.Config.JobConfig).Return(test.jobConfigCreationResult.configName, test.jobConfigCreationResult.error)
			if test.existingServiceAccount.Id != int32(0) {
				mockMlpAPIClient.On("GetPlainSecretByNameAndProjectID", context.Background(), secret.Name, int32(1)).Return(test.existingServiceAccount, nil)
			} else {
				mockMlpAPIClient.On("GetPlainSecretByNameAndProjectID", context.Background(), secret.Name, int32(1)).Return(mlp.Secret{}, errors.New("not found"))
			}
			mockStorage.On("Save", predictionJob).Return(test.saveResult.error)

			if test.wantError {
				mockManifestManager.On("DeleteSecret", jobName, defaultNamespace).Return(nil)
				mockManifestManager.On("DeleteJobSpec", jobName, defaultNamespace).Return(nil)
			}

			err := ctl.Submit(predictionJob, test.namespace)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}
			assert.NoError(t, err)

			// validate spark application submitted to spark client
			expectedSparkApp, _ := CreateSparkApplicationResource(predictionJob)
			expectedSparkApp.Spec.Driver.ServiceAccount = &test.driverAuthzCreationResult.serviceAccountName
			actions := mockSparkClient.Fake.Actions()
			createAction := actions[0].(ktesting.CreateAction)
			assert.Len(t, actions, 1)
			assert.Equal(t, expectedSparkApp, createAction.GetObject())

			// validate all expectations are met
			mockManifestManager.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestCleanupAfterSubmitFailed(t *testing.T) {
	mockStorage := &mocks.PredictionJobStorage{}
	mockStorage.On("Save", predictionJob).Return(nil)

	mockMlpAPIClient := &mlpMock.APIClient{}
	mockMlpAPIClient.On("GetPlainSecretByNameAndProjectID", context.Background(), secret.Name, int32(1)).Return(secret, nil)

	mockSparkClient := &sparkOpFake.Clientset{}
	mockSparkClient.PrependReactor("create", "sparkapplications", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("failed creating spark application")
	})
	mockKubeClient := &fake2.Clientset{}
	mockManifestManager := &batchMock.ManifestManager{}
	clusterMetadata := cluster.Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}
	ctl := NewController(mockStorage, mockMlpAPIClient, mockSparkClient, mockKubeClient, mockManifestManager, clusterMetadata)

	mockManifestManager.On("DeleteSecret", jobName, defaultNamespace).Return(nil)
	mockManifestManager.On("DeleteJobSpec", jobName, defaultNamespace).Return(nil)

	err := ctl.Submit(predictionJob, defaultNamespace)
	assert.Error(t, err)
	mockManifestManager.AssertExpectations(t)
}

func TestOnUpdate(t *testing.T) {
	sparkApp.Namespace = defaultNamespace

	mockStorage := &mocks.PredictionJobStorage{}
	mockStorage.On("Save", predictionJob).Return(nil)
	mockStorage.On("Get", predictionJob.ID).Return(predictionJob, nil)

	mockMlpAPIClient := &mlpMock.APIClient{}
	mockMlpAPIClient.On("GetPlainSecretByNameAndProjectID", context.Background(), secret.Name, int32(1)).Return(secret, nil)

	mockSparkClient := &sparkOpFake.Clientset{}
	mockKubeClient := &fake2.Clientset{}
	mockManifestManager := &batchMock.ManifestManager{}
	clusterMetadata := cluster.Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}
	ctl := NewController(mockStorage, mockMlpAPIClient, mockSparkClient, mockKubeClient, mockManifestManager, clusterMetadata).(*controller)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go ctl.Run(stopCh)
	time.Sleep(5 * time.Millisecond)

	serviceAccountName := "driver-service-account"
	mockManifestManager.On("CreateDriverAuthorization", defaultNamespace).Return(serviceAccountName, nil)
	mockManifestManager.On("CreateSecret", jobName, defaultNamespace, secret.Data).Return(jobName, nil)
	mockManifestManager.On("DeleteSecret", jobName, defaultNamespace).Return(nil)
	mockManifestManager.On("DeleteJobSpec", jobName, defaultNamespace).Return(nil)

	sparkAppNew := sparkApp.DeepCopy()

	// test unchanged state event is not being queued
	ctl.onUpdate(sparkApp, sparkAppNew)
	assert.Equal(t, 0, ctl.queue.Len())

	// test state change event is being queued
	sparkApp.Status.AppState.State = v1beta2.RunningState
	sparkAppNew.Status.AppState.State = v1beta2.CompletedState

	ctl.onUpdate(sparkApp, sparkAppNew)
	item, _ := ctl.queue.Get()
	key, ok := item.(string)
	expectedKey, _ := cache.MetaNamespaceKeyFunc(sparkApp)
	assert.True(t, ok)
	assert.Equal(t, expectedKey, key)
	ctl.queue.Forget(item)
	ctl.queue.Done(item)

	// test state change event update the prediction job status
	_ = ctl.informer.GetIndexer().Add(sparkAppNew)
	ctl.onUpdate(sparkApp, sparkAppNew)

	// wait until the state sync happen in the background
	time.Sleep(100 * time.Millisecond)
	mockStorage.AssertExpectations(t)
	call := mockStorage.Calls[1]
	assert.Equal(t, models.JobCompleted, call.Arguments[0].(*models.PredictionJob).Status)
}

func TestUpdateStatus(t *testing.T) {
	sparkApp.Namespace = defaultNamespace
	sparkApp.Status.AppState.State = v1beta2.NewState

	tests := []struct {
		name       string
		sparkState v1beta2.ApplicationStateType
		wantState  models.State
	}{
		{sparkState: v1beta2.SubmittedState, wantState: models.JobPending},
		{sparkState: v1beta2.RunningState, wantState: models.JobRunning},
		{sparkState: v1beta2.CompletedState, wantState: models.JobCompleted},
		{sparkState: v1beta2.FailedState, wantState: models.JobFailed},
		{sparkState: v1beta2.FailedSubmissionState, wantState: models.JobFailedSubmission},
		{sparkState: v1beta2.PendingRerunState, wantState: models.JobRunning},
		{sparkState: v1beta2.InvalidatingState, wantState: models.JobRunning},
		{sparkState: v1beta2.SucceedingState, wantState: models.JobRunning},
		{sparkState: v1beta2.FailingState, wantState: models.JobRunning},
		{sparkState: v1beta2.UnknownState, wantState: models.JobRunning},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := &mocks.PredictionJobStorage{}
			mockStorage.On("Save", predictionJob).Return(nil)
			mockStorage.On("Get", predictionJob.ID).Return(predictionJob, nil)

			mockMlpAPIClient := &mlpMock.APIClient{}
			mockMlpAPIClient.On("GetPlainSecretByNameAndProjectID", context.Background(), secret.Name, int32(1)).Return(secret, nil)

			mockSparkClient := &sparkOpFake.Clientset{}
			mockKubeClient := &fake2.Clientset{}
			mockManifestManager := &batchMock.ManifestManager{}
			clusterMetadata := cluster.Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}
			ctl := NewController(mockStorage, mockMlpAPIClient, mockSparkClient, mockKubeClient, mockManifestManager, clusterMetadata).(*controller)
			stopCh := make(chan struct{})
			defer close(stopCh)
			go ctl.Run(stopCh)
			time.Sleep(5 * time.Millisecond)

			if test.wantState.IsTerminal() {
				mockManifestManager.On("DeleteSecret", jobName, defaultNamespace).Return(nil)
				mockManifestManager.On("DeleteJobSpec", jobName, defaultNamespace).Return(nil)
			}

			sparkAppNew := sparkApp.DeepCopy()
			sparkAppNew.Status.AppState.State = test.sparkState

			_ = ctl.informer.GetIndexer().Add(sparkAppNew)
			ctl.onUpdate(sparkApp, sparkAppNew)

			// wait until the state sync happen in the background
			time.Sleep(500 * time.Millisecond)
			mockStorage.AssertExpectations(t)
			call := mockStorage.Calls[1]
			assert.Equal(t, test.wantState, call.Arguments[0].(*models.PredictionJob).Status)
			mockManifestManager.AssertExpectations(t)
		})
	}
}

func TestStop(t *testing.T) {
	type sparkResourceListResult struct {
		resource *v1beta2.SparkApplicationList
		error    error
	}

	type jobConfigCreationResult struct {
		configName string
		error      error
	}

	serviceAccountName := "driver-service-account"
	jobConfig := jobConfigCreationResult{
		configName: configName,
		error:      nil,
	}

	tests := []struct {
		name                    string
		namespace               string
		sparkResourceListResult sparkResourceListResult
		wantError               bool
		wantErrorMsg            string
	}{
		{
			name:      "Nominal case",
			namespace: defaultNamespace,
			sparkResourceListResult: sparkResourceListResult{
				resource: &v1beta2.SparkApplicationList{
					Items: []v1beta2.SparkApplication{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   jobName,
								Labels: defaultLabels,
							},
							Spec: sparkApp.Spec,
						},
					},
				},
				error: nil,
			},
		},
		{
			name:      "Prediction job does not exist",
			namespace: defaultNamespace,
			sparkResourceListResult: sparkResourceListResult{
				resource: &v1beta2.SparkApplicationList{},
				error:    errRaised,
			},
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("unable to retrieve spark application of prediction job id %s from spark client", models.ID(1).String()),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := &mocks.PredictionJobStorage{}
			mockMlpAPIClient := &mlpMock.APIClient{}
			mockSparkClient := &sparkOpFake.Clientset{}
			mockKubeClient := &fake2.Clientset{}
			mockManifestManager := &batchMock.ManifestManager{}
			clusterMetadata := cluster.Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}
			ctl := NewController(mockStorage, mockMlpAPIClient, mockSparkClient, mockKubeClient, mockManifestManager, clusterMetadata)

			mockKubeClient.PrependReactor("get", "namespaces", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, kerrors.NewNotFound(schema.GroupResource{}, action.(ktesting.GetAction).GetName())
			})
			mockManifestManager.On("CreateDriverAuthorization", defaultNamespace).Return(serviceAccountName, nil)
			mockManifestManager.On("CreateSecret", jobName, defaultNamespace, secret.Data).Return(jobName, nil)
			mockManifestManager.On("DeleteSecret", jobName, defaultNamespace).Return(nil)
			mockManifestManager.On("CreateJobSpec", jobName, test.namespace, predictionJob.Config.JobConfig).Return(jobConfig.configName, jobConfig.error)
			mockManifestManager.On("DeleteJobSpec", jobName, defaultNamespace).Return(nil)
			mockSparkClient.PrependReactor("create", "sparkapplications", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, sparkApp, nil
			})
			mockSparkClient.PrependReactor("list", "sparkapplications", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, test.sparkResourceListResult.resource, test.sparkResourceListResult.error
			})
			mockStorage.On("Save", predictionJob).Return(nil)

			err := ctl.Stop(predictionJob, namespace.Name)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, models.State("terminated"), predictionJob.Status)
		})
	}
}
