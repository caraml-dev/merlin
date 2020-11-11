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

package models

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
)

type VersionEndpoint struct {
	Id uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	// The field name has to be prefixed with the related struct name
	// in order for gorm Preload to work with association_foreignkey
	VersionId            Id               `json:"version_id"`
	VersionModelId       Id               `json:"model_id"`
	Status               EndpointStatus   `json:"status"`
	Url                  string           `json:"url" gorm:"url"`
	ServiceName          string           `json:"service_name" gorm:"service_name"`
	InferenceServiceName string           `json:"-" gorm:"inference_service_name"`
	Namespace            string           `json:"-" gorm:"namespace"`
	MonitoringUrl        string           `json:"monitoring_url,omitempty" gorm:"-"`
	Environment          *Environment     `json:"environment" gorm:"association_foreignkey:Name;"`
	EnvironmentName      string           `json:"environment_name"`
	Message              string           `json:"message"`
	ResourceRequest      *ResourceRequest `json:"resource_request" gorm:"resource_request"`
	EnvVars              EnvVars          `json:"env_vars" gorm:"column:env_vars"`
	Transformer          *Transformer     `json:"transformer,omitempty" gorm:"foreignKey:VersionEndpointID"`

	CreatedUpdated
}

func NewVersionEndpoint(env *Environment, project mlp.Project, model *Model, version *Version, monitoringConfig config.MonitoringConfig) *VersionEndpoint {
	id := uuid.New()
	e := &VersionEndpoint{
		Id:                   id,
		VersionId:            version.Id,
		VersionModelId:       version.ModelId,
		Namespace:            project.Name,
		InferenceServiceName: fmt.Sprintf("%s-%s", model.Name, version.Id.String()),
		Status:               EndpointPending,
		EnvironmentName:      env.Name,
		Environment:          env,
		ResourceRequest:      env.DefaultResourceRequest,
	}

	if monitoringConfig.MonitoringEnabled {
		e.UpdateMonitoringUrl(monitoringConfig.MonitoringBaseURL, EndpointMonitoringURLParams{env.Cluster, project.Name, model.Name, version.Id.String()})
	}
	return e
}

func (e *VersionEndpoint) IsRunning() bool {
	return e.Status == EndpointPending || e.Status == EndpointRunning
}

func (e *VersionEndpoint) IsServing() bool {
	return e.Status == EndpointServing
}

func (e *VersionEndpoint) HostURL() string {
	if e.Url == "" {
		return ""
	}

	url, err := url.Parse(e.Url)
	if err != nil {
		return ""
	}

	return url.Hostname()
}

type EndpointMonitoringURLParams struct {
	Cluster      string
	Project      string
	Model        string
	ModelVersion string
}

func (e *VersionEndpoint) UpdateMonitoringUrl(baseURL string, params EndpointMonitoringURLParams) {
	url, _ := url.Parse(baseURL)

	q := url.Query()
	if params.Cluster != "" {
		q.Set("var-cluster", params.Cluster)
	}
	if params.Project != "" {
		q.Set("var-project", params.Project)
	}
	if params.Model != "" {
		q.Set("var-model", params.Model)
	}
	if params.ModelVersion != "" {
		q.Set("var-model_version", params.ModelVersion)
	}

	url.RawQuery = q.Encode()

	e.MonitoringUrl = url.String()
}
