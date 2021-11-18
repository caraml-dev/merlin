package transformer

const (
	StandardTransformerConfigEnvName = "STANDARD_TRANSFORMER_CONFIG"
	FeastServingURLsEnvName          = "FEAST_SERVING_URLS"
	FeastFeatureTableSpecsJSON       = "FEAST_FEATURE_TABLE_SPECS_JSONS"
	FeastStorageConfigs              = "FEAST_STORAGE_CONFIGS"
	DefaultFeastSource               = "DEFAULT_FEAST_SOURCE"

	FeastFeatureJSONField = "feast_features"

	JaegerAgentHost    = "JAEGER_AGENT_HOST"
	JaegerAgentPort    = "JAEGER_AGENT_PORT"
	JaegerSamplerType  = "JAEGER_SAMPLER_TYPE"
	JaegerSamplerParam = "JAEGER_SAMPLER_PARAM"
	JaegerDisabled     = "JAEGER_DISABLED"

	PromNamespace = "merlin_transformer"
)
