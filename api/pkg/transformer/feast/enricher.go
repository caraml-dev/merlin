package feast

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
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
func (t *Enricher) Enrich(ctx context.Context, request types.Payload, _ map[string]string) (types.Payload, error) {
	originalRequest, validRequestType := request.(types.BytePayload)
	if !validRequestType {
		return nil, fmt.Errorf("supplied request for Enrich is not byte")
	}
	ctx, span := tracer.Start(ctx, "feast.Enrich")
	defer span.End()

	requestJson, err := request.AsInput()
	if err != nil {
		return nil, err
	}

	requestJsonObj, validType := requestJson.(types.JSONObject)
	if !validType {
		return nil, fmt.Errorf("got unexpected type %T for request", requestJsonObj)
	}

	feastFeatures, err := t.featureRetriever.RetrieveFeatureOfEntityInRequest(ctx, requestJsonObj)
	if err != nil {
		return nil, err
	}

	out, err := enrichRequest(ctx, originalRequest, feastFeatures)
	if err != nil {
		return nil, err
	}

	return types.BytePayload(out), err
}

func enrichRequest(ctx context.Context, request []byte, feastFeatures []*types.FeatureTable) ([]byte, error) {
	_, span := tracer.Start(ctx, "feast.enrichRequest")
	defer span.End()

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
