package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/api/resource"
)

// ModelObservability represents the configuration for model observability features.
type ModelObservability struct {
	Enabled                               bool                   `json:"enabled"`
	GroundTruthSource                     *GroundTruthSource     `json:"ground_truth_source"`
	GroundTruthJob                        *GroundTruthJob        `json:"ground_truth_job"`
	PredictionLogIngestionResourceRequest *WorkerResourceRequest `json:"prediction_log_ingestion_resource_request"`
}

func (mo *ModelObservability) IsEnabled() bool {
	if mo == nil {
		return false
	}

	return mo.Enabled
}

// GroundTruthSource represents the source configuration for ground truth data.
type GroundTruthSource struct {
	TableURN             string `json:"table_urn"`
	EventTimestampColumn string `json:"event_timestamp_column"`
	SourceProject        string `json:"source_project"`
}

// GroundTruthJob represents the configuration for a scheduled job.
type GroundTruthJob struct {
	CronSchedule             string  `json:"cron_schedule"`
	CPURequest               string  `json:"cpu_request"`
	CPULimit                 *string `json:"cpu_limit"`
	MemoryRequest            string  `json:"memory_request"`
	MemoryLimit              *string `json:"memory_limit"`
	StartDayOffsetFromNow    int     `json:"start_day_offset_from_now"`
	EndDayOffsetFromNow      int     `json:"end_day_offset_from_now"`
	GracePeriodDay           int     `json:"grace_period_day"`
	ServiceAccountSecretName string  `json:"service_account_secret_name"`
}

// WorkerResourceRequest represents the resource request for a worker (prediction log ingestion kafka consumer) deployment.
type WorkerResourceRequest struct {
	CPURequest    *resource.Quantity `json:"cpu_request"`
	MemoryRequest *resource.Quantity `json:"memory_request"`
	Replica       int32              `json:"replica"`
}

func (mlob ModelObservability) Value() (driver.Value, error) {
	return json.Marshal(mlob)
}

func (mlob *ModelObservability) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &mlob)
}
