package pipeline

import (
	"context"

	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
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

func (op *FeastOp) Execute(ctx context.Context, env *Environment) error {
	ctx, span := tracer.Start(ctx, "pipeline.FeastOp")
	defer span.End()

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
			if err := op.AddInputOutput(nil, map[string]interface{}{featureTable.Name: tbl}); err != nil {
				return err
			}
		}
		env.LogOperation("feast", featureTable.Name)
	}

	return nil
}
