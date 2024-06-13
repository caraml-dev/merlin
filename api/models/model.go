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
	"strconv"
	"time"

	"github.com/caraml-dev/merlin/mlp"
)

const (
	ModelTypePyFunc     = "pyfunc"
	ModelTypeTensorflow = "tensorflow"
	ModelTypeXgboost    = "xgboost"
	ModelTypeSkLearn    = "sklearn"
	ModelTypePyTorch    = "pytorch"
	ModelTypeOnnx       = "onnx"
	ModelTypePyFuncV2   = "pyfunc_v2"
	ModelTypeCustom     = "custom"
)

type ID int

func (id ID) String() string {
	return strconv.Itoa(int(id))
}

func ParseID(id string) (ID, error) {
	parsed, err := strconv.Atoi(id)
	if err != nil {
		return -1, err
	}
	return ID(parsed), nil
}

type CreatedUpdated struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Model struct {
	ID                     ID          `json:"id"`
	Name                   string      `json:"name" validate:"required,min=3,max=25,subdomain_rfc1123"`
	ProjectID              ID          `json:"project_id"`
	Project                mlp.Project `json:"-" gorm:"-"`
	ExperimentID           ID          `json:"mlflow_experiment_id" gorm:"column:mlflow_experiment_id"`
	Type                   string      `json:"type" gorm:"type"`
	MlflowURL              string      `json:"mlflow_url" gorm:"-"`
	ObservabilitySupported bool        `json:"observability_supported" gorm:"column:observability_supported"`

	Endpoints []*ModelEndpoint `json:"endpoints" gorm:"foreignkey:ModelID;"`

	CreatedUpdated
}
