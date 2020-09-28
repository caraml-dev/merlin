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

	"github.com/google/uuid"
)

const (
	onlineInferenceLabelTemplate = "serving.kubeflow.org/inferenceservice=%s"
	batchInferenceLabelTemplate  = "prediction-job-id=%s"
)

type Container struct {
	Name              string    `json:"name"`
	PodName           string    `json:"pod_name"`
	Namespace         string    `json:"namespace"`
	Cluster           string    `json:"cluster"`
	GcpProject        string    `json:"gcp_project"`
	VersionEndpointId uuid.UUID `json:"version_endpoint_id"`
}

func OnlineInferencePodLabelSelector(modelName string, versionId string) string {
	serviceName := CreateInferenceServiceName(modelName, versionId)
	return fmt.Sprintf(onlineInferenceLabelTemplate, serviceName)
}

func BatchInferencePodLabelSelector(predictionJobId string) string {
	return fmt.Sprintf(batchInferenceLabelTemplate, predictionJobId)
}
