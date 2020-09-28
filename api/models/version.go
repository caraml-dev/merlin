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
	"errors"

	"github.com/jinzhu/gorm"
)

const PropertyPyTorchClassName = "pytorch_class_name"

type Version struct {
	Id          Id                 `json:"id" gorm:"primary_key"`
	ModelId     Id                 `json:"model_id" gorm:"primary_key"`
	Model       *Model             `json:"model" gorm:"foreignkey:ModelId;"`
	RunId       string             `json:"mlflow_run_id" gorm:"column:mlflow_run_id"`
	MlflowUrl   string             `json:"mlflow_url" gorm:"-"`
	ArtifactUri string             `json:"artifact_uri" gorm:"artifact_uri"`
	Endpoints   []*VersionEndpoint `json:"endpoints" gorm:"foreignkey:VersionId,VersionModelId;association_foreignkey:Id,ModelId;"`
	Properties  KV                 `json:"properties" gorm:"properties"`
	CreatedUpdated
}

type VersionPatch struct {
	Properties *KV `json:"properties,omitempty"`
}

type KV map[string]interface{}

func (kv KV) Value() (driver.Value, error) {
	return json.Marshal(kv)
}

func (kv *KV) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &kv)
}

func (v *Version) Patch(patch *VersionPatch) {
	if patch.Properties != nil {
		v.Properties = *patch.Properties
	}
}

func (v *Version) BeforeCreate(scope *gorm.Scope) {
	if v.Id == 0 {
		var maxModelVersionId int

		scope.DB().
			Table("versions").
			Select("COALESCE(MAX(id), 0)").
			Where("model_id = ?", v.ModelId).
			Row().
			Scan(&maxModelVersionId)

		v.Id = Id(maxModelVersionId + 1)
	}
	return
}

// GetEndpointByEnvironmentName return endpoint of this model version which is deployed in environment name
// specified by envName.
// It returns the endpoint if exists (otherwise null) and the boolean `ok`
func (v *Version) GetEndpointByEnvironmentName(envName string) (endpoint *VersionEndpoint, ok bool) {
	for _, ep := range v.Endpoints {
		if envName == ep.EnvironmentName {
			return ep, true
		}
	}
	return nil, false
}

type ModelOption struct {
	PyFuncImageName       string
	PyTorchModelClassName string
}

const DefaultPyTorchClassName = "PyTorchModel"

func NewPyTorchModelOption(version *Version) *ModelOption {
	// Fallback to default if it's empty
	clsName, ok := version.Properties[PropertyPyTorchClassName]
	if !ok {
		return &ModelOption{PyTorchModelClassName: DefaultPyTorchClassName}
	}

	// Fallback to default if it's not castable to string
	clsNameStr, ok := clsName.(string)
	if !ok {
		return &ModelOption{PyTorchModelClassName: DefaultPyTorchClassName}
	}
	return &ModelOption{PyTorchModelClassName: clsNameStr}
}
