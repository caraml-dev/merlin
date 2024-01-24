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

	"gorm.io/gorm"
)

type Version struct {
	ID              ID                 `json:"id" gorm:"primary_key"`
	ModelID         ID                 `json:"model_id" gorm:"primary_key"`
	Model           *Model             `json:"model" gorm:"foreignkey:ModelID;"`
	RunID           string             `json:"mlflow_run_id" gorm:"column:mlflow_run_id"`
	MlflowURL       string             `json:"mlflow_url" gorm:"-"`
	ArtifactURI     string             `json:"artifact_uri" gorm:"artifact_uri"`
	Endpoints       []*VersionEndpoint `json:"endpoints" gorm:"foreignkey:VersionID,VersionModelID;references:ID,ModelID;"`
	ModelSchemaID   *ID                `json:"-"`
	ModelSchema     *ModelSchema       `json:"model_schema" gorm:"foreignkey:ModelSchemaID;"`
	Properties      KV                 `json:"properties" gorm:"properties"`
	Labels          KV                 `json:"labels" gorm:"labels"`
	PythonVersion   string             `json:"python_version" gorm:"python_version"`
	CustomPredictor *CustomPredictor   `json:"custom_predictor"`
	CreatedUpdated
}

type VersionPost struct {
	Labels        KV           `json:"labels" gorm:"labels"`
	PythonVersion string       `json:"python_version" gorm:"python_version"`
	ModelSchema   *ModelSchema `json:"model_schema"`
}

type VersionPatch struct {
	Properties      *KV              `json:"properties,omitempty"`
	CustomPredictor *CustomPredictor `json:"custom_predictor,omitempty"`
	ModelSchema     *ModelSchema     `json:"model_schema"`
}

type CustomPredictor struct {
	Image   string `json:"image"`
	Command string `json:"command"`
	Args    string `json:"args"`
}

func (cp CustomPredictor) IsValid() error {
	if cp.Image == "" {
		return errors.New("custom predictor image must be set")
	}
	return nil
}

func (cp CustomPredictor) Value() (driver.Value, error) {
	return json.Marshal(cp)
}

func (cp *CustomPredictor) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &cp)
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

func (v *Version) Patch(patch *VersionPatch) error {
	if patch.Properties != nil {
		v.Properties = *patch.Properties
	}
	if patch.CustomPredictor != nil && v.Model.Type == ModelTypeCustom {
		if err := patch.CustomPredictor.IsValid(); err != nil {
			return err
		}
		v.CustomPredictor = patch.CustomPredictor
	}
	if patch.ModelSchema != nil {
		v.ModelSchema = patch.ModelSchema
	}

	return nil
}

func (v *Version) BeforeCreate(db *gorm.DB) error {
	if v.ID == 0 {
		var maxModelVersionID int

		db.
			Table("versions").
			Select("COALESCE(MAX(id), 0)").
			Where("model_id = ?", v.ModelID).
			Row().
			Scan(&maxModelVersionID) //nolint:errcheck

		v.ID = ID(maxModelVersionID + 1)
	}
	return nil
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
	PyFuncImageName string
	CustomPredictor *CustomPredictor
}

func NewCustomModelOption(version *Version) *ModelOption {
	return &ModelOption{CustomPredictor: version.CustomPredictor}
}
