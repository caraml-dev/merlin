package pipeline

import (
	ptc "github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"

	"go.uber.org/zap"
)

type CompilerOptions func(compiler *Compiler)

func WithLogger(logger *zap.Logger) CompilerOptions {
	return func(compiler *Compiler) {
		compiler.logger = logger
	}
}

func WithOperationTracingEnabled(enabled bool) CompilerOptions {
	return func(compiler *Compiler) {
		compiler.operationTracingEnabled = enabled
	}
}

func WithProtocol(protocol ptc.Protocol) CompilerOptions {
	return func(compiler *Compiler) {
		if protocol == ptc.UpiV1 {
			compiler.transformerValidationFn = upiTransformerValidation
			compiler.jsonpathSourceType = jsonpath.Proto
		} else {
			compiler.transformerValidationFn = httpTransformerValidation
		}
	}
}

func WithPredictionLogProducer(producer PredictionLogProducer) CompilerOptions {
	return func(compiler *Compiler) {
		compiler.predictionLogProducer = producer
	}
}
