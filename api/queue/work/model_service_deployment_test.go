package work

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/caraml-dev/merlin/cluster"
	clusterMock "github.com/caraml-dev/merlin/cluster/mocks"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	imageBuilderMock "github.com/caraml-dev/merlin/pkg/imagebuilder/mocks"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestExecuteDeployment(t *testing.T) {
	isDefaultTrue := true
	loggerDestinationURL := "http://logger.default"

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
			GPURequest:    resource.MustParse("0"),
		},
	}

	mlpLabels := mlp.Labels{
		{Key: "key-1", Value: "value-1"},
	}

	versionLabels := models.KV{
		"key-1": "value-11",
		"key-2": "value-2",
	}

	svcMetadata := models.Metadata{
		Labels: mlp.Labels{
			{Key: "key-1", Value: "value-11"},
			{Key: "key-2", Value: "value-2"},
		},
	}

	project := mlp.Project{Name: "project", Labels: mlpLabels}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1, Labels: versionLabels}
	iSvcName := fmt.Sprintf("%s-%d-1", model.Name, version.ID)
	svcName := fmt.Sprintf("%s-%d-1.project.svc.cluster.local", model.Name, version.ID)
	url := fmt.Sprintf("%s-%d-1.example.com", model.Name, version.ID)

	tests := []struct {
		name              string
		endpoint          *models.VersionEndpoint
		model             *models.Model
		version           *models.Version
		deployErr         error
		deploymentStorage func() *mocks.DeploymentStorage
		storage           func() *mocks.VersionEndpointStorage
		controller        func() *clusterMock.Controller
		imageBuilder      func() *imageBuilderMock.ImageBuilder
	}{
		{
			name:    "Success: Default",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:    "Success: Pytorch Model",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:    "Success: empty pyfunc model",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", context.Background(), mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", context.Background(), project, mock.Anything, mock.Anything).
					Return("gojek/mymodel-1:latest", nil)
				return mockImgBuilder
			},
		},
		{
			name:    "Success: pytorch model with transformer",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", context.Background(), mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", context.Background(), project, mock.Anything, mock.Anything).
					Return("gojek/mymodel-1:latest", nil)
				return mockImgBuilder
			},
		},
		{
			name:    "Success: Default With GPU",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    0,
					MaxReplica:    1,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GPUName:       "NVIDIA P4",
					GPURequest:    resource.MustParse("1"),
				},
				VersionID: version.ID,
				Namespace: project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:      "Failed: deployment failed",
			model:     model,
			version:   version,
			deployErr: errors.New("Failed to deploy"),
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(nil, errors.New("Failed to deploy"))
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:      "Failed: image builder failed",
			model:     &models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
			version:   version,
			deployErr: errors.New("Failed to build image"),
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("Failed to build image"))
				return mockImgBuilder
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := tt.controller()
			controllers := map[string]cluster.Controller{env.Name: ctrl}
			imgBuilder := tt.imageBuilder()
			mockStorage := tt.storage()
			mockDeploymentStorage := tt.deploymentStorage()
			job := &queue.Job{
				Name: "job",
				Arguments: queue.Arguments{
					dataArgKey: EndpointJob{
						Endpoint: tt.endpoint,
						Version:  tt.version,
						Model:    tt.model,
						Project:  tt.model.Project,
					},
				},
			}
			svc := &ModelServiceDeployment{
				ClusterControllers:   controllers,
				ImageBuilder:         imgBuilder,
				Storage:              mockStorage,
				DeploymentStorage:    mockDeploymentStorage,
				LoggerDestinationURL: loggerDestinationURL,
			}

			err := svc.Deploy(job)
			assert.Equal(t, tt.deployErr, err)

			if len(ctrl.ExpectedCalls) > 0 && ctrl.ExpectedCalls[0].ReturnArguments[0] != nil {
				deployedSvc := ctrl.ExpectedCalls[0].ReturnArguments[0].(*models.Service)
				assert.Equal(t, svcMetadata, deployedSvc.Metadata)
				assert.Equal(t, iSvcName, deployedSvc.Name)
			}

			mockStorage.AssertNumberOfCalls(t, "Save", 1)
			mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 2)

			savedEndpoint := mockStorage.Calls[0].Arguments[0].(*models.VersionEndpoint)
			assert.Equal(t, tt.model.ID, savedEndpoint.VersionModelID)
			assert.Equal(t, tt.version.ID, savedEndpoint.VersionID)
			assert.Equal(t, tt.model.Project.Name, savedEndpoint.Namespace)
			assert.Equal(t, env.Name, savedEndpoint.EnvironmentName)

			if tt.endpoint.ResourceRequest != nil {
				assert.Equal(t, tt.endpoint.ResourceRequest, savedEndpoint.ResourceRequest)
			} else {
				assert.Equal(t, env.DefaultResourceRequest, savedEndpoint.ResourceRequest)
			}

			if tt.deployErr != nil {
				assert.Equal(t, models.EndpointFailed, savedEndpoint.Status)
			} else {
				assert.Equal(t, models.EndpointRunning, savedEndpoint.Status)
				assert.Equal(t, url, savedEndpoint.URL)
				assert.Equal(t, "", savedEndpoint.InferenceServiceName)
			}
		})
	}
}

func TestExecuteRedeployment(t *testing.T) {
	isDefaultTrue := true
	loggerDestinationURL := "http://logger.default"

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
			GPURequest:    resource.MustParse("0"),
		},
	}

	mlpLabels := mlp.Labels{
		{Key: "key-1", Value: "value-1"},
	}

	versionLabels := models.KV{
		"key-1": "value-11",
		"key-2": "value-2",
	}

	svcMetadata := models.Metadata{
		Labels: mlp.Labels{
			{Key: "key-1", Value: "value-11"},
			{Key: "key-2", Value: "value-2"},
		},
	}

	project := mlp.Project{Name: "project", Labels: mlpLabels}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1, Labels: versionLabels}

	modelSvcName := fmt.Sprintf("%s-%d-2", model.Name, version.ID)
	svcName := fmt.Sprintf("%s-%d-2.project.svc.cluster.local", model.Name, version.ID)
	url := fmt.Sprintf("%s-%d-2.example.com", model.Name, version.ID)

	tests := []struct {
		name                   string
		endpoint               *models.VersionEndpoint
		model                  *models.Model
		version                *models.Version
		expectedEndpointStatus models.EndpointStatus
		deployErr              error
		deploymentStorage      func() *mocks.DeploymentStorage
		storage                func() *mocks.VersionEndpointStorage
		controller             func() *clusterMock.Controller
		imageBuilder           func() *imageBuilderMock.ImageBuilder
	}{
		{
			name:    "Success: Redeploy running endpoint",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointRunning,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointRunning,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(2),
					InferenceServiceName: modelSvcName,
					Status:               models.EndpointRunning,
					URL:                  url,
					ServiceName:          svcName,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:            fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						CurrentIsvcName: fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						RevisionID:      models.ID(2),
						Namespace:       project.Name,
						ServiceName:     fmt.Sprintf("%s-%d-2.project.svc.cluster.local", model.Name, version.ID),
						URL:             fmt.Sprintf("%s-%d-2.example.com", model.Name, version.ID),
						Metadata:        svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:    "Success: Redeploy serving endpoint",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointServing,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointServing,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(2),
					InferenceServiceName: modelSvcName,
					Status:               models.EndpointServing,
					URL:                  url,
					ServiceName:          svcName,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:            fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						CurrentIsvcName: fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						RevisionID:      models.ID(2),
						Namespace:       project.Name,
						ServiceName:     fmt.Sprintf("%s-%d-2.project.svc.cluster.local", model.Name, version.ID),
						URL:             fmt.Sprintf("%s-%d-2.example.com", model.Name, version.ID),
						Metadata:        svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:    "Success: Redeploy failed endpoint",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointFailed,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointRunning,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(2),
					InferenceServiceName: modelSvcName,
					Status:               models.EndpointRunning,
					URL:                  url,
					ServiceName:          svcName,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:            fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						CurrentIsvcName: fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						RevisionID:      models.ID(2),
						Namespace:       project.Name,
						ServiceName:     fmt.Sprintf("%s-%d-2.project.svc.cluster.local", model.Name, version.ID),
						URL:             fmt.Sprintf("%s-%d-2.example.com", model.Name, version.ID),
						Metadata:        svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := tt.controller()
			controllers := map[string]cluster.Controller{env.Name: ctrl}
			imgBuilder := tt.imageBuilder()
			mockStorage := tt.storage()
			mockDeploymentStorage := tt.deploymentStorage()
			job := &queue.Job{
				Name: "job",
				Arguments: queue.Arguments{
					dataArgKey: EndpointJob{
						Endpoint: tt.endpoint,
						Version:  tt.version,
						Model:    tt.model,
						Project:  tt.model.Project,
					},
				},
			}
			svc := &ModelServiceDeployment{
				ClusterControllers:   controllers,
				ImageBuilder:         imgBuilder,
				Storage:              mockStorage,
				DeploymentStorage:    mockDeploymentStorage,
				LoggerDestinationURL: loggerDestinationURL,
			}

			err := svc.Deploy(job)
			assert.Equal(t, tt.deployErr, err)

			if len(ctrl.ExpectedCalls) > 0 && ctrl.ExpectedCalls[0].ReturnArguments[0] != nil {
				deployedSvc := ctrl.ExpectedCalls[0].ReturnArguments[0].(*models.Service)
				assert.Equal(t, svcMetadata, deployedSvc.Metadata)
				assert.Equal(t, modelSvcName, deployedSvc.Name)
			}

			mockStorage.AssertNumberOfCalls(t, "Save", 1)
			mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 2)

			savedEndpoint := mockStorage.Calls[0].Arguments[0].(*models.VersionEndpoint)
			assert.Equal(t, tt.model.ID, savedEndpoint.VersionModelID)
			assert.Equal(t, tt.version.ID, savedEndpoint.VersionID)
			assert.Equal(t, tt.model.Project.Name, savedEndpoint.Namespace)
			assert.Equal(t, env.Name, savedEndpoint.EnvironmentName)

			if tt.endpoint.ResourceRequest != nil {
				assert.Equal(t, tt.endpoint.ResourceRequest, savedEndpoint.ResourceRequest)
			} else {
				assert.Equal(t, env.DefaultResourceRequest, savedEndpoint.ResourceRequest)
			}

			if tt.deployErr != nil {
				assert.Equal(t, models.EndpointFailed, savedEndpoint.Status)
			} else {
				assert.Equal(t, tt.expectedEndpointStatus, savedEndpoint.Status)
				assert.Equal(t, url, savedEndpoint.URL)
				assert.Equal(t, modelSvcName, savedEndpoint.InferenceServiceName)
			}
		})
	}
}
