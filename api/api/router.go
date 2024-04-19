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
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	mlflowDelete "github.com/caraml-dev/mlp/api/pkg/client/mlflow"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"

	"github.com/caraml-dev/mlp/api/pkg/authz/enforcer"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"

	"github.com/caraml-dev/merlin/config"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/middleware"
	"github.com/caraml-dev/merlin/mlflow"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	internalValidator "github.com/caraml-dev/merlin/pkg/validator"
	"github.com/caraml-dev/merlin/service"
)

// AppContext contains the services of the Merlin application.
type AppContext struct {
	DB       *gorm.DB
	Enforcer enforcer.Enforcer

	DeploymentService         service.DeploymentService
	EnvironmentService        service.EnvironmentService
	ProjectsService           service.ProjectsService
	ModelsService             service.ModelsService
	ModelEndpointsService     service.ModelEndpointsService
	VersionsService           service.VersionsService
	VersionImageService       service.VersionImageService
	EndpointsService          service.EndpointsService
	LogService                service.LogService
	PredictionJobService      service.PredictionJobService
	SecretService             service.SecretService
	ModelEndpointAlertService service.ModelEndpointAlertService
	TransformerService        service.TransformerService
	MlflowDeleteService       mlflowDelete.Service
	ModelSchemaService        service.ModelSchemaService

	AuthorizationEnabled      bool
	FeatureToggleConfig       config.FeatureToggleConfig
	StandardTransformerConfig config.StandardTransformerConfig

	FeastCoreClient core.CoreServiceClient
	MlflowClient    mlflow.Client
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
				if errors.Is(err, io.EOF) {
					// empty body
					return route.handler(r, vars, body)
				}

				if err != nil {
					return BadRequest(fmt.Sprintf("Failed to deserialize request body: %s", err.Error()))
				}

				if err := validate.Struct(body); err != nil {
					s := err.(validator.ValidationErrors)[0].Translate(internalValidator.EN) // nolint:errorlint
					return BadRequest(s)
				}
			}
			return route.handler(r, vars, body)
		}()

		// Log unsuccessful responses
		if response != nil && !response.IsSuccess() {
			log.GetLogger().Errorw("Request failed",
				"route", route.name,
				"path", r.URL.Path,
				"status", response.code,
				"response", response.data,
				"stacktrace", response.stacktrace, // Override default stacktrace produced by logger
			)
		}

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

func NewRouter(appCtx AppContext) (*mux.Router, error) {
	validate, err := internalValidator.NewValidator()
	if err != nil {
		return nil, err
	}
	deploymentController := DeploymentController{&appCtx}
	environmentController := EnvironmentController{&appCtx}
	projectsController := ProjectsController{&appCtx}
	modelEndpointsController := ModelEndpointsController{&appCtx}
	versionsController := VersionsController{&appCtx}
	versionImageController := VersionImageController{&appCtx}
	modelsController := ModelsController{&appCtx, &versionsController}
	endpointsController := EndpointsController{&appCtx}
	predictionJobController := PredictionJobController{&appCtx}
	logController := LogController{&appCtx}
	secretController := SecretsController{&appCtx}
	alertsController := AlertsController{&appCtx}
	transformerController := TransformerController{&appCtx}
	modelSchemaController := ModelSchemaController{&appCtx}

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
		{http.MethodDelete, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}", nil, versionsController.DeleteVersion, "DeleteVersion"},

		// Version Image API
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/image", nil, versionImageController.GetImage, "GetImage"},
		{http.MethodPut, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/image", models.BuildImageOptions{}, versionImageController.BuildImage, "BuildImage"},

		// Version Endpoint API
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint", nil, endpointsController.ListEndpoint, "ListEndpoint"},
		{http.MethodPost, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint", models.VersionEndpoint{}, endpointsController.CreateEndpoint, "CreateEndpoint"},
		// To maintain backward compatibility with SDK v0.1.0
		{http.MethodDelete, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoint", nil, endpointsController.DeleteEndpoint, "DeleteDefaultEndpoint"},

		// Deployments API
		{http.MethodGet, "/models/{model_id:[0-9]+}/versions/{version_id:[0-9]+}/endpoints/{endpoint_id}/deployments", nil, deploymentController.ListDeployments, "ListDeployments"},

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

		// Standard Transformer Simulation API
		{http.MethodPost, "/standard_transformer/simulate", models.TransformerSimulation{}, transformerController.SimulateTransformer, "SimulateTransformer"},

		// Model Schema API
		{http.MethodGet, "/models/{model_id:[0-9]+}/schemas", models.ModelSchema{}, modelSchemaController.GetAllSchemas, "GetAllSchemas"},
		{http.MethodGet, "/models/{model_id:[0-9]+}/schemas/{schema_id:[0-9]+}", models.ModelSchema{}, modelSchemaController.GetSchema, "GetSchemaDetail"},
		{http.MethodPut, "/models/{model_id:[0-9]+}/schemas", models.ModelSchema{}, modelSchemaController.CreateOrUpdateSchema, "CreateOrUpdateSchema"},
		{http.MethodDelete, "/models/{model_id:[0-9]+}/schemas/{schema_id:[0-9]+}", models.ModelSchema{}, modelSchemaController.DeleteSchema, "DeleteSchema"},
	}

	if appCtx.FeatureToggleConfig.ModelDeletionConfig.Enabled {
		// Model deletion API
		routes = append(routes, Route{http.MethodDelete, "/projects/{project_id:[0-9]+}/models/{model_id:[0-9]+}", nil, modelsController.DeleteModel, "DeleteModel"})
	}

	if appCtx.FeatureToggleConfig.AlertConfig.AlertEnabled {
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
		{http.MethodGet, "/logs", http.HandlerFunc(logController.ReadLog), "ReadLogs"},
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

	router.Use(prometheusMiddleware)
	router.Use(recoveryHandler)

	return router, nil
}

func recoveryHandler(next http.Handler) http.Handler {
	return sentry.Recoverer(next)
}

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "merlin_http_duration_ms",
	Help:    "Duration of HTTP requests.",
	Buckets: prometheus.ExponentialBuckets(10, 2, 10),
}, []string{"name", "path", "user_agent", "status_code"})

type statusCodeRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusCodeRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusCodeRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusCodeRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		route := mux.CurrentRoute(r)

		name := route.GetName()
		path, _ := route.GetPathTemplate() // returns template (`/obj/{id}`) rather than actual path (`/obj/123`) which would avoid a cardinality explosion

		userAgent := r.UserAgent()
		ua := ""
		if strings.Contains(userAgent, "merlin-sdk") || strings.Contains(userAgent, "python") {
			ua = userAgent
		}

		startTime := time.Now()
		next.ServeHTTP(recorder, r)
		durationMs := time.Since(startTime).Milliseconds()

		httpDuration.WithLabelValues(name, path, ua, fmt.Sprint(recorder.statusCode)).Observe(float64(durationMs))
	})
}
