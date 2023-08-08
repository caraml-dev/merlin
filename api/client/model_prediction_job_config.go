/*
 * Merlin
 *
 * API Guide for accessing Merlin's model management, deployment, and serving functionalities
 *
 * API version: 0.14.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package client

type PredictionJobConfig struct {
	Version        string                             `json:"version,omitempty"`
	Kind           string                             `json:"kind,omitempty"`
	Name           string                             `json:"name,omitempty"`
	BigquerySource *PredictionJobConfigBigquerySource `json:"bigquery_source,omitempty"`
	GcsSource      *PredictionJobConfigGcsSource      `json:"gcs_source,omitempty"`
	Model          *PredictionJobConfigModel          `json:"model,omitempty"`
	BigquerySink   *PredictionJobConfigBigquerySink   `json:"bigquery_sink,omitempty"`
	GcsSink        *PredictionJobConfigGcsSink        `json:"gcs_sink,omitempty"`
}
