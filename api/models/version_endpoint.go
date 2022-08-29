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

	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
	"github.com/gojek/merlin/pkg/protocol"
	"github.com/google/uuid"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
)

// VersionEndpoint represents the deployment of a model version in certain environment
type VersionEndpoint struct {
	// ID unique id of the version endpoint
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	// VersionID model version id from which the version endpoint is created
	// The field name has to be prefixed with the related struct name
	// in order for gorm Preload to work with association_foreignkey
	VersionID ID `json:"version_id"`
	// VersionModelID model id from which the version endpoint is created
	VersionModelID ID `json:"model_id"`
	// Status status of the version endpoint
	Status EndpointStatus `json:"status"`
	// URL url of the version endpoint
	URL string `json:"url" gorm:"url"`
	// ServiceName service name
	ServiceName string `json:"service_name" gorm:"service_name"`
	// InferenceServiceName name of inference service
	InferenceServiceName string `json:"-" gorm:"inference_service_name"`
	// Namespace namespace where the version is deployed at
	Namespace string `json:"-" gorm:"namespace"`
	// MonitoringURL URL pointing to the version endpoint's dashboard
	MonitoringURL string `json:"monitoring_url,omitempty" gorm:"-"`
	// Environment environment where the version endpoint is deployed
	Environment *Environment `json:"environment" gorm:"association_foreignkey:Name;"`
	// EnvironmentName environment name where the version endpoint is deployed
	EnvironmentName string `json:"environment_name"`
	// Message message containing the latest deployment result
	Message string `json:"message" gorm:"message"`
	// ResourceRequest resource requested by this version endpoint (CPU, Memory, replicas)
	ResourceRequest *ResourceRequest `json:"resource_request" gorm:"resource_request"`
	// EnvVars environment variable to be set in the version endpoints'deployment
	EnvVars EnvVars `json:"env_vars" gorm:"column:env_vars"`
	// Transformer transformer configuration
	Transformer *Transformer `json:"transformer,omitempty" gorm:"foreignKey:VersionEndpointID"`
	// Logger logger configuration
	Logger *Logger `json:"logger,omitempty" gorm:"logger"`
	// DeploymentMode deployment mode of the version endpoint, it can be raw_deployment or serverless
	DeploymentMode deployment.Mode `json:"deployment_mode" gorm:"deployment_mode"`
	// AutoscalingPolicy controls the conditions when autoscaling should be triggered
	AutoscalingPolicy *autoscaling.AutoscalingPolicy `json:"autoscaling_policy" gorm:"autoscaling_policy"`
	// Protocol to be used when deploying the model
	Protocol protocol.Protocol `json:"protocol" gorm:"protocol"`
	CreatedUpdated
}

const defaultWorkers = 1

// NewVersionEndpoint create a version endpoint with default configurations
func NewVersionEndpoint(env *Environment, project mlp.Project, model *Model, version *Version, monitoringConfig config.MonitoringConfig, deploymentMode deployment.Mode) *VersionEndpoint {
	id := uuid.New()

	var envVars EnvVars

	if deploymentMode == deployment.EmptyDeploymentMode {
		deploymentMode = deployment.ServerlessDeploymentMode
	}

	autoscalingPolicy := autoscaling.DefaultServerlessAutoscalingPolicy
	if deploymentMode == deployment.RawDeploymentMode {
		autoscalingPolicy = autoscaling.DefaultRawDeploymentAutoscalingPolicy
	}

	ve := &VersionEndpoint{
		ID:                   id,
		VersionID:            version.ID,
		VersionModelID:       version.ModelID,
		Namespace:            project.Name,
		InferenceServiceName: fmt.Sprintf("%s-%s", model.Name, version.ID.String()),
		Status:               EndpointPending,
		EnvironmentName:      env.Name,
		Environment:          env,
		ResourceRequest:      env.DefaultResourceRequest,
		DeploymentMode:       deploymentMode,
		AutoscalingPolicy:    autoscalingPolicy,
		EnvVars:              envVars,
		Protocol:             protocol.HttpJson,
	}

	if monitoringConfig.MonitoringEnabled {
		ve.UpdateMonitoringURL(monitoringConfig.MonitoringBaseURL, EndpointMonitoringURLParams{env.Cluster, project.Name, model.Name, version.ID.String()})
	}
	return ve
}

func (ve *VersionEndpoint) IsRunning() bool {
	return ve.Status == EndpointPending || ve.Status == EndpointRunning
}

func (ve *VersionEndpoint) IsServing() bool {
	return ve.Status == EndpointServing
}

func (ve *VersionEndpoint) Hostname() string {
	if ve.URL == "" {
		return ""
	}

	parsedURL, err := ve.ParsedURL()
	if err != nil {
		return ""
	}

	return parsedURL.Hostname()
}

func (ve *VersionEndpoint) Path() string {
	if ve.URL == "" {
		return ""
	}

	parsedURL, err := ve.ParsedURL()
	if err != nil {
		return ""
	}

	return parsedURL.Path
}

func (ve *VersionEndpoint) ParsedURL() (*url.URL, error) {
	parsedURL, err := url.Parse(ve.URL)
	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme == "" {
		veURL := ve.URL
		veURL = "//" + veURL
		parsedURL, err = url.Parse(veURL)
		if err != nil {
			return nil, err
		}
	}

	return parsedURL, nil
}

type EndpointMonitoringURLParams struct {
	Cluster      string
	Project      string
	Model        string
	ModelVersion string
}

func (ve *VersionEndpoint) UpdateMonitoringURL(baseURL string, params EndpointMonitoringURLParams) {
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

	ve.MonitoringURL = url.String()
}
