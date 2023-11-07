package service

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/storage/mocks"
	"github.com/google/uuid"
)

func Test_deploymentService_ListDeployments(t *testing.T) {
	endpointUUID := uuid.New()
	endpointUUIDString := fmt.Sprint(endpointUUID)

	createdUpdated := models.CreatedUpdated{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	type args struct {
		modelID      string
		versionID    string
		endpointUUID string
	}
	tests := []struct {
		name                  string
		args                  args
		mockDeploymentStorage func() *mocks.DeploymentStorage
		want                  []*models.Deployment
		wantErr               bool
	}{
		{
			name: "success",
			args: args{
				modelID:      "model",
				versionID:    "1",
				endpointUUID: endpointUUIDString,
			},
			mockDeploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("ListInModelVersion", "model", "1", endpointUUIDString).Return([]*models.Deployment{
					{
						ID:                models.ID(1),
						ProjectID:         models.ID(1),
						VersionModelID:    models.ID(1),
						VersionID:         models.ID(1),
						VersionEndpointID: endpointUUID,
						Status:            models.EndpointRunning,
						Error:             "",
						CreatedUpdated:    createdUpdated,
					},
				}, nil)
				return mockStorage
			},
			want: []*models.Deployment{{
				ID:                models.ID(1),
				ProjectID:         models.ID(1),
				VersionModelID:    models.ID(1),
				VersionID:         models.ID(1),
				VersionEndpointID: endpointUUID,
				Status:            models.EndpointRunning,
				Error:             "",
				CreatedUpdated:    createdUpdated,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeploymentStorage := tt.mockDeploymentStorage()

			service := &deploymentService{
				storage: mockDeploymentStorage,
			}
			got, err := service.ListDeployments(tt.args.modelID, tt.args.versionID, tt.args.endpointUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.ListDeployments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deploymentService.ListDeployments() = %v, want %v", got, tt.want)
			}
		})
	}
}
