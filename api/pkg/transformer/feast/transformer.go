package feast

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
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
	Columns []string        `json:"columns"`
	Data    [][]interface{} `json:"data"`
}

// Transform retrieves the Feast features values and add them into the request.
func (t *Transformer) Transform(ctx context.Context, request []byte) ([]byte, error) {
	feastFeatures := make(map[string]FeastFeature, len(t.config.TransformerConfig.Feast))

	for _, config := range t.config.TransformerConfig.Feast {
		var entities []feast.Row
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
		}

		var features []string
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

		var columns []string
		for _, entity := range config.Entities {
			columns = append(columns, entity.Name)
		}
		columns = append(columns, features...)

		var data [][]interface{}
		status := feastResponse.Statuses()
		for i, feastRow := range feastResponse.Rows() {
			var row []interface{}
			for _, column := range columns {
				switch status[i][column] {
				case serving.GetOnlineFeaturesResponse_PRESENT:
					rawValue := feastRow[column]
					featVal, err := getFeatureValue(rawValue)
					if err != nil {
						return nil, err
					}
					row = append(row, featVal)
				case serving.GetOnlineFeaturesResponse_NOT_FOUND:
					row = append(row, nil)
				case serving.GetOnlineFeaturesResponse_NULL_VALUE:
					row = append(row, nil)
				case serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE:
					row = append(row, nil)
				default:
					return nil, fmt.Errorf("")
				}
			}
			data = append(data, row)
		}

		tableName := createTableName(config.Entities)
		feastFeatures[tableName] = FeastFeature{
			Columns: columns,
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

func createTableName(entities []*transformer.Entity) string {
	entityNames := make([]string, 0)
	for _, n := range entities {
		entityNames = append(entityNames, n.Name)
	}

	return strings.Join(entityNames, "_")
}

func getFeatureValue(val *types.Value) (interface{}, error) {
	switch val.Val.(type) {
	case *types.Value_StringVal:
		return val.GetStringVal(), nil
	case *types.Value_DoubleVal:
		return val.GetDoubleVal(), nil
	case *types.Value_FloatVal:
		return val.GetFloatVal(), nil
	case *types.Value_Int32Val:
		return val.GetInt32Val(), nil
	case *types.Value_Int64Val:
		return val.GetInt64Val(), nil
	case *types.Value_BoolVal:
		return val.GetBoolVal(), nil
	case *types.Value_StringListVal:
		return val.GetStringListVal(), nil
	case *types.Value_DoubleListVal:
		return val.GetDoubleListVal(), nil
	case *types.Value_FloatListVal:
		return val.GetFloatListVal(), nil
	case *types.Value_Int32ListVal:
		return val.GetInt32ListVal(), nil
	case *types.Value_Int64ListVal:
		return val.GetInt64ListVal(), nil
	case *types.Value_BoolListVal:
		return val.GetBoolListVal(), nil
	case *types.Value_BytesVal:
		return val.GetBytesVal(), nil
	case *types.Value_BytesListVal:
		return val.GetBytesListVal(), nil
	default:
		return nil, fmt.Errorf("unknown feature value type: %T", val.Val)
	}
}
