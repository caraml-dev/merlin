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

package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/caraml-dev/mlp/api/pkg/authz/enforcer"
	"github.com/gorilla/mux"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service"
)

func NewAuthorizer(enforcer enforcer.Enforcer, endpointService service.EndpointsService, modelService service.ModelsService) *Authorizer {
	return &Authorizer{
		authEnforcer:    enforcer,
		endpointService: endpointService,
		modelService:    modelService,
	}
}

type Authorizer struct {
	authEnforcer    enforcer.Enforcer
	endpointService service.EndpointsService
	modelService    service.ModelsService
}

type Operation struct {
	RequestPath   string
	RequestMethod []string
}

var publicOperations = []Operation{
	{"/projects", []string{http.MethodGet}},
	{"/environments", []string{http.MethodGet}},
	{"/standard-transformer/simulate", []string{http.MethodGet}},
	{"/alerts/teams", []string{http.MethodGet}},
}

func (a *Authorizer) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.RequireAuthorization(r.URL.Path, r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		permission, err := a.GetPermission(r)
		if err != nil {
			jsonError(w, fmt.Sprintf("Error while checking authorization: %s", err), http.StatusInternalServerError)
			return
		}

		user := r.Header.Get("User-Email")

		allowed, err := a.authEnforcer.IsUserGrantedPermission(r.Context(), user, permission)
		if err != nil {
			jsonError(w, fmt.Sprintf("Error while checking authorization: %s", err), http.StatusInternalServerError)
			return
		}
		if !allowed {
			jsonError(w, fmt.Sprintf("%s does not have the permission:%s ", user, permission), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authorizer) RequireAuthorization(requestPath string, requestMethod string) bool {
	if requestMethod == http.MethodOptions {
		return false
	}

	for _, operation := range publicOperations {
		if operation.RequestPath == requestPath && slices.Contains(operation.RequestMethod, requestMethod) {
			return false
		}
	}
	return true
}

func (a *Authorizer) GetPermission(r *http.Request) (string, error) {
	ctx := r.Context()
	resource := strings.Replace(strings.TrimPrefix(r.URL.Path, "/"), "/", ".", -1)

	// Current paths registered in Merlin are of the following formats:
	// - /alerts/teams
	// - /environments
	// - /logs
	// - /models/{model_id}/**
	// - /projects/{project_id}/**
	// - /standard_transformer/simulate
	//
	// /models and /logs endpoints will be authorized based on the respective project they belong to
	// (see implementation below), which is effectively the same as /projects/{project_id}/**.
	// Given this, we only care about the permissions up-to 2 levels deep. The rationale is that
	// if a user has READ/WRITE permissions on /projects/{project_id}, they would also have the same
	// permissions on all its sub-resources. Thus, trimming the resource identifier to aid quicker
	// authz matching and to efficiently make use of the in-memory authz cache, if enabled.

	// workaround since models/** APIS is not under projects/**
	// TODO: move the models endpoint under projects/
	if strings.HasPrefix(resource, "models") {
		vars := mux.Vars(r)
		modelID, _ := models.ParseID(vars["model_id"])
		model, err := a.modelService.FindByID(ctx, modelID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("mlp.projects.%d.%s", model.Project.ID, strings.ToLower(r.Method)), nil
	}

	// workaround since logs API is not under projects/**
	// TODO: move the logs endpoint under projects/
	if strings.HasPrefix(resource, "logs") {
		modelID, err := models.ParseID(r.URL.Query().Get("model_id"))
		if err != nil {
			return "", err
		}

		model, err := a.modelService.FindByID(ctx, modelID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("mlp.projects.%d.%s", model.Project.ID, strings.ToLower(r.Method)), nil
	}

	// Trim the resource info, greater than 2 segments
	parts := strings.Split(resource, ".")
	if len(parts) > 1 {
		parts = parts[:2]
	}
	return fmt.Sprintf("mlp.%s.%s", strings.Join(parts, "."), strings.ToLower(r.Method)), nil
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if len(msg) > 0 {
		errJSON, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{msg})

		w.Write(errJSON) //nolint:errcheck
	}
}
