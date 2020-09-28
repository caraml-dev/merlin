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

package service

import (
	"context"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/jinzhu/gorm"
)

type ModelsService interface {
	ListModels(ctx context.Context, projectId models.Id, name string) ([]*models.Model, error)
	Save(ctx context.Context, model *models.Model) (*models.Model, error)
	Update(ctx context.Context, model *models.Model) (*models.Model, error)
	FindById(ctx context.Context, modelId models.Id) (*models.Model, error)
}

func NewModelsService(db *gorm.DB, mlpApiClient mlp.APIClient) ModelsService {
	return &modelsService{db: db, mlpApiClient: mlpApiClient}
}

type modelsService struct {
	db           *gorm.DB
	mlpApiClient mlp.APIClient
}

func (service *modelsService) query() *gorm.DB {
	return service.db.
		Preload("Endpoints", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Environment")
		}).
		Select("models.*")
}

func (service *modelsService) ListModels(ctx context.Context, projectId models.Id, name string) (models []*models.Model, err error) {
	query := service.query().Where("models.name LIKE ?", name+"%")
	if projectId > 0 {
		query = query.Where("models.project_id = ?", projectId)
	}

	err = query.Find(&models).Error
	if len(models) == 0 {
		return
	}

	project, err := service.mlpApiClient.GetProjectByID(ctx, int32(projectId))
	if err != nil {
		return nil, err
	}

	for k := range models {
		models[k].Project = project
		models[k].MlflowUrl = project.MlflowExperimentURL(models[k].ExperimentId.String())
	}

	return
}

func (service *modelsService) Save(ctx context.Context, model *models.Model) (*models.Model, error) {
	if err := service.db.Create(model).Error; err != nil {
		return nil, err
	}
	return service.FindById(ctx, model.Id)
}

func (service *modelsService) Update(ctx context.Context, model *models.Model) (*models.Model, error) {
	if err := service.db.Save(model).Error; err != nil {
		return nil, err
	}
	return service.FindById(ctx, model.Id)
}

func (service *modelsService) FindById(ctx context.Context, modelId models.Id) (*models.Model, error) {
	var model models.Model
	if err := service.query().
		Where("models.id = ?", modelId).
		First(&model).
		Error; err != nil {

		return nil, err
	}

	project, err := service.mlpApiClient.GetProjectByID(ctx, int32(model.ProjectId))
	if err != nil {
		return nil, err
	}
	model.Project = project
	model.MlflowUrl = project.MlflowExperimentURL(model.ExperimentId.String())

	return &model, nil
}
