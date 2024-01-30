package service

import (
	"context"
	"errors"

	"github.com/caraml-dev/merlin/models"
	mErrors "github.com/caraml-dev/merlin/pkg/errors"
	"github.com/caraml-dev/merlin/storage"
	"gorm.io/gorm"
)

type ModelSchemaService interface {
	List(ctx context.Context, modelID models.ID) ([]*models.ModelSchema, error)
	Save(ctx context.Context, modelSchema *models.ModelSchema) (*models.ModelSchema, error)
	Delete(ctx context.Context, modelSchema *models.ModelSchema) error
	FindByID(ctx context.Context, modelSchemaID models.ID, modelID models.ID) (*models.ModelSchema, error)
}

type modelSchemaService struct {
	modelSchemaStorage storage.ModelSchemaStorage
}

func NewModelSchemaService(storage storage.ModelSchemaStorage) ModelSchemaService {
	return &modelSchemaService{
		modelSchemaStorage: storage,
	}
}

func (m *modelSchemaService) List(ctx context.Context, modelID models.ID) ([]*models.ModelSchema, error) {
	schemas, err := m.modelSchemaStorage.FindAll(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mErrors.NewNotfoundErrorf("model schema with model id: %d are not found", modelID)
		}
		return nil, err
	}
	return schemas, nil
}
func (m *modelSchemaService) Save(ctx context.Context, modelSchema *models.ModelSchema) (*models.ModelSchema, error) {
	return m.modelSchemaStorage.Save(ctx, modelSchema)
}
func (m *modelSchemaService) Delete(ctx context.Context, modelSchema *models.ModelSchema) error {
	return m.modelSchemaStorage.Delete(ctx, modelSchema)
}
func (m *modelSchemaService) FindByID(ctx context.Context, modelSchemaID models.ID, modelID models.ID) (*models.ModelSchema, error) {
	schema, err := m.modelSchemaStorage.FindByID(ctx, modelSchemaID, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mErrors.NewNotfoundErrorf("model schema with id: %d are not found", modelSchemaID)
		}
		return nil, err
	}
	return schema, nil
}
