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

type batchCall struct {
	featureTableSpec *spec.FeatureTable
	columns          []string
	entityIndexMap   map[int]int
	defaultValues    defaultValues

	feastClient feast.Client
	feastURL    string

	logger *zap.Logger

	statusMonitoringEnabled bool
	valueMonitoringEnabled  bool
}

func (fc *batchCall) do(ctx context.Context, entityList []feast.Row, features []string) callResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.doBatchCall")
	span.SetTag("feast.url", fc.feastURL)
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
		feastLatency.WithLabelValues("error", fc.feastURL).Observe(float64(durationMs))
		feastError.WithLabelValues(fc.feastURL).Inc()

		return callResult{featureTable: nil, err: err}
	}
	feastLatency.WithLabelValues("success", fc.feastURL).Observe(float64(durationMs))

	featureTable, err := fc.processResponse(feastResponse)
	if err != nil {
		return callResult{featureTable: nil, err: err}
	}

	return callResult{tableName: GetTableName(fc.featureTableSpec), featureTable: featureTable, err: nil}
}

// processResponse process response from feast serving and create an internal feature table representation of it
func (fc *batchCall) processResponse(feastResponse *feast.OnlineFeaturesResponse) (*internalFeatureTable, error) {
	responseStatus := feastResponse.Statuses()
	responseRows := feastResponse.Rows()
	entities := make([]feast.Row, len(responseRows))
	valueRows := make([]transTypes.ValueRow, len(responseRows))
	columnTypes := make([]types.ValueType_Enum, len(fc.columns))

	for rowIdx, feastRow := range responseRows {
		valueRow := make(transTypes.ValueRow, len(fc.columns))

		// create entity object, for cache key purpose
		entity := feast.Row{}
		for colIdx, column := range fc.columns {
			featureStatus := responseStatus[rowIdx][column]
			_, isEntityIndex := fc.entityIndexMap[colIdx]
			var rawValue *types.Value
			switch featureStatus {
			case serving.GetOnlineFeaturesResponse_PRESENT:
				rawValue = feastRow[column]
				// set value of entity
				if isEntityIndex {
					entity[column] = rawValue
				}
			case serving.GetOnlineFeaturesResponse_NOT_FOUND, serving.GetOnlineFeaturesResponse_NULL_VALUE, serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE:
				defVal, ok := fc.defaultValues.GetDefaultValue(fc.featureTableSpec.Project, column)
				if !ok {
					// no default value is specified, we populate with nil
					valueRow[colIdx] = nil
					continue
				}
				rawValue = defVal
			default:
				return nil, fmt.Errorf("unsupported feature retrieval status: %s", featureStatus)
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

func (fc *batchCall) recordMetrics(val interface{}, column string, featureStatus serving.GetOnlineFeaturesResponse_FieldStatus) {
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
