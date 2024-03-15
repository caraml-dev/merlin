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
	"strconv"
	"strings"

	mlpclient "github.com/caraml-dev/mlp/api/client"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/pkg/autoscaling"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/caraml-dev/merlin/pkg/protocol"
	transformerpkg "github.com/caraml-dev/merlin/pkg/transformer"
	"knative.dev/pkg/apis"
)

const (
	RevisionPrefix = "r"
)

type Service struct {
	Name              string
	ModelName         string
	ModelVersion      string
	RevisionID        ID
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
	// CurrentIsvcName is the name of the current running/serving InferenceService's revision
	CurrentIsvcName             string
	EnabledModelObservability   bool
	ModelSchema                 *ModelSchema
	PredictorUPIOverHTTPEnabled bool
}

func NewService(model *Model, version *Version, modelOpt *ModelOption, endpoint *VersionEndpoint) *Service {
	return &Service{
		Name:            CreateInferenceServiceName(model.Name, version.ID.String(), endpoint.RevisionID.String()),
		ModelName:       model.Name,
		ModelVersion:    version.ID.String(),
		RevisionID:      endpoint.RevisionID,
		Namespace:       model.Project.Name,
		ArtifactURI:     version.ArtifactURI,
		Type:            model.Type,
		Options:         modelOpt,
		ResourceRequest: endpoint.ResourceRequest,
		EnvVars:         endpoint.EnvVars,
		Metadata: Metadata{
			App:       model.Name,
			Component: ComponentModelVersion,
			Stream:    model.Project.Stream,
			Team:      model.Project.Team,
			Labels:    MergeProjectVersionLabels(model.Project.Labels, version.Labels),
		},
		Transformer:                 endpoint.Transformer,
		Logger:                      endpoint.Logger,
		DeploymentMode:              endpoint.DeploymentMode,
		AutoscalingPolicy:           endpoint.AutoscalingPolicy,
		Protocol:                    endpoint.Protocol,
		CurrentIsvcName:             endpoint.InferenceServiceName,
		EnabledModelObservability:   endpoint.EnableModelObservability,
		ModelSchema:                 version.ModelSchema,
		PredictorUPIOverHTTPEnabled: predictorUPIOverHTTPEnabled(endpoint.Transformer, endpoint.Protocol),
	}
}

func predictorUPIOverHTTPEnabled(transformer *Transformer, modelProtocol protocol.Protocol) bool {
	if transformer == nil || !transformer.Enabled || modelProtocol != protocol.UpiV1 {
		return false
	}
	for _, val := range transformer.EnvVars {
		if val.Name == transformerpkg.PredictorUPIHTTPEnabled {
			value, err := strconv.ParseBool(val.Value)
			if err != nil {
				return false
			}
			return value
		}
	}
	return false
}

func (svc *Service) PredictorProtocol() protocol.Protocol {
	if svc.PredictorUPIOverHTTPEnabled {
		return protocol.HttpJson
	}
	return svc.Protocol
}

func (svc *Service) GetPredictionLogTopic() string {
	return fmt.Sprintf("caraml-%s-%s-prediction-log", svc.Namespace, svc.ModelName)
}

func (svc *Service) GetPredictionLogTopicForVersion() string {
	return getPredictionLogTopicForVersion(svc.Namespace, svc.ModelName, svc.ModelVersion)
}

func getPredictionLogTopicForVersion(project string, modelName string, modelVersion string) string {
	return fmt.Sprintf("caraml-%s-%s-%s-prediction-log", project, modelName, modelVersion)
}

func MergeProjectVersionLabels(projectLabels mlp.Labels, versionLabels KV) mlp.Labels {
	projectLabelsMap := map[string]int{}
	updatedLabels := make(mlp.Labels, 0)
	for index, projectLabel := range projectLabels {
		projectLabelsMap[projectLabel.Key] = index
		updatedLabels = append(updatedLabels, projectLabel)
	}

	for versionLabelKey, versionLabelValue := range versionLabels {
		if _, labelExists := projectLabelsMap[versionLabelKey]; labelExists {
			index := projectLabelsMap[versionLabelKey]
			updatedLabels[index].Value = fmt.Sprint(versionLabelValue)
			continue
		}

		updatedLabels = append(updatedLabels, mlpclient.Label{
			Key:   versionLabelKey,
			Value: fmt.Sprint(versionLabelValue),
		})
	}

	return updatedLabels
}

func CreateInferenceServiceName(modelName, versionID, revisionID string) string {
	if revisionID == "" || revisionID == "0" {
		// This is for backward compatibility, when the endpoint / isvc name didn't include the revision number
		return fmt.Sprintf("%s-%s", modelName, versionID)
	}
	return fmt.Sprintf("%s-%s-%s%s", modelName, versionID, RevisionPrefix, revisionID)
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
