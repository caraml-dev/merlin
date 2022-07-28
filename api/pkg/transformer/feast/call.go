package feast

import (
	"context"
	"fmt"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/spec"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type call struct {
	featureTableSpec *spec.FeatureTable
	columns          []string
	entitySet        map[string]bool
	defaultValues    defaultValues

	feastClient   StorageClient
	servingSource spec.ServingSource

	logger *zap.Logger

	statusMonitoringEnabled bool
	valueMonitoringEnabled  bool
}

// do create request to feast and return the result as table
func (fc *call) do(ctx context.Context, entityList []feast.Row, features []string) callResult {
	tableName := GetTableName(fc.featureTableSpec)

	feastSource := spec.ServingSource_name[int32(fc.servingSource)]
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.doBatchCall")
	span.SetTag("feast.source", feastSource)
	span.SetTag("table", tableName)
	defer span.Finish()

	feastRequest := feast.OnlineFeaturesRequest{
		Project:  fc.featureTableSpec.Project,
		Entities: entityList,
		Features: features,
	}

	startTime := time.Now()
	feastResponse, err := fc.feastClient.GetOnlineFeatures(ctx, &feastRequest)
	durationMs := time.Now().Sub(startTime).Milliseconds()
	if err != nil {
		feastLatency.WithLabelValues("error", feastSource).Observe(float64(durationMs))
		feastError.WithLabelValues(feastSource).Inc()

		return callResult{featureTable: nil, err: err}
	}
	feastLatency.WithLabelValues("success", feastSource).Observe(float64(durationMs))
	featureTable, err := fc.processResponse(feastResponse)
	if err != nil {
		return callResult{featureTable: nil, err: err}
	}

	return callResult{tableName: tableName, featureTable: featureTable, err: nil}
}

func getFeatureTypeMapping(featureTableSpec *spec.FeatureTable) map[string]types.ValueType_Enum {
	mapping := make(map[string]types.ValueType_Enum, len(featureTableSpec.Features))
	for _, feature := range featureTableSpec.Features {
		feastValType := types.ValueType_Enum(types.ValueType_Enum_value[feature.ValueType])
		mapping[feature.Name] = feastValType
	}
	return mapping
}

// processResponse process response from feast serving and create an internal feature table representation of it
func (fc *call) processResponse(feastResponse *feast.OnlineFeaturesResponse) (*internalFeatureTable, error) {
	responseStatus := feastResponse.Statuses()
	responseRows := feastResponse.Rows()
	entities := make([]feast.Row, len(responseRows))
	valueRows := make([]transTypes.ValueRow, len(responseRows))
	columnTypes := make([]types.ValueType_Enum, len(fc.columns))
	columnTypeMapping := getFeatureTypeMapping(fc.featureTableSpec)

	for rowIdx, feastRow := range responseRows {
		valueRow := make(transTypes.ValueRow, len(fc.columns))

		// create entity object, for cache key purpose
		entity := feast.Row{}
		for colIdx, column := range fc.columns {
			var rawValue *types.Value

			featureStatus := responseStatus[rowIdx][column]
			switch featureStatus {
			case serving.GetOnlineFeaturesResponse_PRESENT:
				rawValue = feastRow[column]
				// set value of entity
				_, isEntity := fc.entitySet[column]
				if isEntity {
					entity[column] = rawValue
				}
			case serving.GetOnlineFeaturesResponse_NOT_FOUND, serving.GetOnlineFeaturesResponse_NULL_VALUE, serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE:
				if columnTypes[colIdx] == types.ValueType_INVALID {
					columnTypes[colIdx] = columnTypeMapping[column]
				}
				defVal, ok := fc.defaultValues.GetDefaultValue(fc.featureTableSpec.Project, column)
				if !ok {
					// no default value is specified, we populate with nil
					valueRow[colIdx] = nil
					continue
				}
				rawValue = defVal

			default:
				return nil, fmt.Errorf("unsupported feature retrieval status for column %s: %s", column, featureStatus)
			}

			val, valType, err := converter.ExtractFeastValue(rawValue)
			if err != nil {
				return nil, err
			}

			// if previously we detected that the column type is invalid then we set it to the correct type
			if valType != types.ValueType_INVALID {
				columnTypes[colIdx] = valType
			}
			valueRow[colIdx] = val

			fc.recordMetrics(val, column, featureStatus)
		}

		entities[rowIdx] = entity
		valueRows[rowIdx] = valueRow
	}

	return &internalFeatureTable{
		entities:    entities,
		columnNames: fc.columns,
		columnTypes: columnTypes,
		valueRows:   valueRows,
	}, nil
}

func (fc *call) recordMetrics(val interface{}, column string, featureStatus serving.GetOnlineFeaturesResponse_FieldStatus) {
	// put behind feature toggle since it will generate high cardinality metrics
	if fc.valueMonitoringEnabled {
		v, err := converter.ToFloat64(val)
		if err == nil {
			feastFeatureSummary.WithLabelValues(column).Observe(v)
		}
	}

	// put behind feature toggle since it will generate high cardinality metrics
	if fc.statusMonitoringEnabled {
		feastFeatureStatus.WithLabelValues(column, featureStatus.String()).Inc()
	}
}
