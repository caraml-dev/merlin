package executor

import (
	prt "github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"go.uber.org/zap"
)

type TransformerOptions func(cfg *transformerExecutorConfig)

func WithTraceEnabled(tracingEnabled bool) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.traceEnabled = tracingEnabled
	}
}

func WithFeastOptions(feastOpts feast.Options) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.feastOpts = feastOpts
	}
}

func WithLogger(logger *zap.Logger) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.logger = logger
	}
}

func WithModelPredictor(modelPredictor ModelPredictor) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.modelPredictor = modelPredictor
	}
}

func WithProtocol(protocol prt.Protocol) TransformerOptions {
	return func(cfg *transformerExecutorConfig) {
		cfg.protocol = protocol
	}
}
