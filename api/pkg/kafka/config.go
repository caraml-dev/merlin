package kafka

// Serialization format
type SerializationFormat string

const (
	Protobuf SerializationFormat = "protobuf"
	JSON     SerializationFormat = "json"
)

// Kafka configuration
type Config struct {
	Topic               string              `envconfig:"KAFKA_TOPIC"`
	Brokers             string              `envconfig:"KAFKA_BROKERS"`
	CompressionType     string              `envconfig:"KAFKA_COMPRESSION_TYPE" default:"none"`
	MaxMessageSizeBytes int                 `envconfig:"KAFKA_MAX_MESSAGE_SIZE_BYTES" default:"1048588"`
	SerializationFmt    SerializationFormat `envconfig:"KAFKA_SERIALIZATION_FORMAT" default:"protobuf"`
}
