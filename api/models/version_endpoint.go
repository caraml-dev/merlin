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
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	// The field name has to be prefixed with the related struct name
	// in order for gorm Preload to work with association_foreignkey
	VersionID            ID               `json:"version_id"`
	VersionModelID       ID               `json:"model_id"`
	Status               EndpointStatus   `json:"status"`
	URL                  string           `json:"url" gorm:"url"`
	ServiceName          string           `json:"service_name" gorm:"service_name"`
	InferenceServiceName string           `json:"-" gorm:"inference_service_name"`
	Namespace            string           `json:"-" gorm:"namespace"`
	MonitoringURL        string           `json:"monitoring_url,omitempty" gorm:"-"`
	Environment          *Environment     `json:"environment" gorm:"association_foreignkey:Name;"`
	EnvironmentName      string           `json:"environment_name"`
	Message              string           `json:"message"`
	ResourceRequest      *ResourceRequest `json:"resource_request" gorm:"resource_request"`
	EnvVars              EnvVars          `json:"env_vars" gorm:"column:env_vars"`
	Transformer          *Transformer     `json:"transformer,omitempty" gorm:"foreignKey:VersionEndpointID"`
	Logger               *Logger          `json:"logger,omitempty" gorm:"logger"`
	DeploymentType       DeploymentMode   `json:"deployment_mode" gorm:"deployment_mode"`

	CreatedUpdated
}

type DeploymentMode string

const (
	ServerlessDeploymentMode = "serverless"
	RawDeploymentMode        = "raw_deployment"
)

func NewVersionEndpoint(env *Environment, project mlp.Project, model *Model, version *Version, monitoringConfig config.MonitoringConfig) *VersionEndpoint {
	id := uuid.New()
	errRaised := &VersionEndpoint{
		ID:                   id,
		VersionID:            version.ID,
		VersionModelID:       version.ModelID,
		Namespace:            project.Name,
		InferenceServiceName: fmt.Sprintf("%s-%s", model.Name, version.ID.String()),
		Status:               EndpointPending,
		EnvironmentName:      env.Name,
		Environment:          env,
		ResourceRequest:      env.DefaultResourceRequest,
	}

	if monitoringConfig.MonitoringEnabled {
		errRaised.UpdateMonitoringURL(monitoringConfig.MonitoringBaseURL, EndpointMonitoringURLParams{env.Cluster, project.Name, model.Name, version.ID.String()})
	}
	return errRaised
}

func (errRaised *VersionEndpoint) IsRunning() bool {
	return errRaised.Status == EndpointPending || errRaised.Status == EndpointRunning
}

func (errRaised *VersionEndpoint) IsServing() bool {
	return errRaised.Status == EndpointServing
}

func (errRaised *VersionEndpoint) HostURL() string {
	if errRaised.URL == "" {
		return ""
	}

	url, err := url.Parse(errRaised.URL)
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

func (errRaised *VersionEndpoint) UpdateMonitoringURL(baseURL string, params EndpointMonitoringURLParams) {
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

	errRaised.MonitoringURL = url.String()
}
