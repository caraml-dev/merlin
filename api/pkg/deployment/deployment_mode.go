package deployment

// Mode mode of the deployment
type Mode string

const (
	// ServerlessDeploymentMode uses knative service as deployment method
	ServerlessDeploymentMode Mode = "serverless"
	// RawDeploymentMode uses k8s deployment as deployment method
	RawDeploymentMode Mode = "raw_deployment"
	// EmptyMode
	EmptyDeploymentMode Mode = ""
)
