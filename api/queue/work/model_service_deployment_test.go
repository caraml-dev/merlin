package work

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gojek/merlin/cluster"
	clusterMock "github.com/gojek/merlin/cluster/mocks"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	imageBuilderMock "github.com/gojek/merlin/pkg/imagebuilder/mocks"
	"github.com/gojek/merlin/queue"
	"github.com/gojek/merlin/storage/mocks"
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
		},
	}
	project := mlp.Project{Name: "project"}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1}
	iSvcName := fmt.Sprintf("%s-%d", model.Name, version.ID)
	svcName := fmt.Sprintf("%s-%d.project.svc.cluster.local", model.Name, version.ID)
	url := fmt.Sprintf("%s-%d.example.com", model.Name, version.ID)

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
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
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
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:  "Success: Pytorch Model",
			model: &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1, Properties: models.KV{
				models.PropertyPyTorchClassName: "MyModel",
			}},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
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
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
		},
		{
			name:    "Success: empty pytorch class name will fallback to PyTorchModel",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
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
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
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
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", project, mock.Anything, mock.Anything).
					Return(fmt.Sprintf("gojek/mymodel-1:latest"), nil)
				return mockImgBuilder
			},
		},
		{
			name:  "Success: pytorch model with transformer",
			model: &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1, Properties: models.KV{
				models.PropertyPyTorchClassName: "MyModel",
			}},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
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
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", project, mock.Anything, mock.Anything).
					Return(fmt.Sprintf("gojek/mymodel-1:latest"), nil)
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
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
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
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					InferenceServiceName: iSvcName,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("Failed to build image"))
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

			mockStorage.AssertNumberOfCalls(t, "Save", 1)
			savedEndpoint := mockStorage.Calls[1].Arguments[0].(*models.VersionEndpoint)
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
				assert.Equal(t, iSvcName, savedEndpoint.InferenceServiceName)
			}
		})
	}
}
