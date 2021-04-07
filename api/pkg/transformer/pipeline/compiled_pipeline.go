package pipeline

import (
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer"
)

type CompiledPipeline struct {
	compiledJsonpath   map[string]*CompiledJSONPath
	compiledExpression map[string]*vm.Program

	preprocessOps  []Op
	postprocessOps []Op
}

func NewCompiledPipeline(compiledJSONPath map[string]*CompiledJSONPath, compiledExpression map[string]*vm.Program, preprocessOps []Op, postprocessOps []Op) *CompiledPipeline {
	return &CompiledPipeline{
		compiledJsonpath:   compiledJSONPath,
		compiledExpression: compiledExpression,

		preprocessOps:  preprocessOps,
		postprocessOps: postprocessOps,
	}
}

func (p *CompiledPipeline) Preprocess(env *Environment) (transformer.UnmarshalledJSON, error) {
	for _, op := range p.preprocessOps {
		err := op.Execute(env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing preprocessing operation: %T", op)
		}
	}

	// Get output
	return nil, nil
}

func (p *CompiledPipeline) Postprocess(env *Environment) (transformer.UnmarshalledJSON, error) {
	for _, op := range p.postprocessOps {
		err := op.Execute(env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing postprocessing operation: %T", op)
		}
	}

	return nil, nil
}

func (p *CompiledPipeline) CompiledJSONPath(name string) *CompiledJSONPath {
	return p.compiledJsonpath[name]
}

func (p *CompiledPipeline) SetCompiledJSONPath(name string, compiled *CompiledJSONPath) {
	p.compiledJsonpath[name] = compiled
}

func (p *CompiledPipeline) CompiledExpression(name string) *vm.Program {
	return p.compiledExpression[name]
}
