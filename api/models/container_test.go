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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlineInferencePodLabelSelector(t *testing.T) {
	modelName := "my-model"
	versionID := "1"
	result := OnlineInferencePodLabelSelector(modelName, versionID)
	assert.Equal(t, "serving.kubeflow.org/inferenceservice=my-model-1", result)
}

func TestBatchInferencePodLabelSelector(t *testing.T) {
	predictionJobID := "33"
	result := BatchInferencePodLabelSelector(predictionJobID)
	assert.Equal(t, "prediction-job-id=33", result)
}
