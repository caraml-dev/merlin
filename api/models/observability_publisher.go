package models

import (
	"fmt"
)

// PublisherStatus
type PublisherStatus string

const (
	Pending    PublisherStatus = "pending"
	Running    PublisherStatus = "running"
	Failed     PublisherStatus = "failed"
	Terminated PublisherStatus = "terminated"
)

// ObservabilityPublisher
type ObservabilityPublisher struct {
	ID              ID              `gorm:"id"`
	VersionModelID  ID              `gorm:"version_model_id"`
	VersionID       ID              `gorm:"version_id"`
	Revision        int             `gorm:"revision"`
	Status          PublisherStatus `gorm:"status"`
	ModelSchemaSpec *SchemaSpec     `gorm:"model_schema_spec"`
	CreatedUpdated
}

type ActionType string

const (
	DeployPublisher   ActionType = "deploy"
	UndeployPublisher ActionType = "delete"
)

type WorkerData struct {
	Project         string
	ModelSchemaSpec *SchemaSpec
	Metadata        Metadata
	ModelName       string
	ModelVersion    string
	Revision        int
	TopicSource     string
}

func NewWorkerData(modelVersion *Version, model *Model, observabilityPublisher *ObservabilityPublisher) *WorkerData {
	return &WorkerData{
		ModelName:       model.Name,
		Project:         model.Project.Name,
		ModelSchemaSpec: observabilityPublisher.ModelSchemaSpec,
		Metadata: Metadata{
			App:       fmt.Sprintf("%s-observability-publisher", model.Name),
			Component: "worker",
			Stream:    model.Project.Stream,
			Team:      model.Project.Team,
			Labels:    model.Project.Labels,
		},
		ModelVersion: modelVersion.ID.String(),
		Revision:     observabilityPublisher.Revision,
		TopicSource:  getPredictionLogTopicForVersion(model.Project.Name, model.Name, modelVersion.ID.String()),
	}
}

type ObservabilityPublisherJob struct {
	ActionType ActionType
	Publisher  *ObservabilityPublisher
	WorkerData *WorkerData
}
