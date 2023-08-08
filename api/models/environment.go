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

type Environment struct {
	ID                                ID               `json:"id"`
	Name                              string           `json:"name" gorm:"unique;not null"`
	Cluster                           string           `json:"cluster"`
	IsDefault                         *bool            `json:"is_default"`
	Region                            string           `json:"region"`
	GcpProject                        string           `json:"gcp_project"`
	MaxCPU                            string           `json:"max_cpu"`
	MaxMemory                         string           `json:"max_memory"`
	Gpus                              Gpus             `json:"gpus"`
	DefaultResourceRequest            *ResourceRequest `json:"default_resource_request"`
	DefaultTransformerResourceRequest *ResourceRequest `json:"default_transformer_resource_request"`

	IsPredictionJobEnabled              bool                          `json:"is_prediction_job_enabled"`
	IsDefaultPredictionJob              *bool                         `json:"is_default_prediction_job"`
	DefaultPredictionJobResourceRequest *PredictionJobResourceRequest `json:"default_prediction_job_resource_request"`
	CreatedUpdated
}
