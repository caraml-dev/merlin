package pipeline

import (
	"context"

	"go.uber.org/zap"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/opentracing/opentracing-go"

	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type FeastOp struct {
	feastRetriever feast.FeatureRetriever
	logger         *zap.Logger
}

func NewFeastOp(feastClient feastSdk.Client, feastOptions *feast.Options, cache cache.Cache, entityExtractor *feast.EntityExtractor, featureTableSpecs []*spec.FeatureTable, logger *zap.Logger) Op {
	feastRetriever := feast.NewFeastRetriever(
		feastClient,
		entityExtractor,
		featureTableSpecs,
		feastOptions,
		cache,
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
