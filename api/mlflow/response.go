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

package mlflow

// Mlflow API Request / Response
type createExperimentRequest struct {
	Name string `json:"name"`
}

type createExperimentResponse struct {
	ExperimentID string `json:"experiment_id" required:"true"`
}

type createRunRequest struct {
	ExperimentID string `json:"experiment_id"`
	StartTime    int64  `json:"start_time"`
}

type createRunResponse struct {
	Run *Run `json:"run"`
}

type Run struct {
	Info struct {
		RunID          string `json:"run_id"`
		ExperimentID   string `json:"experiment_id"`
		StartTime      string `json:"start_time"`
		EndTime        string `json:"end_time"`
		ArtifactURI    string `json:"artifact_uri"`
		LifecycleStage string `json:"lifecycle_stage"`
		Status         string `json:"status"`
	} `json:"info"`
}

type errorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func (errRaised *errorResponse) Error() string {
	return errRaised.ErrorCode
}
