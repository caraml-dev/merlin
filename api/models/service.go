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
	"strings"

	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
)

type Service struct {
	Name              string
	Namespace         string
	ServiceName       string
	URL               string
	ArtifactURI       string
	Type              string
	Options           *ModelOption
	ResourceRequest   *ResourceRequest
	EnvVars           EnvVars
	Metadata          Metadata
	Transformer       *Transformer
	Logger            *Logger
	DeploymentMode    deployment.Mode
	AutoscalingPolicy *autoscaling.AutoscalingPolicy
}

func NewService(model *Model, version *Version, modelOpt *ModelOption, resource *ResourceRequest, envVars EnvVars, environment string, transformer *Transformer, logger *Logger, deploymentMode deployment.Mode, autoscalingPolicy *autoscaling.AutoscalingPolicy) *Service {
	return &Service{
		Name:            CreateInferenceServiceName(model.Name, version.ID.String()),
		Namespace:       model.Project.Name,
		ArtifactURI:     version.ArtifactURI,
		Type:            model.Type,
		Options:         modelOpt,
		ResourceRequest: resource,
		EnvVars:         envVars,
		Metadata: Metadata{
			Team:        model.Project.Team,
			Stream:      model.Project.Stream,
			App:         model.Name,
			Environment: environment,
			Labels:      model.Project.Labels,
		},
		Transformer:       transformer,
		Logger:            logger,
		DeploymentMode:    deploymentMode,
		AutoscalingPolicy: autoscalingPolicy,
	}
}

func CreateInferenceServiceName(modelName string, versionID string) string {
	return fmt.Sprintf("%s-%s", modelName, versionID)
}

func GetValidInferenceURL(url string, inferenceServiceName string) string {
	modelPath := "v1/models"
	inferenceURLSuffix := fmt.Sprintf("%s/%s", modelPath, inferenceServiceName)

	if strings.HasSuffix(url, inferenceURLSuffix) {
		return url
	}

	return fmt.Sprintf("%s/%s", url, inferenceURLSuffix)
}
