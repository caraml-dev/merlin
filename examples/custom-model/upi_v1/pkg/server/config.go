package server

// Config
type Config struct {
	CaraMLConfig CaraMLModelConfig
}

// CaraMLModelConfig contains configuration that supplied by merlin control plan
type CaraMLModelConfig struct {
	// Port that must be listen and opened by gRPC server
	GRPCPort int `envconfig:"CARAML_GRPC_PORT"`
	// Name of the deployed model
	ModelName string `envconfig:"CARAML_MODEL_NAME"`
	// Version of the deployed model
	ModelVersion string `envconfig:"CARAML_MODEL_VERSION"`
	// Full name of model, per merlin 0.27.0 the full name format is {ModelName}-{ModelVersion}
	ModelFullName string `envconfig:"CARAML_MODEL_FULL_NAME"`
	// Path where all the artifacts that uploaded during model upload will be stored
	ArtifactLocation string `envconfig:"CARAML_ARTIFACT_LOCATION"`
}
