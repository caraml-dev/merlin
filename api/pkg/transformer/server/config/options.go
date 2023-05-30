package config

import (
	"time"

	"github.com/caraml-dev/merlin/pkg/protocol"
)

// Option show all configuration for transformer server
type Options struct {
	// Assigned port number for HTTP endpoint
	HTTPPort string `envconfig:"CARAML_HTTP_PORT" default:"8081"`
	// Assigned port number for gRPC service
	GRPCPort string `envconfig:"CARAML_GRPC_PORT" default:"9000"`

	// Full name of the model, usually including model name and model version
	ModelFullName string `envconfig:"CARAML_MODEL_FULL_NAME" default:"model"`
	// URL for model predictor
	ModelPredictURL string `envconfig:"CARAML_PREDICTOR_HOST" default:"localhost:8080"`
	// Choosen protocol for transformer server. There are two protocols "HTTP_JSON" and "UPI_V1", by default "HTTP_JSON" is choosed
	Protocol     protocol.Protocol `envconfig:"CARAML_PROTOCOL" default:"HTTP_JSON"`
	Project      string            `envconfig:"CARAML_PROJECT"`
	ModelVersion string            `envconfig:"CARAML_MODEL_VERSION"`
	ModelName    string            `envconfig:"CARAML_MODEL_NAME"`

	// Timeout for http server
	ServerTimeout time.Duration `envconfig:"SERVER_TIMEOUT" default:"30s"`

	// Timeout duration of model prediction, the default is 1 second
	ModelTimeout time.Duration `envconfig:"MODEL_TIMEOUT" default:"1s"`
	// Name of hystrix command for model prediction that use HTTP_JSON protocol
	ModelHTTPHystrixCommandName string `envconfig:"MODEL_HTTP_HYSTRIX_COMMAND_NAME" default:"http_model_predict"`
	// Name of hystrix command for model prediction that use UPI_V1 protocol
	ModelGRPCHystrixCommandName string `envconfig:"MODEL_GRPC_HYSTRIX_COMMAND_NAME" default:"grpc_model_predict"`
	// Number of model predictor connection
	ModelServerConnCount int `envconfig:"MODEL_SERVER_CONN_COUNT" default:"5"`

	// Maximum concurrent requests when call model predictor
	ModelHystrixMaxConcurrentRequests int `envconfig:"MODEL_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	// Threshold of error percentage, once breach circuit will be open
	ModelHystrixErrorPercentageThreshold int `envconfig:"MODEL_HYSTRIX_ERROR_PERCENTAGE_THRESHOLD" default:"25"`
	// Threshold of number of request to model predictor
	ModelHystrixRequestVolumeThreshold int `envconfig:"MODEL_HYSTRIX_REQUEST_VOLUME_THRESHOLD" default:"100"`
	// Sleep window is duration of rejecting calling model predictor once the circuit is open
	ModelHystrixSleepWindowMs int `envconfig:"MODEL_HYSTRIX_SLEEP_WINDOW_MS" default:"10"`

	// Flag to enable UPI_V1 model predictor keep alive
	ModelGRPCKeepAliveEnabled bool `envconfig:"MODEL_GRPC_KEEP_ALIVE_ENABLED" default:"false"`
	// Duration of interval between keep alive PING
	ModelGRPCKeepAliveTime time.Duration `envconfig:"MODEL_GRPC_KEEP_ALIVE_TIME" default:"60s"`
	// Duration of PING that considered as TIMEOUT
	ModelGRPCKeepAliveTimeout time.Duration `envconfig:"MODEL_GRPC_KEEP_ALIVE_TIMEOUT" default:"5s"`
}
