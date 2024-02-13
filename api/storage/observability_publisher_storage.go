package storage

import (
	"context"
	"errors"

	"github.com/caraml-dev/merlin/models"
	"gorm.io/gorm"
)

// ObservabilityPublisherStorage
type ObservabilityPublisherStorage interface {
	GetByModelID(ctx context.Context, modelID models.ID) (*models.ObservabilityPublisher, error)
	Get(ctx context.Context, publisherID models.ID) (*models.ObservabilityPublisher, error)
	Create(ctx context.Context, publisher *models.ObservabilityPublisher) (*models.ObservabilityPublisher, error)
	Update(ctx context.Context, publisher *models.ObservabilityPublisher, increseRevision bool) (*models.ObservabilityPublisher, error)
}

type obsPublisherStorage struct {
	db *gorm.DB
}

func NewObservabilityPublisherStorage(db *gorm.DB) *obsPublisherStorage {
	return &obsPublisherStorage{
		db: db,
	}
}

func (op *obsPublisherStorage) GetByModelID(ctx context.Context, modelID models.ID) (*models.ObservabilityPublisher, error) {
	var publisher models.ObservabilityPublisher
	if err := op.db.Where("version_model_id = ?", modelID).First(&publisher).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &publisher, nil
}

func (op *obsPublisherStorage) Get(ctx context.Context, publisherID models.ID) (*models.ObservabilityPublisher, error) {
	var publisher *models.ObservabilityPublisher
	if err := op.db.Where("id = ?", publisherID).First(&publisher).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return publisher, nil
}
func (op *obsPublisherStorage) Create(ctx context.Context, publisher *models.ObservabilityPublisher) (*models.ObservabilityPublisher, error) {
	publisher.Revision = 1
	if err := op.db.Create(publisher).Error; err != nil {
		return nil, err
	}
	return publisher, nil
}

func (op *obsPublisherStorage) Update(ctx context.Context, publisher *models.ObservabilityPublisher, increseRevision bool) (*models.ObservabilityPublisher, error) {
	currentRevision := publisher.Revision
	if increseRevision {
		publisher.Revision++
	}
	result := op.db.Model(&models.ObservabilityPublisher{}).Where("id = ? AND revision = ?", publisher.ID, currentRevision).Updates(publisher)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return publisher, nil
}
