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
	"fmt"

	"github.com/caraml-dev/merlin/models"
	"gorm.io/gorm"
)

type PredictionJobStorage interface {
	// Get get prediction job with given ID
	Get(ID models.ID) (*models.PredictionJob, error)
	// List lists all prediction job matching the given query
	List(query *models.PredictionJob, search string, offset *int, limit *int) (endpoints []*models.PredictionJob, err error)
	// Count returns the count of rows in the given query
	Count(query *models.PredictionJob, search string) int
	// Save save the prediction job to underlying storage
	Save(predictionJob *models.PredictionJob) error
	// GetFirstSuccessModelVersionPerModel get first model version resulting in a successful batch prediction job
	GetFirstSuccessModelVersionPerModel() (map[models.ID]models.ID, error)
	Delete(*models.PredictionJob) error
}

type predictionJobStorage struct {
	db *gorm.DB
}

func NewPredictionJobStorage(db *gorm.DB) PredictionJobStorage {
	return &predictionJobStorage{db: db}
}

// Get get prediction job with given ID
func (p *predictionJobStorage) Get(id models.ID) (*models.PredictionJob, error) {
	var predictionJob models.PredictionJob
	if err := p.query().Where("id = ?", id).First(&predictionJob).Error; err != nil {
		return nil, err
	}
	return &predictionJob, nil
}

// List list all prediction job matching the given query
func (p *predictionJobStorage) List(query *models.PredictionJob, search string, offset *int, limit *int) (predictionJobs []*models.PredictionJob, err error) {
	q := p.query().Select("id, name, version_id, version_model_id, project_id, environment_name, status, error, created_at, updated_at").
		Where(query).
		Order("updated_at desc") // preserve order in case offset or limit are being set

	// Do a partial match on the name
	if search != "" {
		q = q.Where(fmt.Sprintf("name ILIKE '%%%s%%'", search))
	}

	if offset != nil {
		q = q.Offset(*offset)
	}
	if limit != nil {
		q = q.Limit(*limit)
	}
	err = q.Find(&predictionJobs).Error
	return
}

func (p *predictionJobStorage) Count(query *models.PredictionJob, search string) int {
	var count int64
	var jobs []*models.PredictionJob

	q := p.query().Model(&jobs).Where(query)

	// Do a partial match on the name
	if search != "" {
		q = q.Where(fmt.Sprintf("name ILIKE '%%%s%%'", search))
	}

	_ = q.Count(&count)
	return int(count)
}

// Save save the prediction job to underlying storage
func (p *predictionJobStorage) Save(predictionJob *models.PredictionJob) error {
	return p.db.Save(predictionJob).Error
}

func (p *predictionJobStorage) Delete(predictionJob *models.PredictionJob) error {
	return p.db.Delete(predictionJob).Error
}

func (p *predictionJobStorage) GetFirstSuccessModelVersionPerModel() (map[models.ID]models.ID, error) {
	rows, err := p.db.Table("prediction_jobs").
		Select("version_model_id , min(version_id)").
		Where("status = 'completed' ").
		Group("version_model_id").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint: errcheck

	resultMap := make(map[models.ID]models.ID)
	for rows.Next() {
		var modelID models.ID
		var versionID models.ID
		if err := rows.Scan(&modelID, &versionID); err != nil {
			return nil, err
		}
		resultMap[modelID] = versionID
	}
	return resultMap, nil
}

func (p *predictionJobStorage) query() *gorm.DB {
	return p.db.
		Preload("Environment")
}
