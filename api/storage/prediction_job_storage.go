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

package storage

import (
	"github.com/gojek/merlin/models"
	"github.com/jinzhu/gorm"
)

type PredictionJobStorage interface {
	// Get get prediction job with given ID
	Get(Id models.Id) (*models.PredictionJob, error)
	// List list all prediction job matching the given query
	List(query *models.PredictionJob) (endpoints []*models.PredictionJob, err error)
	// Save save the prediction job to underlying storage
	Save(predictionJob *models.PredictionJob) error
	// GetFirstSuccessModelVersionPerModel get first model version resulting in a successful batch prediction job
	// GetFirstSuccessModelVersionPerModel get first model version resulting in a successful batch prediction job
	GetFirstSuccessModelVersionPerModel() (map[models.Id]models.Id, error)
}

type predictionJobStorage struct {
	db *gorm.DB
}

func NewPredictionJobStorage(db *gorm.DB) PredictionJobStorage {
	return &predictionJobStorage{db: db}
}

// Get get prediction job with given ID
func (p *predictionJobStorage) Get(id models.Id) (*models.PredictionJob, error) {
	var predictionJob models.PredictionJob
	if err := p.query().Where("id = ?", id).First(&predictionJob).Error; err != nil {
		return nil, err
	}
	return &predictionJob, nil
}

// List list all prediction job matching the given query
func (p *predictionJobStorage) List(query *models.PredictionJob) (predictionJobs []*models.PredictionJob, err error) {
	err = p.query().Select("id, name, version_id, version_model_id, project_id, environment_name, status, error, created_at, updated_at").
		Where(query).Find(&predictionJobs).Error
	return
}

// Save save the prediction job to underlying storage
func (p *predictionJobStorage) Save(predictionJob *models.PredictionJob) error {
	return p.db.Save(predictionJob).Error
}

func (p *predictionJobStorage) GetFirstSuccessModelVersionPerModel() (map[models.Id]models.Id, error) {
	rows, err := p.db.Table("prediction_jobs").
		Select("version_model_id , min(version_id)").
		Where("status = 'completed' ").
		Group("version_model_id").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultMap := make(map[models.Id]models.Id)
	for rows.Next() {
		var modelId models.Id
		var versionId models.Id
		if err := rows.Scan(&modelId, &versionId); err != nil {
			return nil, err
		}
		resultMap[modelId] = versionId
	}
	return resultMap, nil
}

func (p *predictionJobStorage) query() *gorm.DB {
	return p.db.
		Preload("Environment")
}
