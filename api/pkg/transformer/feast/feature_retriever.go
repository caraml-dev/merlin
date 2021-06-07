package feast

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
)

const (
	hystrixCommandName = "feast-hystrix"

	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

type FeatureRetriever interface {
	RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error)
	RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error)
}

type FeastRetriever struct {
	feastClient       feast.Client
	entityExtractor   *EntityExtractor
	featureTableSpecs []*spec.FeatureTable

	defaultValues defaultValues
	options       *Options
	cache         cache.Cache
	logger        *zap.Logger
	retrier       heimdall.Retriable
}

func NewFeastRetriever(
	feastClient feast.Client,
	entityExtractor *EntityExtractor,
	featureTableSpecs []*spec.FeatureTable,
	options *Options,
	cache cache.Cache,
	logger *zap.Logger) *FeastRetriever {

	defaultValues := compileDefaultValues(featureTableSpecs)

	retrier := heimdall.NewNoRetrier()
	if options.FeastHystrixRetryCount > 0 {
		backoffInterval := options.FeastHystrixRetryMaxJitterInterval
		maximumJitterInterval := options.FeastHystrixRetryBackoffInterval
		backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)
		retrier = heimdall.NewRetrier(backoff)
	}

	hystrix.ConfigureCommand("feast-hystrix", hystrix.CommandConfig{
		Timeout:                durationToInt(options.FeastHystrixTimeout, time.Millisecond),
		MaxConcurrentRequests:  options.FeastHystrixMaxConcurrentRequests,
		RequestVolumeThreshold: options.FeastHystrixRequestVolumeThreshold,
		SleepWindow:            options.FeastHystrixSleepWindow,
		ErrorPercentThreshold:  options.FeastHystrixErrorPercentThreshold,
	})

	return &FeastRetriever{
		feastClient:       feastClient,
		entityExtractor:   entityExtractor,
		featureTableSpecs: featureTableSpecs,
		defaultValues:     defaultValues,
		options:           options,
		cache:             cache,
		logger:            logger,
		retrier:           retrier,
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

	FeastHystrixTimeout                time.Duration `envconfig:"FEAST_HYSTRIX_TIMEOUT" default:"1s"`
	FeastHystrixMaxConcurrentRequests  int           `envconfig:"FEAST_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	FeastHystrixRequestVolumeThreshold int           `envconfig:"FEAST_HYSTRIX_REQUEST_VOLUME_THRESHOLD" default:"100"`
	FeastHystrixSleepWindow            int           `envconfig:"FEAST_HYSTRIX_SLEEP_WINDOW" default:"10"`
	FeastHystrixErrorPercentThreshold  int           `envconfig:"FEAST_HYSTRIX_ERROR_PERCENT_THRESHOLD" default:"25"`
	FeastHystrixRetryCount             int           `envconfig:"FEAST_HYSTRIX_RETRY_COUNT" default:"0"`
	FeastHystrixRetryBackoffInterval   time.Duration `envconfig:"FEAST_HYSTRIX_RETRY_BACKOFF_INTERVAL" default:"5ms"`
	FeastHystrixRetryMaxJitterInterval time.Duration `envconfig:"FEAST_HYSTRIX_RETRY_MAX_JITTER_INTERVAL" default:"5ms"`
}

const defaultProjectName = "default"

type entityFeaturePair struct {
	entity      feast.Row
	value       transTypes.ValueRow
	columnTypes []types.ValueType_Enum
}

type batchResult struct {
	featuresData transTypes.ValueRows
	columnTypes  []types.ValueType_Enum
	err          error
}

type parallelCallResult struct {
	featureTable *transTypes.FeatureTable
	err          error
}

func (fr *FeastRetriever) RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.RetrieveFromRequest")
	defer span.Finish()

	sr := symbol.NewRegistryWithCompiledJSONPath(fr.entityExtractor.compiledJsonPath)
	sr.SetRawRequestJSON(requestJson)

	return fr.RetrieveFeatureOfEntityInSymbolRegistry(ctx, sr)
}

func (fr *FeastRetriever) RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.RetrieveFromSymbolRegistry")
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
		featureTableSpec.TableName = GetTableName(featureTableSpec)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.getFeaturePerTable")
	span.SetTag("table.name", featureTableSpec.TableName)
	defer span.Finish()

	entities, err := fr.buildEntityRows(ctx, symbolRegistry, featureTableSpec.Entities, featureTableSpec.TableName)
	if err != nil {
		return nil, err
	}

	featureTable, err := fr.getFeatureTable(ctx, entities, featureTableSpec)
	if err != nil {
		return nil, err
	}
	return featureTable, nil
}

func (fr *FeastRetriever) buildEntityRows(ctx context.Context, symbolRegistry symbol.Registry, configEntities []*spec.Entity, tableName string) ([]feast.Row, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.buildEntityRows")
	span.SetTag("table.name", tableName)
	defer span.Finish()

	var allSeries [][]*types.Value
	maxLength := 1

	for k, configEntity := range configEntities {
		vals, err := fr.entityExtractor.ExtractValuesFromSymbolRegistry(symbolRegistry, configEntity)
		if err != nil {
			return nil, fmt.Errorf("unable to extract entity %s: %v", configEntity.Name, err)
		}

		seriesLength := len(vals)

		if seriesLength != 1 && maxLength != 1 && seriesLength != maxLength {
			return nil, fmt.Errorf("entity %s has different dimension", configEntities[k].Name)
		}

		if seriesLength > maxLength {
			maxLength = seriesLength
		}

		allSeries = append(allSeries, vals)
	}

	entities := make([]feast.Row, maxLength)
	for k := range entities {
		entities[k] = feast.Row{}
	}

	for s, series := range allSeries {
		entityName := configEntities[s].Name
		if len(series) == 1 {
			entities = broadcastSeries(entityName, series, entities)
		}

		if len(series) > 1 {
			entities = addSeries(entityName, series, entities)
		}
	}

	return entities, nil
}

func broadcastSeries(entityName string, series []*types.Value, entities []feast.Row) []feast.Row {
	for _, entity := range entities {
		entity[entityName] = series[0]
	}
	return entities
}

func addSeries(entityName string, series []*types.Value, entities []feast.Row) []feast.Row {
	for idx, entity := range entities {
		entity[entityName] = series[idx]
	}
	return entities
}

func (fr *FeastRetriever) getFeatureTable(ctx context.Context, entities []feast.Row, featureTableSpec *spec.FeatureTable) (*transTypes.FeatureTable, error) {

	var cachedValues transTypes.ValueRows
	entityNotInCache := entities
	var columnTypes []types.ValueType_Enum

	if fr.options.CacheEnabled {
		cachedValues, columnTypes, entityNotInCache = fetchFeaturesFromCache(fr.cache, entities, featureTableSpec.Project)
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

			var feastResponse *feast.OnlineFeaturesResponse
			var err error

			for i := 0; i <= fr.options.FeastHystrixRetryCount; i++ {
				err = hystrix.Do(hystrixCommandName, func() error {
					startTime := time.Now()
					feastResponse, err = fr.feastClient.GetOnlineFeatures(ctx, &feastRequest)
					durationMs := time.Now().Sub(startTime).Milliseconds()
					if err != nil {
						feastLatency.WithLabelValues("error", fmt.Sprint(i)).Observe(float64(durationMs))
						feastError.Inc()

						batchResultChan <- batchResult{featuresData: nil, err: err}
						return err
					}

					feastLatency.WithLabelValues("success", fmt.Sprint(i)).Observe(float64(durationMs))
					return nil
				}, nil)

				if err != nil {
					backoffTime := fr.retrier.NextInterval(i)
					time.Sleep(backoffTime)
					continue
				}

				break
			}

			fr.logger.Debug("feast_response", zap.Any("feast_response", feastResponse.Rows()))

			entityFeaturePairs, err := fr.buildFeastFeaturesData(ctx, feastResponse, featureTableSpec.Project, columns, entityIndices)
			if err != nil {
				batchResultChan <- batchResult{featuresData: nil, columnTypes: nil, err: err}
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

			batchResultChan <- batchResult{featuresData: featuresData, columnTypes: entityFeaturePairs[0].columnTypes, err: nil}
		}(featureTableSpec.Project, batchedEntities, columns)
	}

	data := cachedValues
	for i := 0; i < numOfBatch; i++ {
		res := <-batchResultChan
		if res.err != nil {
			return nil, res.err
		}
		data = append(data, res.featuresData...)
		columnTypes = mergeColumnTypes(columnTypes, res.columnTypes)
	}

	return &transTypes.FeatureTable{
		Name:        featureTableSpec.TableName,
		Columns:     columns,
		ColumnTypes: columnTypes,
		Data:        data,
	}, nil
}

func (fr *FeastRetriever) buildFeastFeaturesData(ctx context.Context, feastResponse *feast.OnlineFeaturesResponse, projectName string, columns []string, entityIndexMap map[int]int) ([]entityFeaturePair, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.buildFeastFeaturesData")
	defer span.Finish()

	var data []entityFeaturePair
	status := feastResponse.Statuses()

	columnTypes := make([]types.ValueType_Enum, len(columns))
	for rowIdx, feastRow := range feastResponse.Rows() {
		var row transTypes.ValueRow

		// create entity object, for cache key purpose
		entity := feast.Row{}
		for colIdx, column := range columns {
			featureStatus := status[rowIdx][column]
			_, isEntityIndex := entityIndexMap[colIdx]
			switch featureStatus {
			case serving.GetOnlineFeaturesResponse_PRESENT:
				rawValue := feastRow[column]

				// set value of entity
				if isEntityIndex {
					entity[column] = rawValue
				}

				featVal, valType, err := getFeatureValue(rawValue)
				if err != nil {
					return nil, err
				}

				if columnTypes[colIdx] == types.ValueType_INVALID {
					columnTypes[colIdx] = valType
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
				defVal, ok := fr.defaultValues.GetDefaultValue(projectName, column)
				if !ok {
					row = append(row, nil)
					continue
				}
				featVal, valType, err := getFeatureValue(defVal)
				if err != nil {
					return nil, err
				}

				if columnTypes[colIdx] == types.ValueType_INVALID {
					columnTypes[colIdx] = valType
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
		data = append(data, entityFeaturePair{entity: entity, value: row, columnTypes: columnTypes})
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

func GetTableName(featureTableSpec *spec.FeatureTable) string {
	if featureTableSpec.TableName != "" {
		return featureTableSpec.TableName
	}

	entityNames := make([]string, 0)
	for _, n := range featureTableSpec.Entities {
		entityNames = append(entityNames, n.Name)
	}

	tableName := strings.Join(entityNames, "_")
	if featureTableSpec.Project != defaultProjectName {
		tableName = featureTableSpec.Project + "_" + tableName
	}

	return tableName
}

func getFeatureValue(val *types.Value) (interface{}, types.ValueType_Enum, error) {
	switch val.Val.(type) {
	case *types.Value_StringVal:
		return val.GetStringVal(), types.ValueType_STRING, nil
	case *types.Value_DoubleVal:
		return val.GetDoubleVal(), types.ValueType_DOUBLE, nil
	case *types.Value_FloatVal:
		return val.GetFloatVal(), types.ValueType_FLOAT, nil
	case *types.Value_Int32Val:
		return val.GetInt32Val(), types.ValueType_INT32, nil
	case *types.Value_Int64Val:
		return val.GetInt64Val(), types.ValueType_INT64, nil
	case *types.Value_BoolVal:
		return val.GetBoolVal(), types.ValueType_BOOL, nil
	case *types.Value_StringListVal:
		return val.GetStringListVal().GetVal(), types.ValueType_STRING_LIST, nil
	case *types.Value_DoubleListVal:
		return val.GetDoubleListVal().GetVal(), types.ValueType_DOUBLE_LIST, nil
	case *types.Value_FloatListVal:
		return val.GetFloatListVal().GetVal(), types.ValueType_FLOAT_LIST, nil
	case *types.Value_Int32ListVal:
		return val.GetInt32ListVal().GetVal(), types.ValueType_INT32_LIST, nil
	case *types.Value_Int64ListVal:
		return val.GetInt64ListVal().GetVal(), types.ValueType_INT64_LIST, nil
	case *types.Value_BoolListVal:
		return val.GetBoolListVal().GetVal(), types.ValueType_BOOL_LIST, nil
	case *types.Value_BytesVal:
		return base64.StdEncoding.EncodeToString(val.GetBytesVal()), types.ValueType_STRING, nil
	case *types.Value_BytesListVal:
		results := make([]string, 0)
		for _, bytes := range val.GetBytesListVal().GetVal() {
			results = append(results, base64.StdEncoding.EncodeToString(bytes))
		}
		return results, types.ValueType_STRING_LIST, nil
	default:
		return nil, types.ValueType_INVALID, fmt.Errorf("unknown feature cacheValue type: %T", val.Val)
	}
}

func mergeColumnTypes(dst []types.ValueType_Enum, src []types.ValueType_Enum) []types.ValueType_Enum {
	if len(dst) == 0 {
		dst = src
		return dst
	}

	for i, t := range dst {
		if t == types.ValueType_INVALID {
			dst[i] = src[i]
		}
	}
	return dst
}

func durationToInt(duration, unit time.Duration) int {
	durationAsNumber := duration / unit

	if int64(durationAsNumber) > int64(maxInt) {
		// Returning max possible value seems like best possible solution here
		// the alternative is to panic as there is no way of returning an error
		// without changing the NewClient API
		return maxInt
	}
	return int(durationAsNumber)
}
