package pipeline

import (
	"github.com/antonmedv/expr/vm"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type Environment struct {
	symbolRegistry   SymbolRegistry
	compiledPipeline *CompiledPipeline

	sourceJson map[spec.FromJson_SourceEnum]transformer.UnmarshalledJSON
	outputJson transformer.UnmarshalledJSON
}

const (
	EnvironmentKey = "__transformer_environment__"
)

func NewEnvironment(symbolRegistry SymbolRegistry, compiledPipeline *CompiledPipeline) *Environment {
	env := &Environment{
		symbolRegistry:   symbolRegistry,
		compiledPipeline: compiledPipeline,

		sourceJson: make(map[spec.FromJson_SourceEnum]transformer.UnmarshalledJSON),
	}
	// add reference to the environment in the symbol registry, so that built-in function can access the environment
	symbolRegistry[EnvironmentKey] = env
	return env
}

func (e *Environment) Preprocess(rawRequest transformer.UnmarshalledJSON) (transformer.UnmarshalledJSON, error) {
	e.SetSourceJSON(spec.FromJson_RAW_REQUEST, rawRequest)
	e.SetOutputJSON(rawRequest)

	return e.compiledPipeline.Preprocess(e)
}

func (e *Environment) Postprocess(modelResponse transformer.UnmarshalledJSON) (transformer.UnmarshalledJSON, error) {
	e.SetSourceJSON(spec.FromJson_MODEL_RESPONSE, modelResponse)
	e.SetOutputJSON(modelResponse)

	return e.compiledPipeline.Postprocess(e)
}

func (e *Environment) SetOutputJSON(jsonObj transformer.UnmarshalledJSON) {
	e.outputJson = jsonObj
}

func (e *Environment) OutputJSON(jsonObj transformer.UnmarshalledJSON) {
	e.outputJson = jsonObj
}

func (e *Environment) SourceJSON(source spec.FromJson_SourceEnum) transformer.UnmarshalledJSON {
	return e.sourceJson[source]
}

func (e *Environment) SetSourceJSON(source spec.FromJson_SourceEnum, jsonObj transformer.UnmarshalledJSON) {
	e.sourceJson[source] = jsonObj
}

func (e *Environment) SetSymbol(name string, value interface{}) {
	e.symbolRegistry[name] = value
}

func (e *Environment) SymbolRegistry() SymbolRegistry {
	return e.symbolRegistry
}

func (e *Environment) CompiledJSONPath(name string) *CompiledJSONPath {
	return e.compiledPipeline.CompiledJSONPath(name)
}

func (e *Environment) CompiledExpression(name string) *vm.Program {
	return e.compiledPipeline.compiledExpression[name]
}

func (e *Environment) AddCompiledJsonPath(name string, jsonPath *CompiledJSONPath) {
	e.compiledPipeline.SetCompiledJSONPath(name, jsonPath)
}
