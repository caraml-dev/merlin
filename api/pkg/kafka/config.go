package kafka

// Serialization format
type SerializationFormat string

const (
	Protobuf SerializationFormat = "protobuf"
	JSON     SerializationFormat = "json"
)

// Kafka configuration for publishing prediction log
type Config struct {
	Topic               string              `envconfig:"KAFKA_TOPIC"`
	Brokers             string              `envconfig:"KAFKA_BROKERS"`
	CompressionType     string              `envconfig:"KAFKA_COMPRESSION_TYPE" default:"none"`
	MaxMessageSizeBytes int                 `envconfig:"KAFKA_MAX_MESSAGE_SIZE_BYTES" default:"1048588"`
	ConnectTimeoutMS    int                 `envconfig:"KAFKA_CONNECT_TIMEOUT_MS" default:"1000"`
	SerializationFmt    SerializationFormat `envconfig:"KAFKA_SERIALIZATION_FORMAT" default:"protobuf"`
}
