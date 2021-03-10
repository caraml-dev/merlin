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

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/queue"

	"github.com/gojek/merlin/middleware"
	"github.com/gojek/merlin/mlflow"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service"
)

// AppContext contains the services of the Merlin application.
type AppContext struct {
	EnvironmentService        service.EnvironmentService
	ProjectsService           service.ProjectsService
	ModelsService             service.ModelsService
	ModelEndpointsService     service.ModelEndpointsService
	VersionsService           service.VersionsService
	EndpointsService          service.EndpointsService
	LogService                service.LogService
	PredictionJobService      service.PredictionJobService
	SecretService             service.SecretService
	ModelEndpointAlertService service.ModelEndpointAlertService
	DB                        *gorm.DB
	AuthorizationEnabled      bool
	MonitoringConfig          config.MonitoringConfig
	AlertEnabled              bool
	Enforcer                  enforcer.Enforcer
	FeastCoreClient           core.CoreServiceClient
	MlflowClient              mlflow.Client
	Dispatcher                *queue.Dispatcher
}

// Handler handles the API requests and responses.
type Handler func(r *http.Request, vars map[string]string, body interface{}) *Response

type Route struct {
	method  string
	path    string
	body    interface{}
	handler Handler
	name    string
}

func (route Route) HandlerFunc(validate *validator.Validate) http.HandlerFunc {
	var bodyType reflect.Type
	if route.body != nil {
		bodyType = reflect.TypeOf(route.body)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				vars[k] = v[0]
			}
		}

		response := func() *Response {
			vars["user"] = r.Header.Get("User-Email")
			var body interface{} = nil
			if bodyType != nil {
				body = reflect.New(bodyType).Interface()
				err := json.NewDecoder(r.Body).Decode(body)
				if err == io.EOF {
					// empty body
					return route.handler(r, vars, body)
				}

				if err != nil {
					return BadRequest(fmt.Sprintf("Failed to deserialize request body: %s", err.Error()))
				}

				if err := validate.Struct(body); err != nil {
					s := err.(validator.ValidationErrors)[0].Translate(en)
					return BadRequest(s)
				}
			}
			return route.handler(r, vars, body)
		}()

		response.WriteTo(w)
	}
}

// RawRoutes define a low level routes
// User has to implement a http.HandlerFunc to handle the routes
// Most api should use Routes instead of RawRoutes.
// RawRoutes is only used when more control of http.HandlerFunc is required.
type RawRoutes struct {
	method  string
	path    string
	handler http.Handler
	name    string
}

func NewRouter(appCtx AppContext) *mux.Router {
	validate := NewValidator()
	environmentController := EnvironmentController{&appCtx}
	projectsController := ProjectsController{&appCtx}
	modelsController := ModelsController{&appCtx}
	modelEndpointsController := ModelEndpointsController{&appCtx}
	versionsController := VersionsController{&appCtx}
	endpointsController := EndpointsController{&appCtx}
	predictionJobController := PredictionJobController{&appCtx}
	logController := LogController{&appCtx}
	secretController := SecretsController{&appCtx}
	alertsController := AlertsController{&appCtx}

	routes := []Route{
		// Environment API
		{http.MethodGet, "/environments", nil, environmentController.ListEnvironments, "ListEnvironments"},

		// Project API
		{http.MethodGet, "/projects/{project_id:[0-9]+}", nil, projectsController.GetProject, "GetProject"},
		{http.MethodGet, "/projects", nil, projectsController.ListProjects, "ListProjects"},

		// Secret Management API
		{http.MethodGet, "/projects/{project_id:[0-9]+}/secrets", nil, secretController.ListSecret, "ListSecret"},
		{http.MethodPost, "/projects/{project_id:[0-9]+}/secrets", mlp.Secret{}, secretController.CreateSecret, "CreateSecret"},
		{http.MethodPatch, "/projects/{project_id:[0-9]+}/secrets/{secret_id}", mlp.Secret{}, secretController.UpdateSecret, "UpdateSecret"},
		{http.MethodDelete, "/projects/{project_id:[0-9]+}/secrets/{secret_id}", nil, secretController.DeleteSecret, "DeleteSecret"},

		// Model API
		{http.MethodGet, "/projects/{project_id:[0-9]+}/models/{model_id:[0-9]+}", nil, modelsController.GetModel, "GetModel"},
		{http.MethodGet, "/projects/{project_id:[0-9]+}/models", nil, modelsController.ListModels, "ListModels"},
		{http.MethodPost, "/projects/{project_id:[0-9]+}/models", models.Model{}, modelsController.CreateModel, "CreateModel"},

		// Model Endpoints API
		{http.MethodGet, "/projects/{project_id:[0-9]+}/model_endpoints", nil, modelEndpointsController.ListModelEndpointInProject, "ListModelEndpointInProject"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/endpoints", nil, modelEndpointsController.ListModelEndpoints, "ListModelEndpoint"},
		{http.MethodPost, "/models/{model_id:[0-9]+}/endpoints", models.ModelEndpoint{}, modelEndpointsController.CreateModelEndpoint, "CreateModelEndpoint"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/endpoints/{model_endpoint_id}", nil, modelEndpointsController.GetModelEndpoint, "GetModelEndpoint"},
		{http.MethodPut, "/models/{model_id:[0-9]+}/endpoints/{model_endpoint_id}", models.ModelEndpoint{}, modelEndpointsController.UpdateModelEndpoint, "UpdateModelEndpoint"},
		{http.MethodDelete, "/models/{model_id:[0-9]+}/endpoints/{model_endpoint_id}", nil, modelEndpointsController.DeleteModelEndpoint, "DeleteModelEndpoint"},

		// Version API
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions", nil, versionsController.ListVersions, "ListVersions"},
		{http.MethodPost, "/models/{model_id:[0-9]+}/versions", models.VersionPost{}, versionsController.CreateVersion, "CreateVersion"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}", nil, versionsController.GetVersion, "GetVersion"},
		{http.MethodPatch, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}", models.VersionPatch{}, versionsController.PatchVersion, "PatchVersion"},

		// Version Endpoint API
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint", nil, endpointsController.ListEndpoint, "ListEndpoint"},
		{http.MethodPost, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint", models.VersionEndpoint{}, endpointsController.CreateEndpoint, "CreateEndpoint"},
		// To maintain backward compatibility with SDK v0.1.0
		{http.MethodDelete, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint", nil, endpointsController.DeleteEndpoint, "DeleteDefaultEndpoint"},

		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint/{endpoint_id}", nil, endpointsController.GetEndpoint, "GetEndpoint"},
		{http.MethodPut, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint/{endpoint_id}", models.VersionEndpoint{}, endpointsController.UpdateEndpoint, "UpdateEndpoint"},
		{http.MethodDelete, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint/{endpoint_id}", nil, endpointsController.DeleteEndpoint, "DeleteEndpoint"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint/{endpoint_id}/containers", nil, endpointsController.ListContainers, "ListContainers"},

		// Prediction Job API
		{http.MethodGet, "/projects/{project_id:[0-9]+}/jobs", nil, predictionJobController.ListAllInProject, "ListAllPredictionJobInProject"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/jobs", nil, predictionJobController.List, "ListPredictionJob"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/jobs/{job_id:[0-9]+}", nil, predictionJobController.Get, "GetPredictionJob"},
		{http.MethodPut, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/jobs/{job_id:[0-9]+}/stop", nil, predictionJobController.Stop, "StopPredictionJob"},
		{http.MethodPost, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/jobs", models.PredictionJob{}, predictionJobController.Create, "CreatePredictionJob"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/jobs/{job_id:[0-9]+}/containers", nil, predictionJobController.ListContainers, "ListJobContainers"},
	}

	if appCtx.AlertEnabled {
		routes = append(routes, []Route{
			// Model Endpoint Alerts API
			{http.MethodGet, "/alerts/teams", nil, alertsController.ListTeams, "ListTeams"},
			{http.MethodGet, "/models/{model_id:[0-9]+}/alerts", nil, alertsController.ListModelEndpointAlerts, "ListModelEndpointAlerts"},
			{http.MethodGet, "/models/{model_id:[0-9]+}/endpoints/{model_endpoint_id}/alert", nil, alertsController.GetModelEndpointAlert, "GetModelEndpointAlert"},
			{http.MethodPost, "/models/{model_id:[0-9]+}/endpoints/{model_endpoint_id}/alert", models.ModelEndpointAlert{}, alertsController.CreateModelEndpointAlert, "CreateModelEndpointAlert"},
			{http.MethodPut, "/models/{model_id:[0-9]+}/endpoints/{model_endpoint_id}/alert", models.ModelEndpointAlert{}, alertsController.UpdateModelEndpointAlert, "UpdateModelEndpointAlert"},
		}...)
	}

	rawRoutes := []RawRoutes{
		{
			http.MethodGet, "/logs", http.HandlerFunc(logController.ReadLog), "ReadLogs",
		},
	}

	var authzMiddleware *middleware.Authorizer
	if appCtx.AuthorizationEnabled {
		authzMiddleware = middleware.NewAuthorizer(appCtx.Enforcer, appCtx.EndpointsService, appCtx.ModelsService)
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, r := range routes {
		_, handler := newrelic.WrapHandle(r.name, r.HandlerFunc(validate))

		if appCtx.AuthorizationEnabled {
			handler = authzMiddleware.AuthorizationMiddleware(handler)
		}

		router.Name(r.name).
			Methods(r.method).
			Path(r.path).
			Handler(handler)
	}

	for _, rr := range rawRoutes {
		handler := rr.handler
		if appCtx.AuthorizationEnabled {
			handler = authzMiddleware.AuthorizationMiddleware(handler)
		}

		router.Name(rr.name).
			Methods(rr.method).
			Path(rr.path).
			Handler(handler)
	}

	router.Use(recoveryHandler)

	return router
}

func recoveryHandler(next http.Handler) http.Handler {
	return sentry.Recoverer(next)
}
