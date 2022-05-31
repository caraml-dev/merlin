package pipeline

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	table "github.com/gojek/merlin/pkg/transformer/types/table"
)

type FeastOp struct {
	feastRetriever feast.FeatureRetriever
	logger         *zap.Logger
	*OperationTracing
}

func NewFeastOp(feastClients feast.Clients, feastOptions *feast.Options, entityExtractor *feast.EntityExtractor, featureTableSpecs []*spec.FeatureTable, logger *zap.Logger, tracingEnabled bool) Op {
	feastRetriever := feast.NewFeastRetriever(
		feastClients,
		entityExtractor,
		featureTableSpecs,
		feastOptions,
		logger,
	)

	feastOp := &FeastOp{
		feastRetriever: feastRetriever,
		logger:         logger,
	}

	if tracingEnabled {
		feastOp.OperationTracing = NewOperationTracing(featureTableSpecs, types.FeastOpType)
	}

	return feastOp
}

func (op *FeastOp) Execute(context context.Context, env *Environment) error {
	span, ctx := opentracing.StartSpanFromContext(context, "pipeline.FeastOp")
	defer span.Finish()

	featureTables, err := op.feastRetriever.RetrieveFeatureOfEntityInSymbolRegistry(ctx, env.SymbolRegistry())
	if err != nil {
		return err
	}

	for _, featureTable := range featureTables {
		tbl, err := featureTable.AsTable()
		if err != nil {
			return err
		}

		env.SetSymbol(featureTable.Name, tbl)
		if op.OperationTracing != nil {

			tableJson, err := table.TableToJson(tbl, spec.FromTable_RECORD)
			if err != nil {
				return err
			}
			op.AddInputOutput(nil, map[string]interface{}{featureTable.Name: tableJson})
		}
		env.LogOperation("feast", featureTable.Name)
	}

	return nil
}
