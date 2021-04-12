package feast

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
)

type FeatureRetriever interface {
	RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error)
	RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error)
}

type FeastRetriever struct {
	feastClient       feast.Client
	entityExtractor   *EntityExtractor
	featureTableSpecs []*spec.FeatureTable

	defaultValues map[string]*types.Value
	options       *Options
	cache         cache.Cache
	logger        *zap.Logger
}

func NewFeastRetriever(
	feastClient feast.Client,
	entityExtractor *EntityExtractor,
	featureTableSpecs []*spec.FeatureTable,
	options *Options,
	cache cache.Cache,
	logger *zap.Logger) *FeastRetriever {

	defaultValues := compileDefaultValues(featureTableSpecs)

	return &FeastRetriever{
		feastClient:       feastClient,
		entityExtractor:   entityExtractor,
		featureTableSpecs: featureTableSpecs,
		defaultValues:     defaultValues,
		options:           options,
		cache:             cache,
		logger:            logger,
	}
}

// Options for the Feast transformer.
type Options struct {
	ServingURL              string        `envconfig:"FEAST_SERVING_URL" required:"true"`
	StatusMonitoringEnabled bool          `envconfig:"FEAST_FEATURE_STATUS_MONITORING_ENABLED" default:"false"`
	ValueMonitoringEnabled  bool          `envconfig:"FEAST_FEATURE_VALUE_MONITORING_ENABLED" default:"false"`
	BatchSize               int           `envconfig:"FEAST_BATCH_SIZE" default:"50"`
	CacheEnabled            bool          `envconfig:"FEAST_CACHE_ENABLED" default:"true"`
	CacheTTL                time.Duration `envconfig:"FEAST_CACHE_TTL" default:"60s"`
}

const defaultProjectName = "default"

type entityFeaturePair struct {
	entity feast.Row
	value  transTypes.ValueRow
}

type batchResult struct {
	featuresData transTypes.ValueRows
	err          error
}

type parallelCallResult struct {
	featureTable *transTypes.FeatureTable
	err          error
}

func (fr *FeastRetriever) RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.Transform")
	defer span.Finish()

	sr := symbol.NewRegistryWithCompiledJSONPath(fr.entityExtractor.compiledJsonPath)
	sr.SetRawRequestJSON(requestJson)

	return fr.RetrieveFeatureOfEntityInSymbolRegistry(ctx, sr)
}

func (fr *FeastRetriever) RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.Transform")
	defer span.Finish()

	nbTables := len(fr.featureTableSpecs)
	feastFeatures := make([]*transTypes.FeatureTable, 0)

	// parallelize feast call per feature table
	resChan := make(chan parallelCallResult, nbTables)
	for _, config := range fr.featureTableSpecs {
		go func(featureTableSpec *spec.FeatureTable) {
			featureTable, err := fr.getFeaturePerTable(ctx, symbolRegistry, featureTableSpec)
			resChan <- parallelCallResult{featureTable, err}
		}(config)
	}

	// collect result
	for i := 0; i < cap(resChan); i++ {
		res := <-resChan
		if res.err != nil {
			return nil, res.err
		}
		feastFeatures = append(feastFeatures, res.featureTable)
	}

	return feastFeatures, nil
}

func (fr *FeastRetriever) getFeaturePerTable(ctx context.Context, symbolRegistry symbol.Registry, featureTableSpec *spec.FeatureTable) (*transTypes.FeatureTable, error) {
	if featureTableSpec.TableName == "" {
		featureTableSpec.TableName = createTableName(featureTableSpec.Entities, featureTableSpec.Project)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.getFeaturePerTable")
	span.SetTag("table.name", featureTableSpec.TableName)
	defer span.Finish()

	entities, err := fr.buildEntityRows(ctx, symbolRegistry, featureTableSpec.Entities)
	if err != nil {
		return nil, err
	}

	featureTable, err := fr.getFeatureTable(ctx, entities, featureTableSpec)
	if err != nil {
		return nil, err
	}
	return featureTable, nil
}

func (fr *FeastRetriever) buildEntityRows(ctx context.Context, symbolRegistry symbol.Registry, configEntities []*spec.Entity) ([]feast.Row, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.buildEntityRows")
	defer span.Finish()

	var entities []feast.Row

	for _, configEntity := range configEntities {
		vals, err := fr.entityExtractor.ExtractValuesFromSymbolRegistry(symbolRegistry, configEntity)
		if err != nil {
			return nil, fmt.Errorf("unable to extract entity %s: %v", configEntity.Name, err)
		}

		if len(entities) == 0 {
			for _, val := range vals {
				entities = append(entities, feast.Row{
					configEntity.Name: val,
				})
			}
		} else {
			newEntities := []feast.Row{}
			for _, entity := range entities {
				for _, val := range vals {
					newFeastRow := feast.Row{}
					for k, v := range entity {
						newFeastRow[k] = v
					}

					newFeastRow[configEntity.Name] = val
					newEntities = append(newEntities, newFeastRow)
				}
			}
			entities = newEntities
		}
	}

	return entities, nil
}

func (fr *FeastRetriever) getFeatureTable(ctx context.Context, entities []feast.Row, featureTableSpec *spec.FeatureTable) (*transTypes.FeatureTable, error) {

	var cachedValues transTypes.ValueRows
	entityNotInCache := entities

	if fr.options.CacheEnabled {
		cachedValues, entityNotInCache = fetchFeaturesFromCache(fr.cache, entities, featureTableSpec.Project)
	}

	var features []string
	for _, feature := range featureTableSpec.Features {
		features = append(features, feature.Name)
	}

	numOfBatchBeforeCeil := float64(len(entityNotInCache)) / float64(fr.options.BatchSize)
	numOfBatch := int(math.Ceil(numOfBatchBeforeCeil))

	batchResultChan := make(chan batchResult, numOfBatch)
	columns := getColumnNames(featureTableSpec)
	entityIndices := getEntityIndicesFromColumns(columns, featureTableSpec.Entities)
	for i := 0; i < numOfBatch; i++ {
		startIndex := i * fr.options.BatchSize
		endIndex := len(entityNotInCache)
		if endIndex > startIndex+fr.options.BatchSize {
			endIndex = startIndex + fr.options.BatchSize
		}
		batchedEntities := entityNotInCache[startIndex:endIndex]

		go func(project string, entityList []feast.Row, columns []string) {
			feastRequest := feast.OnlineFeaturesRequest{
				Project:  project,
				Entities: entityList,
				Features: features,
			}
			startTime := time.Now()
			feastResponse, err := fr.feastClient.GetOnlineFeatures(ctx, &feastRequest)
			durationMs := time.Now().Sub(startTime).Milliseconds()
			if err != nil {
				feastLatency.WithLabelValues("error").Observe(float64(durationMs))
				feastError.Inc()

				batchResultChan <- batchResult{featuresData: nil, err: err}
				return
			}
			feastLatency.WithLabelValues("success").Observe(float64(durationMs))

			fr.logger.Debug("feast_response", zap.Any("feast_response", feastResponse.Rows()))

			entityFeaturePairs, err := fr.buildFeastFeaturesData(ctx, feastResponse, columns, entityIndices)
			if err != nil {
				batchResultChan <- batchResult{featuresData: nil, err: err}
				return
			}

			var featuresData transTypes.ValueRows
			for _, data := range entityFeaturePairs {
				featuresData = append(featuresData, data.value)
			}

			if fr.options.CacheEnabled {
				if err := insertMultipleFeaturesToCache(fr.cache, entityFeaturePairs, project, fr.options.CacheTTL); err != nil {
					fr.logger.Error("insert_to_cache", zap.Any("error", err))
				}
			}

			batchResultChan <- batchResult{featuresData: featuresData, err: nil}
		}(featureTableSpec.Project, batchedEntities, columns)
	}

	data := cachedValues

	for i := 0; i < numOfBatch; i++ {
		res := <-batchResultChan
		if res.err != nil {
			return nil, res.err
		}
		data = append(data, res.featuresData...)
	}

	return &transTypes.FeatureTable{
		Name:    featureTableSpec.TableName,
		Columns: columns,
		Data:    data,
	}, nil
}

func (fr *FeastRetriever) buildFeastFeaturesData(ctx context.Context, feastResponse *feast.OnlineFeaturesResponse, columns []string, entityIndexMap map[int]int) ([]entityFeaturePair, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.buildFeastFeaturesData")
	defer span.Finish()

	var data []entityFeaturePair
	status := feastResponse.Statuses()

	for i, feastRow := range feastResponse.Rows() {
		var row transTypes.ValueRow

		// create entity object, for cache key purpose
		entity := feast.Row{}
		for index, column := range columns {
			featureStatus := status[i][column]
			_, isEntityIndex := entityIndexMap[index]
			switch featureStatus {
			case serving.GetOnlineFeaturesResponse_PRESENT:
				rawValue := feastRow[column]

				// set value of entity
				if isEntityIndex {
					entity[column] = rawValue
				}

				featVal, err := getFeatureValue(rawValue)
				if err != nil {
					return nil, err
				}
				row = append(row, featVal)

				// put behind feature toggle since it will generate high cardinality metrics
				if fr.options.ValueMonitoringEnabled {
					v, err := getFloatValue(featVal)
					if err != nil {
						continue
					}
					feastFeatureSummary.WithLabelValues(column).Observe(v)
				}
			case serving.GetOnlineFeaturesResponse_NOT_FOUND, serving.GetOnlineFeaturesResponse_NULL_VALUE, serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE:
				defVal, ok := fr.defaultValues[column]
				if !ok {
					row = append(row, nil)
					continue
				}
				featVal, err := getFeatureValue(defVal)
				if err != nil {
					return nil, err
				}
				row = append(row, featVal)
			default:
				return nil, fmt.Errorf("Unsupported feature retrieval status: %s", featureStatus)
			}
			// put behind feature toggle since it will generate high cardinality metrics
			if fr.options.StatusMonitoringEnabled {
				feastFeatureStatus.WithLabelValues(column, featureStatus.String()).Inc()
			}
		}
		data = append(data, entityFeaturePair{entity: entity, value: row})
	}

	return data, nil
}

func getColumnNames(config *spec.FeatureTable) []string {
	columns := make([]string, 0, len(config.Entities)+len(config.Features))
	for _, entity := range config.Entities {
		columns = append(columns, entity.Name)
	}
	for _, feature := range config.Features {
		columns = append(columns, feature.Name)
	}
	return columns
}

func getEntityIndicesFromColumns(columns []string, entitiesConfig []*spec.Entity) map[int]int {
	indicesMapping := make(map[int]int, len(entitiesConfig))
	entitiesConfigMap := make(map[string]*spec.Entity)
	for _, entityConfig := range entitiesConfig {
		entitiesConfigMap[entityConfig.Name] = entityConfig
	}
	for i, column := range columns {
		if _, found := entitiesConfigMap[column]; found {
			indicesMapping[i] = i
		}
	}
	return indicesMapping
}

func getFloatValue(val interface{}) (float64, error) {
	switch i := val.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	default:
		return math.NaN(), errors.New("getFloat: unknown value is of incompatible type")
	}
}

func createTableName(entities []*spec.Entity, project string) string {
	entityNames := make([]string, 0)
	for _, n := range entities {
		entityNames = append(entityNames, n.Name)
	}

	tableName := strings.Join(entityNames, "_")
	if project != defaultProjectName {
		tableName = project + "_" + tableName
	}

	return tableName
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
		return val.GetStringListVal().GetVal(), nil
	case *types.Value_DoubleListVal:
		return val.GetDoubleListVal().GetVal(), nil
	case *types.Value_FloatListVal:
		return val.GetFloatListVal().GetVal(), nil
	case *types.Value_Int32ListVal:
		return val.GetInt32ListVal().GetVal(), nil
	case *types.Value_Int64ListVal:
		return val.GetInt64ListVal().GetVal(), nil
	case *types.Value_BoolListVal:
		return val.GetBoolListVal().GetVal(), nil
	case *types.Value_BytesVal:
		return base64.StdEncoding.EncodeToString(val.GetBytesVal()), nil
	case *types.Value_BytesListVal:
		results := make([]string, 0)
		for _, bytes := range val.GetBytesListVal().GetVal() {
			results = append(results, base64.StdEncoding.EncodeToString(bytes))
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unknown feature value type: %T", val.Val)
	}
}
