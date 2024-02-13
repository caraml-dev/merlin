package storage

import (
	"context"

	"github.com/caraml-dev/merlin/models"
	"gorm.io/gorm"
)

type VersionStorage interface {
	FindByID(ctx context.Context, versionID models.ID, modelID models.ID) (*models.Version, error)
}

type versionStorage struct {
	db *gorm.DB
}

func NewVersionStorage(db *gorm.DB) VersionStorage {
	return &versionStorage{db: db}
}

func (v *versionStorage) FindByID(ctx context.Context, versionID models.ID, modelID models.ID) (*models.Version, error) {
	var version models.Version
	if err := v.db.Preload("ModelSchema").Where("versions.id = ? AND versions.model_id = ?", versionID, modelID).First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &version, nil
}
