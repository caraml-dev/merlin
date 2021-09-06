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

// +build unit

package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gojek/merlin/cluster"
	clusterMock "github.com/gojek/merlin/cluster/mocks"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	imageBuilderMock "github.com/gojek/merlin/pkg/imagebuilder/mocks"
	queueMock "github.com/gojek/merlin/queue/mocks"
	"github.com/gojek/merlin/storage/mocks"
)

var (
	isDefaultTrue        = true
	loggerDestinationURL = "http://logger.default"
)

func TestDeployEndpoint(t *testing.T) {
	type args struct {
		environment      *models.Environment
		model            *models.Model
		version          *models.Version
		endpoint         *models.VersionEndpoint
		expectedEndpoint *models.VersionEndpoint
	}

	env := &models.Environment{
		Name:       "env1",
		Cluster:    "cluster1",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
		DefaultResourceRequest: &models.ResourceRequest{
			MinReplica:    0,
			MaxReplica:    1,
			CPURequest:    resource.MustParse("1"),
			MemoryRequest: resource.MustParse("1Gi"),
		},
	}
	project := mlp.Project{Name: "project"}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1}

	iSvcName := fmt.Sprintf("%s-%d", model.Name, version.ID)

	tests := []struct {
		name            string
		args            args
		wantDeployError bool
	}{
		{
			"success: new endpoint default resource request",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{},
				&models.VersionEndpoint{},
			},
			false,
		},
		{
			"success: new endpoint non default resource request",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
			},
			false,
		},
		{
			"success: pytorch model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{ID: 1, Properties: models.KV{
					models.PropertyPyTorchClassName: "MyModel",
				}},
				&models.VersionEndpoint{},
				&models.VersionEndpoint{},
			},
			false,
		},
		{
			"success: empty pytorch class name will fallback to PyTorchModel",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
			},
			false,
		},
		{
			"success: empty pyfunc model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
					EnvVars: models.EnvVars{
						{
							Name:  "MODEL_NAME",
							Value: "model-1",
						},
						{
							Name:  "MODEL_DIR",
							Value: "/model",
						},
						{
							Name:  "WORKERS",
							Value: "1",
						},
					},
				},
			},
			false,
		},
		{
			"success: empty custom model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypeCustom},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
					EnvVars: models.EnvVars{
						{
							Name:  "TF_MODEL_NAME",
							Value: "saved_model.pb",
						},
						{
							Name:  "NUM_OF_ITERATION",
							Value: "1",
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
					EnvVars: models.EnvVars{
						{
							Name:  "TF_MODEL_NAME",
							Value: "saved_model.pb",
						},
						{
							Name:  "NUM_OF_ITERATION",
							Value: "1",
						},
					},
				},
			},
			false,
		},
		{
			"failed: error deploying",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{},
				&models.VersionEndpoint{},
			},
			true,
		},
		{
			"success: pytorch model with transformer",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{ID: 1, Properties: models.KV{
					models.PropertyPyTorchClassName: "MyModel",
				}},
				&models.VersionEndpoint{
					Transformer: &models.Transformer{
						Enabled:         true,
						Image:           "ghcr.io/gojek/merlin-transformer-test",
						ResourceRequest: env.DefaultResourceRequest,
					},
				},
				&models.VersionEndpoint{
					Transformer: &models.Transformer{
						Enabled:         true,
						Image:           "ghcr.io/gojek/merlin-transformer-test",
						ResourceRequest: env.DefaultResourceRequest,
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - overwrite logger mode if invalid",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LoggerMode(""),
						},
						Transformer: &models.LoggerConfig{
							Enabled: false,
							Mode:    models.LoggerMode("randomString"),
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogAll,
						},
						Transformer: &models.LoggerConfig{
							Enabled: false,
							Mode:    models.LogAll,
						},
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - only model logger",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - only transformer logger",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - both model and transformer specified with valid value",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envController := &clusterMock.Controller{}
			mockQueueProducer := &queueMock.Producer{}
			if tt.wantDeployError {
				mockQueueProducer.On("EnqueueJob", mock.Anything).Return(errors.New("Failed to queue job"))
			} else {
				mockQueueProducer.On("EnqueueJob", mock.Anything).Return(nil)
			}

			imgBuilder := &imageBuilderMock.ImageBuilder{}
			mockStorage := &mocks.VersionEndpointStorage{}
			mockDeploymentStorage := &mocks.DeploymentStorage{}
			mockStorage.On("Save", mock.Anything).Return(nil)
			mockDeploymentStorage.On("Save", mock.Anything).Return(nil, nil)
			mockCfg := &config.Config{
				Environment: "dev",
				FeatureToggleConfig: config.FeatureToggleConfig{
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: false,
					},
				},
			}

			controllers := map[string]cluster.Controller{env.Name: envController}
			// endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, mockCfg.Environment, mockCfg.FeatureToggleConfig.MonitoringConfig, loggerDestinationURL)
			endpointSvc := NewEndpointService(EndpointServiceParams{
				ClusterControllers:   controllers,
				ImageBuilder:         imgBuilder,
				Storage:              mockStorage,
				DeploymentStorage:    mockDeploymentStorage,
				Environment:          mockCfg.Environment,
				MonitoringConfig:     mockCfg.FeatureToggleConfig.MonitoringConfig,
				LoggerDestinationURL: loggerDestinationURL,
				JobProducer:          mockQueueProducer,
			})
			errRaised, err := endpointSvc.DeployEndpoint(tt.args.environment, tt.args.model, tt.args.version, tt.args.endpoint)

			// delay to make second save happen before checking
			// time.Sleep(20 * time.Millisecond)

			if tt.wantDeployError {
				assert.True(t, err != nil)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "", errRaised.URL)
				assert.Equal(t, models.EndpointPending, errRaised.Status)
				assert.Equal(t, project.Name, errRaised.Namespace)
				assert.Equal(t, iSvcName, errRaised.InferenceServiceName)

				if tt.args.endpoint.ResourceRequest != nil {
					assert.Equal(t, errRaised.ResourceRequest, tt.args.expectedEndpoint.ResourceRequest)
				} else {
					assert.Equal(t, errRaised.ResourceRequest, tt.args.environment.DefaultResourceRequest)
				}

				assert.Equal(t, errRaised.EnvVars, tt.args.expectedEndpoint.EnvVars)
				assert.Equal(t, tt.args.expectedEndpoint.Logger, errRaised.Logger)
				mockStorage.AssertNumberOfCalls(t, "Save", 1)
			}

			if tt.args.endpoint.Transformer != nil {
				assert.Equal(t, tt.args.endpoint.Transformer.Enabled, errRaised.Transformer.Enabled)
			}
		})
	}
}

func TestListContainers(t *testing.T) {
	project := mlp.Project{Id: 1, Name: "my-project"}
	model := &models.Model{ID: 1, Name: "model", Type: models.ModelTypeXgboost, Project: project, ProjectID: models.ID(project.Id)}
	version := &models.Version{ID: 1}
	id := uuid.New()
	env := &models.Environment{Name: "my-env", Cluster: "my-cluster", IsDefault: &isDefaultTrue}
	cfg := &config.Config{
		Environment: "dev",
		FeatureToggleConfig: config.FeatureToggleConfig{
			MonitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: false,
			},
		},
	}

	type args struct {
		model   *models.Model
		version *models.Version
		id      uuid.UUID
	}

	type componentMock struct {
		versionEndpoint       *models.VersionEndpoint
		imageBuilderContainer *models.Container
		modelContainers       []*models.Container
	}

	tests := []struct {
		name      string
		args      args
		mock      componentMock
		wantError bool
	}{
		{
			"success: non-pyfunc model",
			args{
				model, version, id,
			},
			componentMock{
				&models.VersionEndpoint{
					ID:              id,
					VersionID:       version.ID,
					VersionModelID:  model.ID,
					EnvironmentName: env.Name,
				},
				nil,
				[]*models.Container{
					{
						Name:       "user-container",
						PodName:    "mymodel-2-predictor-default-hlqgv-deployment-6f478cbc67-mp7zf",
						Namespace:  project.Name,
						Cluster:    env.Cluster,
						GcpProject: env.GcpProject,
					},
				},
			},
			false,
		},
		{
			"success: pyfunc model",
			args{
				model, version, id,
			},
			componentMock{
				&models.VersionEndpoint{
					ID:              id,
					VersionID:       version.ID,
					VersionModelID:  model.ID,
					EnvironmentName: env.Name,
				},
				&models.Container{
					Name:       "kaniko-0",
					PodName:    "pod-1",
					Namespace:  "mlp",
					Cluster:    env.Cluster,
					GcpProject: env.GcpProject,
				},
				[]*models.Container{
					{
						Name:       "user-container",
						PodName:    "mymodel-2-predictor-default-hlqgv-deployment-6f478cbc67-mp7zf",
						Namespace:  project.Name,
						Cluster:    env.Cluster,
						GcpProject: env.GcpProject,
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		imgBuilder := &imageBuilderMock.ImageBuilder{}
		imgBuilder.On("GetContainers", mock.Anything, mock.Anything, mock.Anything).
			Return(tt.mock.imageBuilderContainer, nil)

		envController := &clusterMock.Controller{}
		envController.On("GetContainers", "my-project", "serving.kubeflow.org/inferenceservice=model-1").
			Return(tt.mock.modelContainers, nil)

		controllers := map[string]cluster.Controller{env.Name: envController}

		mockStorage := &mocks.VersionEndpointStorage{}
		mockDeploymentStorage := &mocks.DeploymentStorage{}
		mockStorage.On("Get", mock.Anything).Return(tt.mock.versionEndpoint, nil)
		mockDeploymentStorage.On("Save", mock.Anything).Return(nil, nil)

		// endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, cfg.Environment, cfg.FeatureToggleConfig.MonitoringConfig, loggerDestinationURL)
		endpointSvc := NewEndpointService(EndpointServiceParams{
			ClusterControllers:   controllers,
			ImageBuilder:         imgBuilder,
			Storage:              mockStorage,
			DeploymentStorage:    mockDeploymentStorage,
			Environment:          cfg.Environment,
			MonitoringConfig:     cfg.FeatureToggleConfig.MonitoringConfig,
			LoggerDestinationURL: loggerDestinationURL,
		})
		containers, err := endpointSvc.ListContainers(tt.args.model, tt.args.version, tt.args.id)
		if !tt.wantError {
			assert.Nil(t, err, "unwanted error %v", err)
		} else {
			assert.NotNil(t, err, "expected error")
		}

		assert.NotNil(t, containers)
		expContainer := len(tt.mock.modelContainers)
		if tt.args.model.Type == models.ModelTypePyFunc {
			expContainer += 1
		}
		assert.Equal(t, expContainer, len(containers))
	}
}
