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

	"github.com/google/uuid"
)

const (
	onlineInferenceLabelTemplate = "serving.kserve.io/inferenceservice=%s"
	batchInferenceLabelTemplate  = "prediction-job-id=%s"

	ImageBuilderComponentType     = "image_builder"
	ModelComponentType            = "model"
	PredictorComponentType        = "predictor"
	TransformerComponentType      = "transformer"
	PDBComponentType              = "pdb" // Pod disruption budget
	BatchJobDriverComponentType   = "batch_job_driver"
	BatchJobExecutorComponentType = "batch_job_executor"
)

type Container struct {
	Name              string    `json:"name"`
	PodName           string    `json:"pod_name"`
	ComponentType     string    `json:"component_type"`
	Namespace         string    `json:"namespace"`
	Cluster           string    `json:"cluster"`
	GcpProject        string    `json:"gcp_project"`
	VersionEndpointID uuid.UUID `json:"version_endpoint_id"`
}

func NewContainer(name string,
	podName string,
	namespace string,
	cluster string,
	gcpProject string,
) *Container {
	return &Container{
		Name:          name,
		PodName:       podName,
		ComponentType: componentType(name, podName),
		Namespace:     namespace,
		Cluster:       cluster,
		GcpProject:    gcpProject,
	}
}

func componentType(containerName, podName string) string {
	componentType := ""
	if strings.Contains(podName, "predictor") {
		componentType = ModelComponentType
	} else if strings.Contains(podName, "transformer") {
		componentType = TransformerComponentType
	} else if strings.Contains(podName, "driver") || containerName == "spark-kubernetes-driver" {
		componentType = BatchJobDriverComponentType
	} else if strings.Contains(podName, "exec") || containerName == "executor" {
		componentType = BatchJobExecutorComponentType
	} else if containerName == "pyfunc-image-builder" {
		componentType = ImageBuilderComponentType
	}
	return componentType
}

func OnlineInferencePodLabelSelector(modelName string, versionID string) string {
	serviceName := CreateInferenceServiceName(modelName, versionID)
	return fmt.Sprintf(onlineInferenceLabelTemplate, serviceName)
}

func BatchInferencePodLabelSelector(predictionJobID string) string {
	return fmt.Sprintf(batchInferenceLabelTemplate, predictionJobID)
}
