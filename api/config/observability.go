package config

import (
	"strings"
	"time"
)

// ObservabilityPublisher
type ObservabilityPublisher struct {
	ArizeSink                ArizeSink
	BigQuerySink             BigQuerySink
	KafkaConsumer            KafkaConsumer
	ImageName                string
	DefaultResources         ResourceRequestsLimits
	EnvironmentName          string
	Replicas                 int32
	TargetNamespace          string
	DeploymentTimeout        time.Duration `default:"30m"`
	ServiceAccountSecretName string
}

// KafkaConsumer
type KafkaConsumer struct {
	Brokers                  string `validate:"required"`
	BatchSize                int
	GroupID                  string
	AdditionalConsumerConfig map[string]string
}

// ArizeSink
type ArizeSink struct {
	APIKey              string
	SpaceKey            string
	EnabledModelSerials string
}

func (az ArizeSink) IsEnabled(modelSerial string) bool {
	for _, ems := range strings.Split(az.EnabledModelSerials, ",") {
		if ems == modelSerial {
			return true
		}
	}

	return false
}

// BigQuerySink
type BigQuerySink struct {
	Project string
	Dataset string
	TTLDays int
}
