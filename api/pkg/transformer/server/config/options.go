package config

import (
	"time"

	"github.com/gojek/merlin/pkg/protocol"
)

type Options struct {
	HTTPPort string `envconfig:"CARAML_HTTP_PORT" default:"8081"`
	GRPCPort string `envconfig:"CARAML_GRPC_PORT" default:"9000"`

	ModelFullName   string            `envconfig:"CARAML_MODEL_FULL_NAME" default:"model"`
	ModelPredictURL string            `envconfig:"CARAML_PREDICTOR_HOST" default:"localhost:8080"`
	Protocol        protocol.Protocol `envconfig:"CARAML_PROTOCOL" default:"HTTP_JSON"`

	ServerTimeout time.Duration `envconfig:"SERVER_TIMEOUT" default:"30s"`
	ClientTimeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"1s"`

	ModelTimeout                         time.Duration `envconfig:"MODEL_TIMEOUT" default:"1s"`
	ModelHTTPHystrixCommandName          string        `envconfig:"MODEL_HTTP_HYSTRIX_COMMAND_NAME" default:"http_model_predict"`
	ModelGRPCHystrixCommandName          string        `envconfig:"MODEL_GRPC_HYSTRIX_COMMAND_NAME" default:"grpc_model_predict"`
	ModelHystrixMaxConcurrentRequests    int           `envconfig:"MODEL_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	ModelHystrixRetryCount               int           `envconfig:"MODEL_HYSTRIX_RETRY_COUNT" default:"0"`
	ModelHystrixRetryBackoffInterval     time.Duration `envconfig:"MODEL_HYSTRIX_RETRY_BACKOFF_INTERVAL" default:"5ms"`
	ModelHystrixRetryMaxJitterInterval   time.Duration `envconfig:"MODEL_HYSTRIX_RETRY_MAX_JITTER_INTERVAL" default:"5ms"`
	ModelHystrixErrorPercentageThreshold int           `envconfig:"MODEL_HYSTRIX_ERROR_PERCENTAGE_THRESHOLD" default:"25"`
	ModelHystrixRequestVolumeThreshold   int           `envconfig:"MODEL_HYSTRIX_REQUEST_VOLUME_THRESHOLD" default:"100"`
	ModelHystrixSleepWindowMs            int           `envconfig:"MODEL_HYSTRIX_SLEEP_WINDOW_MS" default:"10"`
}
