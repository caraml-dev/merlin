/*
 * Merlin
 *
 * API Guide for accessing Merlin's model management, deployment, and serving functionalities
 *
 * API version: 0.14.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package client

import (
	"time"
)

type VersionEndpoint struct {
	Id                string             `json:"id,omitempty"`
	VersionId         int32              `json:"version_id,omitempty"`
	Status            *EndpointStatus    `json:"status,omitempty"`
	Url               string             `json:"url,omitempty"`
	ServiceName       string             `json:"service_name,omitempty"`
	EnvironmentName   string             `json:"environment_name,omitempty"`
	Environment       *Environment       `json:"environment,omitempty"`
	MonitoringUrl     string             `json:"monitoring_url,omitempty"`
	Message           string             `json:"message,omitempty"`
	ResourceRequest   *ResourceRequest   `json:"resource_request,omitempty"`
	EnvVars           []EnvVar           `json:"env_vars,omitempty"`
	Transformer       *Transformer       `json:"transformer,omitempty"`
	Logger            *Logger            `json:"logger,omitempty"`
	DeploymentMode    *DeploymentMode    `json:"deployment_mode,omitempty"`
	AutoscalingPolicy *AutoscalingPolicy `json:"autoscaling_policy,omitempty"`
	Protocol          *Protocol          `json:"protocol,omitempty"`
	CreatedAt         time.Time          `json:"created_at,omitempty"`
	UpdatedAt         time.Time          `json:"updated_at,omitempty"`
}
