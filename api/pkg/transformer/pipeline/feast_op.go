package pipeline

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type FeastOp struct {
	feastRetriever feast.FeatureRetriever
	logger         *zap.Logger
}

func NewFeastOp(feastClients feast.Clients, feastOptions *feast.Options, entityExtractor *feast.EntityExtractor, featureTableSpecs []*spec.FeatureTable, logger *zap.Logger) Op {
	feastRetriever := feast.NewFeastRetriever(
		feastClients,
		entityExtractor,
		featureTableSpecs,
		feastOptions,
		logger,
	)

	return &FeastOp{
		feastRetriever: feastRetriever,
		logger:         logger,
	}
}

func (op *FeastOp) Execute(context context.Context, env *Environment) error {
	span, ctx := opentracing.StartSpanFromContext(context, "pipeline.FeastOp")
	defer span.Finish()

	featureTables, err := op.feastRetriever.RetrieveFeatureOfEntityInSymbolRegistry(ctx, env.SymbolRegistry())
	if err != nil {
		return err
	}

	for _, featureTable := range featureTables {
		table, err := featureTable.AsTable()
		if err != nil {
			return err
		}
		env.SetSymbol(featureTable.Name, table)
		env.LogOperation("feast", featureTable.Name)
	}

	return nil
}
