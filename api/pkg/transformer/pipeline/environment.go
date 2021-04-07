package pipeline

import (
	"github.com/antonmedv/expr/vm"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
)

type Environment struct {
	symbolRegistry   SymbolRegistry
	compiledPipeline *CompiledPipeline

	sourceJSONs types.SourceJSON
	outputJson  types.UnmarshalledJSON
}

const (
	EnvironmentKey = "__transformer_environment__"
)

func NewEnvironment(symbolRegistry SymbolRegistry, compiledPipeline *CompiledPipeline) *Environment {
	env := &Environment{
		symbolRegistry:   symbolRegistry,
		compiledPipeline: compiledPipeline,

		sourceJSONs: make(types.SourceJSON, 2),
	}
	// add reference to the environment in the symbol registry, so that built-in function can access the environment
	symbolRegistry[EnvironmentKey] = env
	return env
}

func (e *Environment) Preprocess(rawRequest types.UnmarshalledJSON) (types.UnmarshalledJSON, error) {
	e.SetSourceJSON(spec.FromJson_RAW_REQUEST, rawRequest)
	e.SetOutputJSON(rawRequest)

	return e.compiledPipeline.Preprocess(e)
}

func (e *Environment) Postprocess(modelResponse types.UnmarshalledJSON) (types.UnmarshalledJSON, error) {
	e.SetSourceJSON(spec.FromJson_MODEL_RESPONSE, modelResponse)
	e.SetOutputJSON(modelResponse)

	return e.compiledPipeline.Postprocess(e)
}

func (e *Environment) SetOutputJSON(jsonObj types.UnmarshalledJSON) {
	e.outputJson = jsonObj
}

func (e *Environment) OutputJSON(jsonObj types.UnmarshalledJSON) {
	e.outputJson = jsonObj
}

func (e *Environment) SourceJSON() types.SourceJSON {
	return e.sourceJSONs
}

func (e *Environment) SetSourceJSON(source spec.FromJson_SourceEnum, jsonObj types.UnmarshalledJSON) {
	e.sourceJSONs[source] = jsonObj
}

func (e *Environment) SetSymbol(name string, value interface{}) {
	e.symbolRegistry[name] = value
}

func (e *Environment) SymbolRegistry() SymbolRegistry {
	return e.symbolRegistry
}

func (e *Environment) CompiledJSONPath(name string) *jsonpath.CompiledJSONPath {
	return e.compiledPipeline.CompiledJSONPath(name)
}

func (e *Environment) CompiledExpression(name string) *vm.Program {
	return e.compiledPipeline.compiledExpression[name]
}

func (e *Environment) AddCompiledJsonPath(name string, jsonPath *jsonpath.CompiledJSONPath) {
	e.compiledPipeline.SetCompiledJSONPath(name, jsonPath)
}
