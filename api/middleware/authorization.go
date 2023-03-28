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

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
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

var methodMapping = map[string]string{
	http.MethodGet:     enforcer.ActionRead,
	http.MethodHead:    enforcer.ActionRead,
	http.MethodPost:    enforcer.ActionCreate,
	http.MethodPut:     enforcer.ActionUpdate,
	http.MethodPatch:   enforcer.ActionUpdate,
	http.MethodDelete:  enforcer.ActionDelete,
	http.MethodConnect: enforcer.ActionRead,
	http.MethodOptions: enforcer.ActionRead,
	http.MethodTrace:   enforcer.ActionRead,
}

func (a *Authorizer) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resource, err := a.getResource(r)
		if err != nil {
			jsonError(w, fmt.Sprintf("Error while checking authorization: %s", err), http.StatusInternalServerError)
			return
		}

		action := methodToAction(r.Method)
		user := r.Header.Get("User-Email")

		allowed, err := a.authEnforcer.Enforce(user, resource, action)
		if err != nil {
			jsonError(w, fmt.Sprintf("Error while checking authorization: %s", err), http.StatusInternalServerError)
			return
		}
		if !*allowed {
			jsonError(w, fmt.Sprintf("%s is not authorized to execute %s on %s", user, action, resource), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authorizer) getResource(r *http.Request) (string, error) {
	ctx := r.Context()

	resource := strings.Replace(strings.TrimPrefix(r.URL.Path, "/"), "/", ":", -1)
	// workaround since models/** APIS is not under projects/**
	if strings.HasPrefix(resource, "models") {
		vars := mux.Vars(r)
		modelID, _ := models.ParseID(vars["model_id"])
		model, err := a.modelService.FindByID(ctx, modelID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("projects:%d:%s", model.Project.ID, resource), nil
	}

	// workaround since logs API is not under projects/**
	if strings.HasPrefix(resource, "logs") {
		modelID, err := models.ParseID(r.URL.Query().Get("model_id"))
		if err != nil {
			return "", err
		}

		model, err := a.modelService.FindByID(ctx, modelID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("projects:%d:logs", model.Project.ID), nil
	}

	return resource, nil
}

func methodToAction(method string) string {
	return methodMapping[method]
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
