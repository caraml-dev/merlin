package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/google/uuid"
)

func TestDeploymentController_ListDeployments(t *testing.T) {
	endpointUUID := uuid.New()
	endpointUUIDString := fmt.Sprint(endpointUUID)

	createdUpdated := models.CreatedUpdated{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	testCases := []struct {
		desc              string
		vars              map[string]string
		deploymentService func() *mocks.DeploymentService
		expected          *Response
	}{
		{
			desc: "Should success list deployments",
			vars: map[string]string{
				"model_id":    "model",
				"version_id":  "1",
				"endpoint_id": endpointUUIDString,
			},
			deploymentService: func() *mocks.DeploymentService {
				mockSvc := &mocks.DeploymentService{}
				mockSvc.On("ListDeployments", "model", "1", endpointUUIDString).Return([]*models.Deployment{
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
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Deployment{
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
				},
			},
		},
		{
			desc: "Should return 500 when failed fetching list of deployments",
			vars: map[string]string{
				"model_id":    "model",
				"version_id":  "1",
				"endpoint_id": endpointUUIDString,
			},
			deploymentService: func() *mocks.DeploymentService {
				mockSvc := &mocks.DeploymentService{}
				mockSvc.On("ListDeployments", "model", "1", endpointUUIDString).Return(nil, fmt.Errorf("Database is down"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{
					Message: "Error listing deployments: Database is down",
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.deploymentService()
			ctl := &DeploymentController{
				AppContext: &AppContext{
					DeploymentService: mockSvc,
				},
			}
			resp := ctl.ListDeployments(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}
