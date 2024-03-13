package config

// ObservabilityPublisher
type ObservabilityPublisher struct {
	ArizeSink          ArizeSink
	BigQuerySink       BigQuerySink
	BatchSize          int
	KafkaConsumer      KafkaConsumer
	ImageName          string
	DefaultResources   ResourceRequestsLimits
	EnvironmentName    string
	Replicas           int32
	TargetNamespace    string
	ServiceAccountName string
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
	APIKey   string
	SpaceKey string
}

// BigQuerySink
type BigQuerySink struct {
	Project string
	Dataset string
	TTLDays int
}
