package storage

import (
	"context"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ModelEndpointStorage interface {
	// FindByID find model endpoint given its ID
	FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error)
	// ListModelEndpoints list all model endpoints owned by a model given the model ID
	ListModelEndpoints(ctx context.Context, modelID models.ID) ([]*models.ModelEndpoint, error)
	// ListModelEndpointsInProject list all model endpoints within a project given the project ID
	ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error)
	// Save save newModelEndpoint and its nested version endpoint objects
	Save(ctx context.Context, prevModelEndpoint, newModelEndpoint *models.ModelEndpoint) error
	// Delete delete a model endpoint from the database
	Delete(endpoint *models.ModelEndpoint) error
}

type modelEndpointStorage struct {
	db                     *gorm.DB
	versionEndpointStorage VersionEndpointStorage
}

func NewModelEndpointStorage(db *gorm.DB) ModelEndpointStorage {
	return &modelEndpointStorage{
		db:                     db,
		versionEndpointStorage: NewVersionEndpointStorage(db),
	}
}

// FindByID find model endpoint given its ID
func (m *modelEndpointStorage) FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error) {
	endpoint := &models.ModelEndpoint{}

	if err := m.query().Where("model_endpoints.id = ?", id.String()).Find(&endpoint).Error; err != nil {
		log.Errorf("failed to find model endpoint by id (%s) %v", id, err)
		return nil, errors.Wrapf(err, "failed to find model endpoint by id (%s)", id)
	}

	return endpoint, nil
}

// ListModelEndpoints list all model endpoints owned by a model given the model ID
func (m *modelEndpointStorage) ListModelEndpoints(ctx context.Context, modelID models.ID) (endpoints []*models.ModelEndpoint, err error) {
	err = m.query().Where("model_id = ?", modelID.String()).Find(&endpoints).Error
	return
}

// ListModelEndpointsInProject list all model endpoints within a project given the project ID
func (m *modelEndpointStorage) ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error) {
	// Run the query
	endpoints := []*models.ModelEndpoint{}

	db := m.query().
		Joins("JOIN models on models.id = model_endpoints.model_id").
		Where("models.project_id = ?", projectID)

	// Filter by optional column
	// Environment's region
	if region != "" {
		db = db.Where("environments.region = ?", region)
	}

	if err := db.Find(&endpoints).Error; err != nil {
		log.Errorf("failed to list Model Endpoints for Project ID (%s), %v", projectID, err)
		return nil, errors.Wrapf(err, "failed to list Model Endpoints for Project ID (%s)", projectID)
	}

	return endpoints, nil
}

// Save save newModelEndpoint and its nested version endpoint objects
func (m *modelEndpointStorage) Save(ctx context.Context, prevModelEndpoint, newModelEndpoint *models.ModelEndpoint) error {
	tx := m.db.Begin()
	var err error
	// At the end of the function, commit or rollback transaction, based on err
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Save(newModelEndpoint).Error
	if err != nil {
		return errors.Wrap(err, "Failed to save model endpoint")
	}

	// Update version and version endpoints from previous model endpoint
	if prevModelEndpoint != nil {
		for _, ruleDestination := range prevModelEndpoint.Rule.Destination {
			versionEndpoint, err := m.versionEndpointStorage.Get(ruleDestination.VersionEndpointID)
			if err != nil {
				return err
			}

			versionEndpoint.Status = models.EndpointRunning
			err = tx.Save(versionEndpoint).Error
			if err != nil {
				return errors.Wrap(err, "Failed to save previous model version endpoint")
			}
		}
	}

	// Update version and version endpoints from new model endpoint
	for _, ruleDestination := range newModelEndpoint.Rule.Destination {
		var versionEndpoint *models.VersionEndpoint
		versionEndpoint, err = m.versionEndpointStorage.Get(ruleDestination.VersionEndpointID)
		if err != nil {
			return err
		}

		versionEndpoint.Status = models.EndpointServing
		if newModelEndpoint.Status == models.EndpointTerminated {
			versionEndpoint.Status = models.EndpointRunning
		}

		err = tx.Save(versionEndpoint).Error
		if err != nil {
			return errors.Wrap(err, "Failed to save new model version endpoint")
		}
	}

	return nil
}

func (m *modelEndpointStorage) Delete(endpoint *models.ModelEndpoint) error {
	return m.db.Delete(endpoint).Error
}

func (m *modelEndpointStorage) query() *gorm.DB {
	return m.db.Preload("Environment").
		Preload("Model").
		Joins("JOIN environments on environments.name = model_endpoints.environment_name")
}
