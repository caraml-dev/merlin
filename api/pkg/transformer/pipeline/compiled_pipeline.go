package pipeline

import (
	"context"
	"github.com/gojek/merlin/pkg/transformer/types/table"

	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
)

type CompiledPipeline struct {
	compiledJsonpath   *jsonpath.Storage
	compiledExpression *expression.Storage
	preloadedTables    map[string]table.Table

	preprocessOps  []Op
	postprocessOps []Op
}

func NewCompiledPipeline(compiledJSONPath *jsonpath.Storage, compiledExpression *expression.Storage,
	preloadedTables map[string]table.Table, preprocessOps []Op, postprocessOps []Op) *CompiledPipeline {
	return &CompiledPipeline{
		compiledJsonpath:   compiledJSONPath,
		compiledExpression: compiledExpression,
		preloadedTables:    preloadedTables,

		preprocessOps:  preprocessOps,
		postprocessOps: postprocessOps,
	}
}

func (p *CompiledPipeline) Preprocess(context context.Context, env *Environment) (types.JSONObject, error) {
	for _, op := range p.preprocessOps {
		err := op.Execute(context, env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing preprocessing operation: %T", op)
		}
	}

	output := env.OutputJSON()
	if output == nil {
		return nil, errors.New("output json is not computed")
	}
	return output, nil
}

func (p *CompiledPipeline) Postprocess(context context.Context, env *Environment) (types.JSONObject, error) {
	for _, op := range p.postprocessOps {
		err := op.Execute(context, env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing postprocessing operation: %T", op)
		}
	}

	output := env.OutputJSON()
	if output == nil {
		return nil, errors.New("output json is not computed")
	}
	return output, nil
}

func (p *CompiledPipeline) CompiledJSONPath(name string) *jsonpath.Compiled {
	return p.compiledJsonpath.Get(name)
}

func (p *CompiledPipeline) SetCompiledJSONPath(name string, compiled *jsonpath.Compiled) {
	p.compiledJsonpath.Set(name, compiled)
}

func (p *CompiledPipeline) CompiledExpression(name string) *vm.Program {
	return p.compiledExpression.Get(name)
}
