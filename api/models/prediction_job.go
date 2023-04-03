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
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/caraml-dev/merlin-pyspark-app/pkg/spec"
)

type PredictionJob struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
	// Keep metadata internal until it's needed to be exposed
	Metadata Metadata `json:"-"`
	// The field name has to be prefixed with the related struct name
	// in order for gorm Preload to work with association_foreignkey
	VersionID       ID           `json:"version_id"`
	VersionModelID  ID           `json:"model_id"`
	ProjectID       ID           `json:"project_id"`
	Environment     *Environment `json:"environment" gorm:"association_foreignkey:Name;association_autoupdate:false;association_autocreate:false;"`
	EnvironmentName string       `json:"environment_name"`
	Config          *Config      `json:"config,omitempty"`
	Status          State        `json:"status"`
	Error           string       `json:"error"`
	CreatedUpdated
}

type Config struct {
	ServiceAccountName string                        `json:"service_account_name"`
	JobConfig          *spec.PredictionJob           `json:"job_config"`
	ImageRef           string                        `json:"image_ref"`
	ResourceRequest    *PredictionJobResourceRequest `json:"resource_request"`
	EnvVars            EnvVars                       `json:"env_vars"`
}

type PredictionJobResourceRequest struct {
	DriverCPURequest    string `json:"driver_cpu_request,omitempty"`
	DriverMemoryRequest string `json:"driver_memory_request,omitempty"`

	ExecutorReplica       int32  `json:"executor_replica,omitempty"`
	ExecutorCPURequest    string `json:"executor_cpu_request,omitempty"`
	ExecutorMemoryRequest string `json:"executor_memory_request,omitempty"`
}

func (r *Config) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *Config) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &r)
}

func (r PredictionJobResourceRequest) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *PredictionJobResourceRequest) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &r)
}

type State string

const (
	JobPending          State = "pending"
	JobRunning          State = "running"
	JobTerminating      State = "terminating"
	JobTerminated       State = "terminated"
	JobCompleted        State = "completed"
	JobFailed           State = "failed"
	JobFailedSubmission State = "failed_submission"
)

func (s State) IsTerminal() bool {
	return s == JobTerminated || s == JobFailedSubmission || s == JobFailed || s == JobCompleted
}

func (s State) IsSuccessful() bool {
	return s == JobCompleted
}
