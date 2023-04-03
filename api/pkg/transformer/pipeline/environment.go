package pipeline

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr/vm"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
)

type Environment struct {
	symbolRegistry   symbol.Registry
	compiledPipeline *CompiledPipeline
	output           types.Payload
	logger           *zap.Logger
}

func NewEnvironment(compiledPipeline *CompiledPipeline, logger *zap.Logger) *Environment {
	sr := symbol.NewRegistryWithCompiledJSONPath(compiledPipeline.compiledJsonpath)
	env := &Environment{
		symbolRegistry:   sr,
		compiledPipeline: compiledPipeline,
		logger:           logger,
	}

	// attach pre-loaded tables to environment
	for k, v := range compiledPipeline.preloadedTables {
		env.SetSymbol(k, &v)
	}

	return env
}

func (e *Environment) Preprocess(ctx context.Context, rawRequest types.Payload, rawRequestHeaders map[string]string) (types.Payload, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "environment.Preprocess")
	defer span.Finish()

	e.symbolRegistry.SetRawRequest(rawRequest)
	e.symbolRegistry.SetRawRequestHeaders(rawRequestHeaders)
	e.SetOutput(rawRequest)

	response, err := e.compiledPipeline.Preprocess(ctx, e)
	if err == nil {
		e.SetPreprocessResponse(response)
	}
	return response, err
}

func (e *Environment) Postprocess(ctx context.Context, modelResponse types.Payload, modelResponseHeaders map[string]string) (types.Payload, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "environment.Postprocess")
	defer span.Finish()

	e.symbolRegistry.SetModelResponse(modelResponse)
	e.symbolRegistry.SetModelResponseHeaders(modelResponseHeaders)
	e.SetOutput(modelResponse)

	return e.compiledPipeline.Postprocess(ctx, e)
}

func (e *Environment) PublishPredictionLog(ctx context.Context, result *types.PredictionResult) {
	predictionLogOp := e.compiledPipeline.predictionLogOp
	if predictionLogOp == nil {
		return
	}
	go predictionLogOp.ProducePredictionLog(ctx, result, e) //nolint:errcheck
}

func (e *Environment) IsPostProcessOpExist() bool {
	return len(e.compiledPipeline.postprocessOps) > 0
}

func (e *Environment) IsPreprocessOpExist() bool {
	return len(e.compiledPipeline.preprocessOps) > 0
}

func (e *Environment) SetPreprocessResponse(payload types.Payload) {
	e.symbolRegistry.SetPreprocessResponse(payload)
}

func (e *Environment) PreprocessResponse() types.Payload {
	return e.symbolRegistry.PreprocessResponse()
}

func (e *Environment) SetOutput(payload types.Payload) {
	e.output = payload
}

func (e *Environment) Output() types.Payload {
	return e.output
}

func (e *Environment) PayloadContainer() types.PayloadObjectContainer {
	return e.symbolRegistry.PayloadContainer()
}

func (e *Environment) SetSymbol(name string, value interface{}) {
	e.symbolRegistry[name] = value
}

func (e *Environment) SymbolRegistry() symbol.Registry {
	return e.symbolRegistry
}

func (e *Environment) CompiledJSONPath(name string) *jsonpath.Compiled {
	return e.compiledPipeline.CompiledJSONPath(name)
}

func (e *Environment) CompiledExpression(name string) *vm.Program {
	return e.compiledPipeline.compiledExpression.Get(name)
}

func (e *Environment) LogOperation(opName string, variables ...string) {
	if ce := e.logger.Check(zap.DebugLevel, "exec operation"); ce != nil {
		fields := make([]zap.Field, len(variables)+1)
		fields[0] = zap.String("op_name", opName)
		for i, varName := range variables {
			fields[i+1] = zap.String(varName, fmt.Sprintf("%v", e.symbolRegistry[varName]))
		}
		ce.Write(fields...)
	}
}

func (e *Environment) PreprocessTracingDetail() []types.TracingDetail {
	details, err := e.symbolRegistry.PreprocessTracingDetail()
	if err != nil {
		e.logger.Warn("error retrieve preprocess tracing detail: " + err.Error())
	}
	return details
}

func (e *Environment) PostprocessTracingDetail() []types.TracingDetail {
	details, err := e.symbolRegistry.PostprocessTracingDetail()
	if err != nil {
		e.logger.Warn("error retrieve postprocess tracing detail: " + err.Error())
	}
	return details
}
