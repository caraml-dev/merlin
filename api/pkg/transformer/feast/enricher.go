package feast

import (
	"context"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
)

// Enricher wraps feast serving client to retrieve features.
type Enricher struct {
	featureRetriever FeatureRetriever
	logger           *zap.Logger
}

// NewEnricher initializes a new Enricher.
func NewEnricher(featureRetriever FeatureRetriever, logger *zap.Logger) (*Enricher, error) {
	return &Enricher{
		featureRetriever: featureRetriever,
		logger:           logger,
	}, nil
}

// Enrich retrieves the Feast features values and add them into the request.
func (t *Enricher) Enrich(ctx context.Context, request []byte, _ map[string]string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.Enrich")
	defer span.Finish()

	var requestJson types.JSONObject
	err := json.Unmarshal(request, &requestJson)
	if err != nil {
		return nil, err
	}

	feastFeatures, err := t.featureRetriever.RetrieveFeatureOfEntityInRequest(ctx, requestJson)
	if err != nil {
		return nil, err
	}

	out, err := enrichRequest(ctx, request, feastFeatures)
	if err != nil {
		return nil, err
	}

	return out, err
}

func enrichRequest(ctx context.Context, request []byte, feastFeatures []*types.FeatureTable) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.enrichRequest")
	defer span.Finish()

	feastFeatureMap := make(map[string]*types.FeatureTable)
	for _, ft := range feastFeatures {
		feastFeatureMap[ft.Name] = ft
	}

	feastFeatureJSON, err := json.Marshal(feastFeatureMap)
	if err != nil {
		return nil, err
	}

	out, err := jsonparser.Set(request, feastFeatureJSON, transformer.FeastFeatureJSONField)
	if err != nil {
		return nil, err
	}

	return out, err
}
