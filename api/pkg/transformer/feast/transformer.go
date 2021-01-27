package feast

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"unsafe"

	"github.com/buger/jsonparser"
	feast "github.com/feast-dev/feast/sdk/go"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
)

// Options for the Feast transformer.
type Options struct {
	ServingAddress string `envconfig:"FEAST_SERVING_ADDRESS" required:"true"`
	ServingPort    int    `envconfig:"FEAST_SERVING_PORT" required:"true"`
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
			val, err := getValueFromJSONPayload(request, entity)
			if err != nil {
				log.Printf("JSON Path %s not found in request payload", entity.JsonPath)
			}

			entities = append(entities, feast.Row{
				entity.Name: val,
			})

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

	out, err := jsonparser.Set(request, feastFeatureJSON, "feast_features")
	if err != nil {
		return nil, err
	}

	return out, err
}

func getValueFromJSONPayload(body []byte, entity *transformer.Entity) (*feastType.Value, error) {
	// Retrieve value using JSON path
	value, typez, _, _ := jsonparser.Get(body, strings.Split(entity.JsonPath, ".")...)

	switch typez {
	case jsonparser.String, jsonparser.Number, jsonparser.Boolean:
		switch entity.ValueType {
		case transformer.StringValueType:
			// See: https://github.com/buger/jsonparser/blob/master/bytes_unsafe.go#L31
			return feast.StrVal(*(*string)(unsafe.Pointer(&value))), nil
		case transformer.DoubleValueType:
			val, err := jsonparser.ParseFloat(value)
			if err != nil {
				return nil, err
			}
			return feast.DoubleVal(val), nil
		case transformer.BooleanValueType:
			val, err := jsonparser.ParseBoolean(value)
			if err != nil {
				return nil, err
			}
			return feast.BoolVal(val), nil
		default:
			return nil, errors.Errorf("ENtity value type %s is not supported", entity.ValueType)
		}
	case jsonparser.Null:
		return nil, nil
	case jsonparser.NotExist:
		return nil, errors.Errorf("Field %s not found in the request payload: Key path not found", entity.JsonPath)
	default:
		return nil, errors.Errorf("Field %s can not be parsed, unsupported type: %s", entity.JsonPath, typez.String())
	}
}
