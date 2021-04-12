package feast

import (
	"context"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
)

// Transformer wraps feast serving client to retrieve features.
type Transformer struct {
	featureRetriever FeatureRetriever
	logger           *zap.Logger
}

// NewTransformer initializes a new Transformer.
func NewTransformer(featureRetriever FeatureRetriever, logger *zap.Logger) (*Transformer, error) {
	return &Transformer{
		featureRetriever: featureRetriever,
		logger:           logger,
	}, nil
}

// Transform retrieves the Feast features values and add them into the request.
func (t *Transformer) Transform(ctx context.Context, request []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.Transform")
	defer span.Finish()

	var requestJson transTypes.JSONObject
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

func enrichRequest(ctx context.Context, request []byte, feastFeatures []*transTypes.FeatureTable) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.enrichRequest")
	defer span.Finish()

	feastFeatureMap := make(map[string]*transTypes.FeatureTable)
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
