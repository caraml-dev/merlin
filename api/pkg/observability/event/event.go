package event

import (
	"context"
	"fmt"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/storage"
)

const (
	ObservabilityPublisherDeployment = "observability_publisher_deployment"
	dataArgKey                       = "data"
)

type EventProducer interface {
	ModelEndpointChangeEvent(modelEndpoint *models.ModelEndpoint, model *models.Model) error
	VersionEndpointChangeEvent(versionEndpoint *models.VersionEndpoint, model *models.Model) error
}

type eventProducer struct {
	jobProducer                   queue.Producer
	observabilityPublisherStorage storage.ObservabilityPublisherStorage
	versionStorage                storage.VersionStorage
}

func NewEventProducer(jobProducer queue.Producer, observabilityPublisherStorage storage.ObservabilityPublisherStorage, versionStorage storage.VersionStorage) *eventProducer {
	return &eventProducer{
		jobProducer:                   jobProducer,
		observabilityPublisherStorage: observabilityPublisherStorage,
		versionStorage:                versionStorage,
	}
}

func (e *eventProducer) ModelEndpointChangeEvent(modelEndpoint *models.ModelEndpoint, model *models.Model) error {
	if !model.ObservabilitySupported {
		return nil
	}

	ctx := context.Background()
	publisher, err := e.observabilityPublisherStorage.GetByModelID(ctx, model.ID)
	if err != nil {
		return err
	}

	// undeploy if
	// model endpoint is nil or
	// version endpoint observability is false
	if isUndeployAction(modelEndpoint) {
		if publisher == nil || publisher.Status == models.Terminated {
			return nil
		}

		var versionID models.ID
		if modelEndpoint == nil {
			versionID = publisher.VersionID
		} else {
			vEndpoint := modelEndpoint.GetVersionEndpoint()
			versionID = vEndpoint.VersionID
		}

		version, err := e.findVersionWithModelSchema(ctx, versionID, model.ID)
		if err != nil {
			return err
		}

		return e.enqueueJob(version, publisher, models.UndeployPublisher)
	}

	versionEndpoint := modelEndpoint.GetVersionEndpoint()
	version, err := e.findVersionWithModelSchema(ctx, versionEndpoint.VersionID, model.ID)
	if err != nil {
		return err
	}

	if publisher == nil {
		publisher = &models.ObservabilityPublisher{
			VersionModelID: modelEndpoint.ModelID,
			Revision:       1,
		}
	}

	publisher.VersionID = versionEndpoint.VersionID
	publisher.ModelSchemaSpec = version.ModelSchema.Spec

	return e.enqueueJob(version, publisher, models.DeployPublisher)
}

func (e *eventProducer) VersionEndpointChangeEvent(versionEndpoint *models.VersionEndpoint, model *models.Model) error {
	if !model.ObservabilitySupported {
		return nil
	}

	// check if version endpoint is used by the model endpoint
	// if version endpoint is not serving skipping deployment
	if versionEndpoint.Status != models.EndpointServing {
		return nil
	}

	ctx := context.Background()
	publisher, err := e.observabilityPublisherStorage.GetByModelID(ctx, model.ID)
	if err != nil {
		return err
	}

	// Undeploy if version endpoint observability is false
	if !versionEndpoint.EnableModelObservability {
		if publisher == nil || publisher.Status == models.Terminated {
			return nil
		}
		version, err := e.findVersionWithModelSchema(ctx, versionEndpoint.VersionID, model.ID)
		if err != nil {
			return err
		}
		return e.enqueueJob(version, publisher, models.UndeployPublisher)
	}

	version, err := e.findVersionWithModelSchema(ctx, versionEndpoint.VersionID, model.ID)
	if err != nil {
		return err
	}

	if publisher == nil {
		publisher = &models.ObservabilityPublisher{
			VersionModelID: versionEndpoint.VersionModelID,
			Revision:       1,
		}
	}

	publisher.VersionID = versionEndpoint.VersionID
	publisher.ModelSchemaSpec = version.ModelSchema.Spec
	return e.enqueueJob(version, publisher, models.DeployPublisher)
}

func isUndeployAction(modelEndpoint *models.ModelEndpoint) bool {
	if modelEndpoint == nil {
		return true
	}
	if len(modelEndpoint.Rule.Destination) == 0 {
		return false
	}
	destination := modelEndpoint.Rule.Destination[0]
	return !destination.VersionEndpoint.EnableModelObservability
}

func (e *eventProducer) findVersionWithModelSchema(ctx context.Context, versionID models.ID, modelID models.ID) (*models.Version, error) {
	version, err := e.versionStorage.FindByID(ctx, versionID, modelID)
	if err != nil {
		return nil, err
	}
	if version.ModelSchema == nil {
		return nil, fmt.Errorf("versionID: %d in modelID: %d doesn't have model schema", versionID, modelID)
	}
	return version, nil
}

func (e *eventProducer) enqueueJob(version *models.Version, publisher *models.ObservabilityPublisher, actionType models.ActionType) error {
	publisher.Status = models.Pending
	if version.ModelSchema != nil {
		publisher.ModelSchemaSpec = version.ModelSchema.Spec
	}
	ctx := context.Background()
	if publisher.ID > 0 {
		increaseRevision := actionType == models.DeployPublisher
		updatedPublisher, err := e.observabilityPublisherStorage.Update(ctx, publisher, increaseRevision)
		if err != nil {
			return err
		}
		publisher = updatedPublisher
	} else {
		updatedPublisher, err := e.observabilityPublisherStorage.Create(ctx, publisher)
		if err != nil {
			return err
		}
		publisher = updatedPublisher
	}

	return e.jobProducer.EnqueueJob(&queue.Job{
		Name: ObservabilityPublisherDeployment,
		Arguments: queue.Arguments{
			dataArgKey: models.ObservabilityPublisherJob{
				ActionType: actionType,
				Publisher:  publisher,
				WorkerData: models.NewWorkerData(version, publisher),
			},
		},
	})
}
