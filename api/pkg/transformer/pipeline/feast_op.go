package pipeline

import (
	"context"

	"go.uber.org/zap"

	feastSdk "github.com/feast-dev/feast/sdk/go"

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

func (op *FeastOp) Execute(context context.Context, environment *Environment) error {
	featureTables, err := op.feastRetriever.RetrieveFeatureOfEntityInSymbolRegistry(context, environment.SymbolRegistry())
	if err != nil {
		return err
	}

	for _, featureTable := range featureTables {
		environment.SetSymbol(featureTable.Name, featureTable.AsTable())
	}

	return nil
}
