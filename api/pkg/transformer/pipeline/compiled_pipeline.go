package pipeline

import (
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/types"
)

type CompiledPipeline struct {
	compiledJsonpath   map[string]*jsonpath.CompiledJSONPath
	compiledExpression map[string]*vm.Program

	preprocessOps  []Op
	postprocessOps []Op
}

func NewCompiledPipeline(compiledJSONPath map[string]*jsonpath.CompiledJSONPath, compiledExpression map[string]*vm.Program, preprocessOps []Op, postprocessOps []Op) *CompiledPipeline {
	return &CompiledPipeline{
		compiledJsonpath:   compiledJSONPath,
		compiledExpression: compiledExpression,

		preprocessOps:  preprocessOps,
		postprocessOps: postprocessOps,
	}
}

func (p *CompiledPipeline) Preprocess(env *Environment) (types.UnmarshalledJSON, error) {
	for _, op := range p.preprocessOps {
		err := op.Execute(env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing preprocessing operation: %T", op)
		}
	}

	// Get output
	return nil, nil
}

func (p *CompiledPipeline) Postprocess(env *Environment) (types.UnmarshalledJSON, error) {
	for _, op := range p.postprocessOps {
		err := op.Execute(env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing postprocessing operation: %T", op)
		}
	}

	return nil, nil
}

func (p *CompiledPipeline) CompiledJSONPath(name string) *jsonpath.CompiledJSONPath {
	return p.compiledJsonpath[name]
}

func (p *CompiledPipeline) SetCompiledJSONPath(name string, compiled *jsonpath.CompiledJSONPath) {
	p.compiledJsonpath[name] = compiled
}

func (p *CompiledPipeline) CompiledExpression(name string) *vm.Program {
	return p.compiledExpression[name]
}
