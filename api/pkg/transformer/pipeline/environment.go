package pipeline

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr/vm"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
)

type Environment struct {
	symbolRegistry   symbol.Registry
	compiledPipeline *CompiledPipeline
	outputJSON       types.JSONObject
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

func (e *Environment) Preprocess(ctx context.Context, rawRequest types.JSONObject, rawRequestHeaders map[string]string) (types.JSONObject, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "environment.Preprocess")
	defer span.Finish()

	e.symbolRegistry.SetRawRequestJSON(rawRequest)
	e.symbolRegistry.SetRawRequestHeaders(rawRequestHeaders)
	e.SetOutputJSON(rawRequest)

	return e.compiledPipeline.Preprocess(ctx, e)
}

func (e *Environment) Postprocess(ctx context.Context, modelResponse types.JSONObject, modelResponseHeaders map[string]string) (types.JSONObject, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "environment.Postprocess")
	defer span.Finish()

	e.symbolRegistry.SetModelResponseJSON(modelResponse)
	e.symbolRegistry.SetModelResponseHeaders(modelResponseHeaders)
	e.SetOutputJSON(modelResponse)

	return e.compiledPipeline.Postprocess(ctx, e)
}

func (e *Environment) IsPostProcessOpExist() bool {
	return len(e.compiledPipeline.postprocessOps) > 0
}

func (e *Environment) IsPreprocessOpExist() bool {
	return len(e.compiledPipeline.preprocessOps) > 0
}

func (e *Environment) SetOutputJSON(jsonObj types.JSONObject) {
	e.outputJSON = jsonObj
}

func (e *Environment) OutputJSON() types.JSONObject {
	return e.outputJSON
}

func (e *Environment) JSONContainer() types.JSONObjectContainer {
	return e.symbolRegistry.JSONContainer()
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
