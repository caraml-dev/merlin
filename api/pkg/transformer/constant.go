package transformer

const (
	StandardTransformerConfigEnvName = "STANDARD_TRANSFORMER_CONFIG"
	FeastServingURLsEnvName          = "FEAST_SERVING_URLS"
	FeastFeatureTableSpecsJSON       = "FEAST_FEATURE_TABLE_SPECS_JSONS"
	FeastStorageConfigs              = "FEAST_STORAGE_CONFIGS"
	FeastServingKeepAliveEnabled     = "FEAST_SERVING_KEEP_ALIVE_ENABLED"
	FeastServingKeepAliveTime        = "FEAST_SERVING_KEEP_ALIVE_TIME"
	FeastServingKeepAliveTimeout     = "FEAST_SERVING_KEEP_ALIVE_TIMEOUT"
	DefaultFeastSource               = "DEFAULT_FEAST_SOURCE"

	ModelGRPCKeepAliveEnabled = "MODEL_GRPC_KEEP_ALIVE_ENABLED"
	ModelGRPCKeepAliveTime    = "MODEL_GRPC_KEEP_ALIVE_TIME"
	ModelGRPCKeepAliveTimeout = "MODEL_GRPC_KEEP_ALIVE_TIMEOUT"

	FeastFeatureJSONField = "feast_features"

	JaegerAgentHost    = "JAEGER_AGENT_HOST"
	JaegerAgentPort    = "JAEGER_AGENT_PORT"
	JaegerSamplerType  = "JAEGER_SAMPLER_TYPE"
	JaegerSamplerParam = "JAEGER_SAMPLER_PARAM"
	JaegerDisabled     = "JAEGER_DISABLED"

	KafkaTopic               = "KAFKA_TOPIC"
	KafkaBrokers             = "KAFKA_BROKERS"
	KafkaMaxMessageSizeBytes = "KAFKA_MAX_MESSAGE_SIZE_BYTES"
	KafkaConnectTimeoutMS    = "KAFKA_CONNECT_TIMEOUT_MS"
	KafkaSerialization       = "KAFKA_SERIALIZATION_FORMAT"

	PromNamespace = "merlin_transformer"
)
