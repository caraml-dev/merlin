package types

type PredictionResult struct {
	Response Payload
	Error    error
	Metadata PredictionMetadata
}

type PredictionMetadata struct {
	ModelName    string
	ModelVersion string
	Project      string
}
