package storage

import (
	"context"

	"github.com/caraml-dev/merlin/models"
	"gorm.io/gorm"
)

type ModelSchemaStorage interface {
	Save(ctx context.Context, modelSchema *models.ModelSchema) (*models.ModelSchema, error)
	FindAll(ctx context.Context, modelID models.ID) ([]*models.ModelSchema, error)
	FindByID(ctx context.Context, modelSchemaID models.ID, modelID models.ID) (*models.ModelSchema, error)
	Delete(ctx context.Context, modelSchema *models.ModelSchema) error
}

type modelSchemaStorage struct {
	db *gorm.DB
}

func NewModelSchemaStorage(db *gorm.DB) ModelSchemaStorage {
	return &modelSchemaStorage{db: db}
}

func (m *modelSchemaStorage) Save(ctx context.Context, modelSchema *models.ModelSchema) (*models.ModelSchema, error) {
	if err := m.db.Save(modelSchema).Error; err != nil {
		return nil, err
	}
	return modelSchema, nil
}
func (m *modelSchemaStorage) FindAll(ctx context.Context, modelID models.ID) ([]*models.ModelSchema, error) {
	var schemas []*models.ModelSchema
	err := m.db.Where("model_id = ?", modelID).Order("id asc").Find(&schemas).Error
	return schemas, err
}
func (m *modelSchemaStorage) FindByID(ctx context.Context, modelSchemaID models.ID, modelID models.ID) (*models.ModelSchema, error) {
	var modelSchema *models.ModelSchema
	if err := m.db.Where("id = ? AND model_id = ?", modelSchemaID, modelID).First(&modelSchema).Error; err != nil {
		return nil, err
	}
	return modelSchema, nil
}
func (m *modelSchemaStorage) Delete(ctx context.Context, modelSchema *models.ModelSchema) error {
	return m.db.Delete(modelSchema).Error
}
