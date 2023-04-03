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

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/jinzhu/gorm"
)

type ModelsService interface {
	ListModels(ctx context.Context, projectID models.ID, name string) ([]*models.Model, error)
	Save(ctx context.Context, model *models.Model) (*models.Model, error)
	Update(ctx context.Context, model *models.Model) (*models.Model, error)
	FindByID(ctx context.Context, modelID models.ID) (*models.Model, error)
}

func NewModelsService(db *gorm.DB, mlpAPIClient mlp.APIClient) ModelsService {
	return &modelsService{db: db, mlpAPIClient: mlpAPIClient}
}

type modelsService struct {
	db           *gorm.DB
	mlpAPIClient mlp.APIClient
}

func (service *modelsService) query() *gorm.DB {
	return service.db.
		Preload("Endpoints", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Environment")
		}).
		Select("models.*")
}

func (service *modelsService) ListModels(ctx context.Context, projectID models.ID, name string) (models []*models.Model, err error) {
	query := service.query().Where("models.name LIKE ?", name+"%")
	if projectID > 0 {
		query = query.Where("models.project_id = ?", projectID)
	}

	err = query.Find(&models).Error
	if len(models) == 0 {
		return
	}

	project, err := service.mlpAPIClient.GetProjectByID(ctx, int32(projectID))
	if err != nil {
		return nil, err
	}

	for k := range models {
		models[k].Project = project
		models[k].MlflowURL = project.MlflowExperimentURL(models[k].ExperimentID.String())
	}

	return
}

func (service *modelsService) Save(ctx context.Context, model *models.Model) (*models.Model, error) {
	if err := service.db.Create(model).Error; err != nil {
		return nil, err
	}
	return service.FindByID(ctx, model.ID)
}

func (service *modelsService) Update(ctx context.Context, model *models.Model) (*models.Model, error) {
	if err := service.db.Save(model).Error; err != nil {
		return nil, err
	}
	return service.FindByID(ctx, model.ID)
}

func (service *modelsService) FindByID(ctx context.Context, modelID models.ID) (*models.Model, error) {
	var model models.Model
	if err := service.query().
		Where("models.id = ?", modelID).
		First(&model).
		Error; err != nil {

		return nil, err
	}

	project, err := service.mlpAPIClient.GetProjectByID(ctx, int32(model.ProjectID))
	if err != nil {
		return nil, err
	}
	model.Project = project
	model.MlflowURL = project.MlflowExperimentURL(model.ExperimentID.String())

	return &model, nil
}
