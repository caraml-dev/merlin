package pipeline

import (
	"context"

	"github.com/caraml-dev/merlin/pkg/transformer/types/table"

	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/expression"
)

type CompiledPipeline struct {
	compiledJsonpath   *jsonpath.Storage
	compiledExpression *expression.Storage
	preloadedTables    map[string]table.Table

	preprocessOps  []Op
	postprocessOps []Op
	tracingEnabled bool
}

func NewCompiledPipeline(
	compiledJSONPath *jsonpath.Storage,
	compiledExpression *expression.Storage,
	preloadedTables map[string]table.Table,
	preprocessOps []Op,
	postprocessOps []Op,
	tracingEnabled bool,
) *CompiledPipeline {
	return &CompiledPipeline{
		compiledJsonpath:   compiledJSONPath,
		compiledExpression: compiledExpression,
		preloadedTables:    preloadedTables,

		preprocessOps:  preprocessOps,
		postprocessOps: postprocessOps,
		tracingEnabled: tracingEnabled,
	}
}

func (p *CompiledPipeline) Preprocess(context context.Context, env *Environment) (types.Payload, error) {
	return p.executePipelineOp(context, types.Preprocess, p.preprocessOps, env)
}

func (p *CompiledPipeline) Postprocess(context context.Context, env *Environment) (types.Payload, error) {
	return p.executePipelineOp(context, types.Postprocess, p.postprocessOps, env)
}

func (p *CompiledPipeline) executePipelineOp(ctx context.Context, pType types.Pipeline, ops []Op, env *Environment) (types.Payload, error) {
	tracingDetails := make([]types.TracingDetail, 0)
	for _, op := range ops {
		err := op.Execute(ctx, env)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing %s operation: %T", pType, op)
		}

		if p.tracingEnabled {
			details, err := op.GetOperationTracingDetail()
			if err != nil {
				return nil, err
			}
			tracingDetails = append(tracingDetails, details...)
		}
	}

	if p.tracingEnabled {
		if pType == types.Preprocess {
			env.SymbolRegistry().SetPreprocessTracingDetail(tracingDetails)
		} else {
			env.SymbolRegistry().SetPostprocessTracingDetail(tracingDetails)
		}
	}

	output := env.Output()
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
