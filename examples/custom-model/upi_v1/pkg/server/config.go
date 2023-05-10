package server

// Config
type Config struct {
	CaraMLConfig CaraMLModelConfig
}

// CaraMLModelConfig
type CaraMLModelConfig struct {
	GRPCPort         int    `envconfig:"CARAML_GRPC_PORT"`
	ModelName        string `envconfig:"CARAML_MODEL_NAME"`
	ModelVersion     string `envconfig:"CARAML_MODEL_VERSION"`
	ModelFullName    string `envconfig:"CARAML_MODEL_FULL_NAME"`
	ArtifactLocation string `envconfig:"CARAML_ARTIFACT_LOCATION"`
}
