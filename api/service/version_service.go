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
	ListVersions(ctx context.Context, modelID models.ID, monitoringConfig config.MonitoringConfig, query VersionQuery) ([]*models.Version, string, error)
	Save(ctx context.Context, version *models.Version, monitoringConfig config.MonitoringConfig) (*models.Version, error)
	FindByID(ctx context.Context, modelID, versionID models.ID, monitoringConfig config.MonitoringConfig) (*models.Version, error)
}

func NewVersionsService(db *gorm.DB, mlpAPIClient mlp.APIClient) VersionsService {
	return &versionsService{db: db, mlpAPIClient: mlpAPIClient}
}

type VersionQuery struct {
	PaginationQuery
	Search string `schema:"search"`
}

type PaginationQuery struct {
	Limit  int    `schema:"limit"`
	Cursor string `schema:"cursor"`
}

type cursorPagination struct {
	versionID      models.ID
	versionModelID models.ID
}

type versionsService struct {
	db           *gorm.DB
	mlpAPIClient mlp.APIClient
}

func (service *versionsService) query() *gorm.DB {
	return service.db.
		Preload("Endpoints", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Environment").
				Preload("Transformer").
				Joins("JOIN models on models.id = version_endpoints.version_model_id").
				Joins("JOIN environments on environments.name = version_endpoints.environment_name").
				Select("version_endpoints.*")
		}).
		Preload("Model").
		Joins("JOIN models on models.id = versions.model_id").
		Select("versions.*")
}

func (service *versionsService) buildListVersionsQuery(modelID models.ID, query VersionQuery) *gorm.DB {
	dbQuery := service.query().
		Where(models.Version{ModelID: modelID})

	// search only based on mlflow run_id
	if query.Search != "" {
		dbQuery = dbQuery.Where("versions.mlflow_run_id = ?", query.Search)
	}
	dbQuery = dbQuery.Order("created_at DESC")
	return dbQuery
}

func (service *versionsService) ListVersions(ctx context.Context, modelID models.ID, monitoringConfig config.MonitoringConfig, query VersionQuery) (versions []*models.Version, nextCursor string, err error) {
	dbQuery := service.buildListVersionsQuery(modelID, query)

	paginationEnabled := query.Limit > 0
	if paginationEnabled {
		paginateEngine := generatePagination(query.PaginationQuery, []string{"ID"}, descOrder)
		err = paginateEngine.Paginate(dbQuery, &versions).Error
		if paginateEngine.GetNextCursor().After != nil {
			nextCursor = *paginateEngine.GetNextCursor().After
		}
	} else {
		err = dbQuery.Find(&versions).Error
	}

	if err != nil {
		return
	}
	if len(versions) == 0 {
		return
	}

	project, err := service.mlpAPIClient.GetProjectByID(ctx, int32(versions[0].Model.ProjectID))
	if err != nil {
		return nil, "", err
	}

	for k := range versions {
		versions[k].Model.Project = project
		versions[k].MlflowURL = project.MlflowRunURL(versions[k].Model.ExperimentID.String(), versions[k].RunID)

		if monitoringConfig.MonitoringEnabled {
			for j := range versions[k].Endpoints {
				versions[k].Endpoints[j].UpdateMonitoringURL(monitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
					Cluster:      versions[k].Endpoints[j].Environment.Cluster,
					Project:      project.Name,
					Model:        versions[k].Model.Name,
					ModelVersion: versions[k].Model.Name + "-" + versions[k].ID.String(),
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
		return service.FindByID(ctx, version.ModelID, version.ID, monitoringConfig)
	}
}

func (service *versionsService) FindByID(ctx context.Context, modelID, versionID models.ID, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	var version models.Version
	if err := service.query().
		Where("models.id = ? AND versions.id = ?", modelID, versionID).
		First(&version).
		Error; err != nil {
		return nil, err
	}

	project, err := service.mlpAPIClient.GetProjectByID(ctx, int32(version.Model.ProjectID))
	if err != nil {
		return nil, err
	}

	version.Model.Project = project
	version.MlflowURL = project.MlflowRunURL(version.Model.ExperimentID.String(), version.RunID)

	if monitoringConfig.MonitoringEnabled {
		for k := range version.Endpoints {
			version.Endpoints[k].UpdateMonitoringURL(monitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
				Cluster:      version.Endpoints[k].Environment.Cluster,
				Project:      project.Name,
				Model:        version.Model.Name,
				ModelVersion: version.Model.Name + "-" + version.ID.String(),
			})
		}
	}

	return &version, nil
}
