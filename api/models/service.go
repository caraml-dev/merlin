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
	"github.com/gojek/merlin/pkg/protocol"
	"knative.dev/pkg/apis"
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
	Protocol          protocol.Protocol
}

func NewService(model *Model, version *Version, modelOpt *ModelOption, endpoint *VersionEndpoint) *Service {
	return &Service{
		Name:            CreateInferenceServiceName(model.Name, version.ID.String()),
		Namespace:       model.Project.Name,
		ArtifactURI:     version.ArtifactURI,
		Type:            model.Type,
		Options:         modelOpt,
		ResourceRequest: endpoint.ResourceRequest,
		EnvVars:         endpoint.EnvVars,
		Metadata: Metadata{
			Team:        model.Project.Team,
			Stream:      model.Project.Stream,
			App:         model.Name,
			Environment: endpoint.EnvironmentName,
			Labels:      model.Project.Labels,
		},
		Transformer:       endpoint.Transformer,
		Logger:            endpoint.Logger,
		DeploymentMode:    endpoint.DeploymentMode,
		AutoscalingPolicy: endpoint.AutoscalingPolicy,
		Protocol:          endpoint.Protocol,
	}
}

func CreateInferenceServiceName(modelName string, versionID string) string {
	return fmt.Sprintf("%s-%s", modelName, versionID)
}

func GetInferenceURL(url *apis.URL, inferenceServiceName string, protocolValue protocol.Protocol) string {
	switch protocolValue {
	case protocol.UpiV1:
		// return only host name
		return url.Host
	default:
		modelPath := "v1/models"
		inferenceURLSuffix := fmt.Sprintf("%s/%s", modelPath, inferenceServiceName)
		if strings.HasSuffix(url.String(), inferenceURLSuffix) {
			return url.String()
		}

		return fmt.Sprintf("%s/%s", url, inferenceURLSuffix)
	}
}
