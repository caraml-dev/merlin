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
	"time"

	clusterMock "github.com/gojek/merlin/cluster/mocks"
	imageBuilderMock "github.com/gojek/merlin/imagebuilder/mocks"
	"github.com/gojek/merlin/mlp"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/storage/mocks"
)

var isDefaultTrue = true

func TestDeployEndpoint(t *testing.T) {
	type args struct {
		environment *models.Environment
		model       *models.Model
		version     *models.Version
		endpoint    *models.VersionEndpoint
	}

	env := &models.Environment{Name: "env1",
		Cluster:    "cluster1",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
		DefaultResourceRequest: &models.ResourceRequest{
			MinReplica:    0,
			MaxReplica:    1,
			CpuRequest:    resource.MustParse("1"),
			MemoryRequest: resource.MustParse("1Gi"),
		},
	}
	project := mlp.Project{Name: "project"}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{Id: 1}

	iSvcName := fmt.Sprintf("%s-%d", model.Name, version.Id)
	svcName := fmt.Sprintf("%s-%d.project.svc.cluster.local", model.Name, version.Id)
	url := fmt.Sprintf("%s-%d.example.com", model.Name, version.Id)

	tests := []struct {
		name                string
		args                args
		wantDeployError     bool
		wantBuildImageError bool
	}{
		{
			"success: new endpoint default resource request",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{},
			},
			false,
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
						CpuRequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
			},
			false,
			false,
		},
		{
			"success: pytorch model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{Id: 1, Properties: models.KV{
					models.PropertyPyTorchClassName: "MyModel",
				}},
				&models.VersionEndpoint{},
			},
			false,
			false,
		},
		{
			"success: empty pytorch class name will fallback to PyTorchModel",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{Id: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
			},
			false,
			false,
		},
		{
			"success: empty pytorch class name will fallback to PyTorchModel",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
				&models.Version{Id: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
			},
			false,
			false,
		},
		{
			"failed: error building pyfunc image",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
				&models.Version{Id: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
			},
			true,
			true,
		},
		{
			"failed: error deploying",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{},
			},
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envController := &clusterMock.Controller{}
			if tt.wantDeployError {
				envController.On("Deploy", mock.Anything, mock.Anything).
					Return(nil, errors.New("error deploying"))
			} else {
				envController.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						Url:         url,
					}, nil)
			}

			imgBuilder := &imageBuilderMock.ImageBuilder{}
			if tt.wantBuildImageError {
				imgBuilder.On("BuildImage", tt.args.model.Project, tt.args.model, tt.args.version).
					Return("", errors.New("error building image"))
			} else {
				imgBuilder.On("BuildImage", tt.args.model.Project, tt.args.model, tt.args.version).
					Return(fmt.Sprintf("gojek/mymodel-1:latest"), nil)
			}
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
			endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, mockCfg.Environment, mockCfg.FeatureToggleConfig.MonitoringConfig)
			e, err := endpointSvc.DeployEndpoint(tt.args.environment, tt.args.model, tt.args.version, tt.args.endpoint)

			// delay to make second save happen before checking
			time.Sleep(20 * time.Millisecond)

			assert.NoError(t, err)
			assert.Equal(t, "", e.Url)
			assert.Equal(t, models.EndpointPending, e.Status)
			assert.Equal(t, project.Name, e.Namespace)
			assert.Equal(t, iSvcName, e.InferenceServiceName)

			if tt.args.endpoint.ResourceRequest != nil {
				assert.Equal(t, e.ResourceRequest, tt.args.endpoint.ResourceRequest)
			} else {
				assert.Equal(t, e.ResourceRequest, tt.args.environment.DefaultResourceRequest)
			}

			mockStorage.AssertNumberOfCalls(t, "Save", 2)
			savedEndpoint := mockStorage.Calls[1].Arguments[0].(*models.VersionEndpoint)
			assert.Equal(t, tt.args.model.Id, savedEndpoint.VersionModelId)
			assert.Equal(t, tt.args.version.Id, savedEndpoint.VersionId)
			assert.Equal(t, tt.args.model.Project.Name, savedEndpoint.Namespace)
			assert.Equal(t, tt.args.environment.Name, savedEndpoint.EnvironmentName)

			if tt.args.endpoint.ResourceRequest != nil {
				assert.Equal(t, tt.args.endpoint.ResourceRequest, savedEndpoint.ResourceRequest)
			} else {
				assert.Equal(t, tt.args.environment.DefaultResourceRequest, savedEndpoint.ResourceRequest)
			}
			if tt.wantDeployError || tt.wantBuildImageError {
				assert.Equal(t, models.EndpointFailed, savedEndpoint.Status)
			} else {
				assert.Equal(t, models.EndpointRunning, savedEndpoint.Status)
				assert.Equal(t, url, savedEndpoint.Url)
				assert.Equal(t, iSvcName, savedEndpoint.InferenceServiceName)
			}
		})
	}
}

func TestListContainers(t *testing.T) {
	project := mlp.Project{Id: 1, Name: "my-project"}
	model := &models.Model{Id: 1, Name: "model", Type: models.ModelTypeXgboost, Project: project, ProjectId: models.Id(project.Id)}
	version := &models.Version{Id: 1}
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
					Id:              id,
					VersionId:       version.Id,
					VersionModelId:  model.Id,
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
					Id:              id,
					VersionId:       version.Id,
					VersionModelId:  model.Id,
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

		endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, cfg.Environment, cfg.FeatureToggleConfig.MonitoringConfig)

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
