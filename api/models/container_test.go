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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlineInferencePodLabelSelector(t *testing.T) {
	modelName := "my-model"
	versionID := "1"
	revisionID := "1"
	result := OnlineInferencePodLabelSelector(modelName, versionID, revisionID)
	assert.Equal(t, "serving.kserve.io/inferenceservice=my-model-1-1", result)
}

func TestBatchInferencePodLabelSelector(t *testing.T) {
	predictionJobID := "33"
	result := BatchInferencePodLabelSelector(predictionJobID)
	assert.Equal(t, "prediction-job-id=33", result)
}

func TestNewContainer(t *testing.T) {
	type args struct {
		name       string
		podName    string
		namespace  string
		cluster    string
		gcpProject string
	}
	tests := []struct {
		name string
		args args
		want *Container
	}{
		{
			"model",
			args{
				name:       "kfserving-container",
				podName:    "test-1-1-predictor-12345-deployment",
				namespace:  "sample",
				cluster:    "test",
				gcpProject: "test-project",
			},
			&Container{
				Name:          "kfserving-container",
				PodName:       "test-1-1-predictor-12345-deployment",
				ComponentType: "model",
				Namespace:     "sample",
				Cluster:       "test",
				GcpProject:    "test-project",
			},
		},
		{
			"transformer",
			args{
				name:       "transformer",
				podName:    "test-1-1-transformer-12345-deployment",
				namespace:  "sample",
				cluster:    "test",
				gcpProject: "test-project",
			},
			&Container{
				Name:          "transformer",
				PodName:       "test-1-1-transformer-12345-deployment",
				ComponentType: "transformer",
				Namespace:     "sample",
				Cluster:       "test",
				GcpProject:    "test-project",
			},
		},
		{
			"batch_job_driver",
			args{
				name:       "spark-kubernetes-driver",
				podName:    "test-1-driver",
				namespace:  "sample",
				cluster:    "test",
				gcpProject: "test-project",
			},
			&Container{
				Name:          "spark-kubernetes-driver",
				PodName:       "test-1-driver",
				ComponentType: "batch_job_driver",
				Namespace:     "sample",
				Cluster:       "test",
				GcpProject:    "test-project",
			},
		},
		{
			"batch_job_executor",
			args{
				name:       "executor",
				podName:    "test-1-exec",
				namespace:  "sample",
				cluster:    "test",
				gcpProject: "test-project",
			},
			&Container{
				Name:          "executor",
				PodName:       "test-1-exec",
				ComponentType: "batch_job_executor",
				Namespace:     "sample",
				Cluster:       "test",
				GcpProject:    "test-project",
			},
		},
		{
			"image_builder",
			args{
				name:       "pyfunc-image-builder",
				podName:    "sample-test-1-12345",
				namespace:  "sample",
				cluster:    "test",
				gcpProject: "test-project",
			},
			&Container{
				Name:          "pyfunc-image-builder",
				PodName:       "sample-test-1-12345",
				ComponentType: "image_builder",
				Namespace:     "sample",
				Cluster:       "test",
				GcpProject:    "test-project",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewContainer(tt.args.name, tt.args.podName, tt.args.namespace, tt.args.cluster, tt.args.gcpProject); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}
