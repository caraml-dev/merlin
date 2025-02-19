package deployment

import (
	"github.com/caraml-dev/merlin/models"
)

type ConsumerConfig struct {
	Project           string             `yaml:"project"`
	ModelID           string             `yaml:"model_id"`
	ModelVersion      string             `yaml:"model_version"`
	InferenceSchema   *models.SchemaSpec `yaml:"inference_schema"`
	ObservationSinks  []ObservationSink  `yaml:"observation_sinks"`
	ObservationSource *ObserVationSource `yaml:"observation_source"`
}

type ObserVationSource struct {
	Type   SourceType `yaml:"type"`
	Config any        `yaml:"config"`
}

type KafkaSource struct {
	Topic                    string            `yaml:"topic"`
	BootstrapServers         string            `yaml:"bootstrap_servers"`
	GroupID                  string            `yaml:"group_id"`
	BatchSize                int               `yaml:"batch_size"`
	AdditionalConsumerConfig map[string]string `yaml:"additional_consumer_config"`
}

type SinkType string
type SourceType string

const (
	Arize      SinkType = "ARIZE"
	BQ         SinkType = "BIGQUERY"
	MaxCompute SinkType = "MAXCOMPUTE"

	Kafka SourceType = "KAFKA"

	PublisherRevisionAnnotationKey = "publisher-revision"
)

type ObservationSink struct {
	Type   SinkType `yaml:"type"`
	Config any      `yaml:"config"`
}

type ArizeSink struct {
	APIKey   string `yaml:"api_key"`
	SpaceKey string `yaml:"space_key"`
}

type BigQuerySink struct {
	Project string `yaml:"project"`
	Dataset string `yaml:"dataset"`
	TTLDays int    `yaml:"ttl_days"`
}

type MaxComputeSink struct {
	Project         string `yaml:"project"`
	Dataset         string `yaml:"dataset"`
	TTLDays         int    `yaml:"ttl_days"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	AccessUrl       string `yaml:"access_url"`
}
