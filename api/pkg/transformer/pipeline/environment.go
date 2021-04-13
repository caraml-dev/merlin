package pipeline

import (
	"context"

	"github.com/antonmedv/expr/vm"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
)

type Environment struct {
	symbolRegistry   symbol.Registry
	compiledPipeline *CompiledPipeline
	outputJSON       types.JSONObject
}

func NewEnvironment(compiledPipeline *CompiledPipeline) *Environment {
	sr := symbol.NewRegistryWithCompiledJSONPath(compiledPipeline.compiledJsonpath)
	env := &Environment{
		symbolRegistry:   sr,
		compiledPipeline: compiledPipeline,
	}

	return env
}

func (e *Environment) Preprocess(ctx context.Context, rawRequest types.JSONObject, rawRequestHeaders map[string]string) (types.JSONObject, error) {
	e.symbolRegistry.SetRawRequestJSON(rawRequest)
	e.symbolRegistry.SetRawRequestHeaders(rawRequestHeaders)
	e.SetOutputJSON(rawRequest)

	return e.compiledPipeline.Preprocess(ctx, e)
}

func (e *Environment) Postprocess(ctx context.Context, modelResponse types.JSONObject, modelResponseHeaders map[string]string) (types.JSONObject, error) {
	e.symbolRegistry.SetModelResponseJSON(modelResponse)
	e.symbolRegistry.SetModelResponseHeaders(modelResponseHeaders)
	e.SetOutputJSON(modelResponse)

	return e.compiledPipeline.Postprocess(ctx, e)
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
