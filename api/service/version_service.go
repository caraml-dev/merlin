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

	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

type VersionsService interface {
	ListVersions(ctx context.Context, modelId models.Id, monitoringConfig config.MonitoringConfig) ([]*models.Version, error)
	Save(ctx context.Context, version *models.Version, monitoringConfig config.MonitoringConfig) (*models.Version, error)
	FindById(ctx context.Context, modelId, versionId models.Id, monitoringConfig config.MonitoringConfig) (*models.Version, error)
}

func NewVersionsService(db *gorm.DB, mlpApiClient mlp.APIClient) VersionsService {
	return &versionsService{db: db, mlpApiClient: mlpApiClient}
}

type versionsService struct {
	db           *gorm.DB
	mlpApiClient mlp.APIClient
}

func (service *versionsService) query() *gorm.DB {
	return service.db.
		Preload("Endpoints", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Environment").
				Preload("Transformer").
				Joins("JOIN models on models.id = version_endpoints.version_model_id").
				Joins("JOIN environments on environments.name = version_endpoints.environment_name").
				Joins("JOIN transformers on transformers.version_endpoint_id = version_endpoints.id").
				Select("version_endpoints.*")
		}).
		Preload("Model").
		Joins("JOIN models on models.id = versions.model_id").
		Select("versions.*")
}

func (service *versionsService) ListVersions(ctx context.Context, modelId models.Id, monitoringConfig config.MonitoringConfig) (versions []*models.Version, err error) {
	err = service.query().
		Where(models.Version{ModelId: modelId}).
		Order("created_at DESC").
		Find(&versions).
		Error

	if len(versions) == 0 {
		return
	}

	project, err := service.mlpApiClient.GetProjectByID(ctx, int32(versions[0].Model.ProjectId))
	if err != nil {
		return nil, err
	}

	for k := range versions {
		versions[k].Model.Project = project
		versions[k].MlflowUrl = project.MlflowRunURL(versions[k].Model.ExperimentId.String(), versions[k].RunId)

		if monitoringConfig.MonitoringEnabled {
			for j := range versions[k].Endpoints {
				versions[k].Endpoints[j].UpdateMonitoringUrl(monitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
					Cluster:      versions[k].Endpoints[j].Environment.Cluster,
					Project:      project.Name,
					Model:        versions[k].Model.Name,
					ModelVersion: versions[k].Model.Name + "-" + versions[k].Id.String(),
				})
			}
		}
	}

	return
}

func (service *versionsService) Save(ctx context.Context, version *models.Version, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	tx := service.db.Begin()
	defer tx.RollbackUnlessCommitted()

	var err error
	if tx.NewRecord(version) {
		err = tx.Create(version).Error
	} else {
		err = tx.Save(version).Error
	}

	if err != nil {
		return nil, err
	} else if err = tx.Commit().Error; err != nil {
		return nil, err
	} else {
		return service.FindById(ctx, version.ModelId, version.Id, monitoringConfig)
	}
}

func (service *versionsService) FindById(ctx context.Context, modelId, versionId models.Id, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	var version models.Version
	if err := service.query().
		Where("models.id = ? AND versions.id = ?", modelId, versionId).
		First(&version).
		Error; err != nil {
		return nil, err
	}

	project, err := service.mlpApiClient.GetProjectByID(ctx, int32(version.Model.ProjectId))
	if err != nil {
		return nil, err
	}

	version.Model.Project = project
	version.MlflowUrl = project.MlflowRunURL(version.Model.ExperimentId.String(), version.RunId)

	if monitoringConfig.MonitoringEnabled {
		for k := range version.Endpoints {
			version.Endpoints[k].UpdateMonitoringUrl(monitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
				Cluster:      version.Endpoints[k].Environment.Cluster,
				Project:      project.Name,
				Model:        version.Model.Name,
				ModelVersion: version.Model.Name + "-" + version.Id.String(),
			})
		}
	}

	return &version, nil
}
