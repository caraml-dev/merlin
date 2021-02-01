package feast

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
	feast "github.com/feast-dev/feast/sdk/go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
)

// Options for the Feast transformer.
type Options struct {
	ServingURL string `envconfig:"FEAST_SERVING_URL" required:"true"`
}

// Transformer wraps feast serving client to retrieve features.
type Transformer struct {
	feastClient feast.Client
	config      *transformer.StandardTransformerConfig
	logger      *zap.Logger
}

// NewTransformer initializes a new Transformer.
func NewTransformer(feastClient feast.Client, config *transformer.StandardTransformerConfig, logger *zap.Logger) *Transformer {
	return &Transformer{
		feastClient: feastClient,
		config:      config,
		logger:      logger,
	}
}

type FeastFeature struct {
	Columns []string  `json:"columns"`
	Data    []float64 `json:"data"`
}

// Transform retrieves the Feast features values and add them into the request.
func (t *Transformer) Transform(ctx context.Context, request []byte) ([]byte, error) {
	feastFeatures := make(map[string]FeastFeature, len(t.config.TransformerConfig.Feast))

	for _, config := range t.config.TransformerConfig.Feast {
		entities := []feast.Row{}
		entityIDs := make(map[string]int, len(config.Entities))

		for _, entity := range config.Entities {
			vals, err := getValuesFromJSONPayload(request, entity)
			if err != nil {
				return nil, fmt.Errorf("unable to extract entity %s: %v", entity.Name, err)
			}

			for _, val := range vals {
				entities = append(entities, feast.Row{
					entity.Name: val,
				})
			}

			entityIDs[entity.Name] = 1
		}

		features := []string{}
		for _, feature := range config.Features {
			features = append(features, feature.Name)
		}

		feastRequest := feast.OnlineFeaturesRequest{
			Project:  config.Project,
			Entities: entities,
			Features: features,
		}
		t.logger.Debug("feast_request", zap.Any("feast_request", feastRequest))

		feastResponse, err := t.feastClient.GetOnlineFeatures(ctx, &feastRequest)
		if err != nil {
			return nil, err
		}
		t.logger.Debug("feast_response", zap.Any("feast_response", feastResponse.Rows()))

		data := []float64{}
		for _, feastRow := range feastResponse.Rows() {
			for featureID, featureValue := range feastRow {
				if _, ok := entityIDs[featureID]; ok {
					continue
				}

				data = append(data, featureValue.GetDoubleVal())
			}
		}

		feastFeatures[config.Entities[0].Name] = FeastFeature{
			Columns: features,
			Data:    data,
		}
	}

	feastFeatureJSON, err := json.Marshal(feastFeatures)
	if err != nil {
		return nil, err
	}

	out, err := jsonparser.Set(request, feastFeatureJSON, transformer.FeastFeatureJSONField)
	if err != nil {
		return nil, err
	}

	return out, err
}
