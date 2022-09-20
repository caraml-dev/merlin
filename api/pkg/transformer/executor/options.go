package executor

import (
	prt "github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"go.uber.org/zap"
)

type TransformerOptions func(cfg *transformerExecutorConfig)

// WithTraceEnabled function to update/set the trace enabled field in executor config
func WithTraceEnabled(tracingEnabled bool) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.traceEnabled = tracingEnabled
	}
}

// WithFeastOptions function to update/set the feast options field in executor config
func WithFeastOptions(feastOpts feast.Options) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.feastOpts = feastOpts
	}
}

// WithLogger function to update/set logger for executor config
func WithLogger(logger *zap.Logger) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.logger = logger
	}
}

// WithModelPredictor function to update/set model predictor that handle prediction to model
func WithModelPredictor(modelPredictor ModelPredictor) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.modelPredictor = modelPredictor
	}
}

// WithProtocol function to update/set protocol for executor config
func WithProtocol(protocol prt.Protocol) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.protocol = protocol
	}
}
