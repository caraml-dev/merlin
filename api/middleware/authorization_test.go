package middleware

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	serviceMock "github.com/caraml-dev/merlin/service/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	enforcerMock "github.com/caraml-dev/mlp/api/pkg/authz/enforcer/mocks"
)

func TestAuthorizer_RequireAuthorization(t *testing.T) {
	endpointService := &serviceMock.EndpointsService{}
	modelService := &serviceMock.ModelsService{}
	enforcer := &enforcerMock.Enforcer{}
	authorizer := NewAuthorizer(enforcer, endpointService, modelService)
	tests := []struct {
		name     string
		path     string
		method   string
		expected bool
	}{
		{"All authenticated users can list projects", "/projects", "GET", false},
		{"All authenticated users can list environments", "/environments", "GET", false},
		{"All authenticated users can make changes to simulator", "/standard_transformer/simulate", "POST", false},
		{"Only authorized users can create model", "/models/100", "POST", true},
		{"Options http request does not require authorization", "/projects", "OPTIONS", false},
		{"Only authorized users can access project sub resources", "/projects/100/model_endpoints", "GET", true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, authorizer.RequireAuthorization(tt.path, tt.method))
	}
}

func TestAuthorizer_GetPermission(t *testing.T) {
	endpointService := &serviceMock.EndpointsService{}
	modelService := &serviceMock.ModelsService{}
	enforcer := &enforcerMock.Enforcer{}
	modelService.On("FindByID", mock.Anything, models.ID(9001)).Return(&models.Model{
		Project: mlp.Project{
			ID: 1003,
		},
	}, nil)
	authorizer := NewAuthorizer(enforcer, endpointService, modelService)

	tests := []struct {
		name           string
		path           string
		method         string
		query          string
		vars           map[string]string
		requestBodyStr string
		expected       string
	}{
		{"project sub-resource permission", "/projects/1003/model_endpoints", "GET", "", map[string]string{}, "", "mlp.projects.1003.get"},
		{"model permission", "/models/9001/endpoints", "POST", "", map[string]string{"model_id": "9001"}, "", "mlp.projects.1003.post"},
		{"model versioned endpoint permission", "/models/9001/versions/3/endpoint/8e9624e0-efd3-44c9-941d-e645d5f680e8/containers", "GET", "", map[string]string{"model_id": "9001"}, "", "mlp.projects.1003.get"},
		{"log permission", "/logs?model_id=9001", "GET", "", map[string]string{}, "", "mlp.projects.1003.get"},
	}
	for _, tt := range tests {
		reqBody := bytes.NewBufferString(tt.requestBodyStr)
		testRequest, err := http.NewRequest(tt.method, tt.path, reqBody)
		if len(tt.vars) > 0 {
			testRequest = mux.SetURLVars(testRequest, tt.vars)
		}
		require.NoError(t, err)
		permission, err := authorizer.GetPermission(testRequest)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, permission)
	}
}
