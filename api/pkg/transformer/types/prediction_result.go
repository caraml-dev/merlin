package types

// PredictionResult holds information about response of gRPC request including its metadata
type PredictionResult struct {
	Response Payload
	Error    error
	Metadata PredictionMetadata
}

// PredictionMetadata contains information about model and project that produce the prediction
type PredictionMetadata struct {
	ModelName    string
	ModelVersion string
	Project      string
}
