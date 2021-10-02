package feast

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/cespare/xxhash"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/golang/protobuf/proto"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	transTypes "github.com/gojek/merlin/pkg/transformer/types"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)

	// DefaultClientURLKey defines a key used to store and retrieve
	// the default Feast gRPC client from Clients map.
	// Used for backward compatibility (transformer will use this default client
	// if standard transformer config does not specify Feast's serving url).
	DefaultClientURLKey URL = "default"
)

type (
	URL     string
	Clients map[URL]feast.Client
)

type FeatureRetriever interface {
	RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error)
	RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error)
}

type FeastRetriever struct {
	feastClients      Clients
	featureCache      *featureCache
	entityExtractor   *EntityExtractor
	featureTableSpecs []*spec.FeatureTable

	defaultValues defaultValues
	options       *Options
	logger        *zap.Logger
}

func NewFeastRetriever(
	feastClients Clients,
	entityExtractor *EntityExtractor,
	featureTableSpecs []*spec.FeatureTable,
	options *Options,
	logger *zap.Logger) *FeastRetriever {

	defaultValues := compileDefaultValues(featureTableSpecs)

	hystrix.ConfigureCommand(options.FeastClientHystrixCommandName, hystrix.CommandConfig{
		Timeout:                durationToInt(options.FeastTimeout, time.Millisecond),
		MaxConcurrentRequests:  options.FeastClientMaxConcurrentRequests,
		RequestVolumeThreshold: options.FeastClientRequestVolumeThreshold,
		SleepWindow:            options.FeastClientSleepWindow,
		ErrorPercentThreshold:  options.FeastClientErrorPercentThreshold,
	})

	return &FeastRetriever{
		feastClients:      feastClients,
		entityExtractor:   entityExtractor,
		featureCache:      newFeatureCache(options.CacheTTL, options.CacheSizeInMB),
		featureTableSpecs: featureTableSpecs,
		defaultValues:     defaultValues,
		options:           options,
		cache:             cache,
		logger:            logger,
	}
}

// Options for the Feast transformer.
type Options struct {
	DefaultServingURL string   `envconfig:"DEFAULT_FEAST_SERVING_URL" required:"true"`
	ServingURLs       []string `envconfig:"FEAST_SERVING_URLS" required:"true"`

	StatusMonitoringEnabled bool          `envconfig:"FEAST_FEATURE_STATUS_MONITORING_ENABLED" default:"false"`
	ValueMonitoringEnabled  bool          `envconfig:"FEAST_FEATURE_VALUE_MONITORING_ENABLED" default:"false"`
	BatchSize               int           `envconfig:"FEAST_BATCH_SIZE" default:"50"`
	CacheEnabled            bool          `envconfig:"FEAST_CACHE_ENABLED" default:"true"`
	CacheTTL                time.Duration `envconfig:"FEAST_CACHE_TTL" default:"60s"`
	CacheSizeInMB           int           `envconfig:"CACHE_SIZE_IN_MB" default:"100"`

	FeastTimeout                      time.Duration `envconfig:"FEAST_TIMEOUT" default:"1s"`
	FeastClientHystrixCommandName     string        `envconfig:"FEAST_HYSTRIX_COMMAND_NAME" default:"feast_retrieval"`
	FeastClientMaxConcurrentRequests  int           `envconfig:"FEAST_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	FeastClientRequestVolumeThreshold int           `envconfig:"FEAST_HYSTRIX_REQUEST_VOLUME_THRESHOLD" default:"100"`
	FeastClientSleepWindow            int           `envconfig:"FEAST_HYSTRIX_SLEEP_WINDOW" default:"1000"` // How long, in milliseconds, to wait after a circuit opens before testing for recovery
	FeastClientErrorPercentThreshold  int           `envconfig:"FEAST_HYSTRIX_ERROR_PERCENT_THRESHOLD" default:"25"`
}

func (o Options) IsServingURLSupported(url string) bool {
	for _, supportedURL := range o.ServingURLs {
		if supportedURL == url {
			return true
		}
	}
	return false
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
	for _, featureTableSpec := range fr.featureTableSpecs {
		go func(featureTableSpec *spec.FeatureTable) {
			featureTable, err := fr.getFeaturePerTable(ctx, symbolRegistry, featureTableSpec)
			resChan <- parallelCallResult{featureTable, err}
		}(featureTableSpec)
	}

	// collect result
	for i := 0; i < cap(resChan); i++ {
		res := <-resChan
		if res.err != nil {
			ctx.Done()
			return nil, res.err
		}
		feastFeatures = append(feastFeatures, res.featureTable)
	}

	return feastFeatures, nil
}

func (fr *FeastRetriever) getFeaturePerTable(ctx context.Context, symbolRegistry symbol.Registry, featureTableSpec *spec.FeatureTable) (*transTypes.FeatureTable, error) {
	if featureTableSpec.TableName == "" {
		t := proto.Clone(featureTableSpec).(*spec.FeatureTable)
		t.TableName = GetTableName(t)
		featureTableSpec = t
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.getFeaturePerTable")
	span.SetTag("table.name", featureTableSpec.TableName)
	span.SetTag("feast.url", fr.getFeastURL(featureTableSpec.ServingUrl))
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

	uniqueEntities, err := dedupEntities(entities)
	if err != nil {
		return nil, err
	}

	return uniqueEntities, nil
}

func dedupEntities(rows []feast.Row) ([]feast.Row, error) {
	uniqueRows := make([]feast.Row, 0, len(rows))
	rowLookup := make(map[uint64]bool)
	for _, row := range rows {
		rowByte, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		rowHashVal := xxhash.Sum64(rowByte)
		if _, found := rowLookup[rowHashVal]; !found {
			uniqueRows = append(uniqueRows, row)
			rowLookup[rowHashVal] = true
		}
	}
	return uniqueRows, nil
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

type feastCall struct {
	featureTableSpec *spec.FeatureTable
	featureCache     *featureCache
	entityList       []feast.Row
	defaultValues    defaultValues

	feastClient feast.Client
	feastURL    string

	cacheEnabled bool
	logger       *zap.Logger

	statusMonitoringEnabled bool
	valueMonitoringEnabled  bool
}

func (fr *FeastRetriever) newFeastCall(
	featureTableSpec *spec.FeatureTable,
	entityList []feast.Row,
) (*feastCall, error) {
	feastClient, err := fr.getFeastClient(featureTableSpec.ServingUrl)
	if err != nil {
		return nil, err
	}

	return &feastCall{
		featureTableSpec: featureTableSpec,
		entityList:       entityList,
		defaultValues:    fr.defaultValues,

		feastClient: feastClient,
		feastURL:    fr.getFeastURL(featureTableSpec.ServingUrl),

		featureCache: fr.featureCache,
		cacheEnabled: fr.options.CacheEnabled,
		logger:       fr.logger,

		statusMonitoringEnabled: fr.options.StatusMonitoringEnabled,
		valueMonitoringEnabled:  fr.options.ValueMonitoringEnabled,
	}, nil
}

func (fr *FeastRetriever) getFeastClient(url string) (feast.Client, error) {
	if url == "" {
		return fr.feastClients[DefaultClientURLKey], nil
	}

	client, ok := fr.feastClients[URL(url)]
	if ok {
		return client, nil
	}

	return nil, errors.New("invalid feast serving url")
}

func (fr *FeastRetriever) getFeastURL(url string) string {
	if url == "" {
		return fr.options.DefaultServingURL
	}
	return url
}

func (fc *feastCall) do(
	ctx context.Context,
	features []string,
	columns []string,
	entityIndexMap map[int]int,
) batchResult {
	feastRequest := feast.OnlineFeaturesRequest{
		Project:  fc.featureTableSpec.Project,
		Entities: fc.entityList,
		Features: features,
	}

	feastURL := fc.feastURL

	startTime := time.Now()
	feastResponse, err := fc.feastClient.GetOnlineFeatures(ctx, &feastRequest)
	durationMs := time.Now().Sub(startTime).Milliseconds()
	if err != nil {
		feastLatency.WithLabelValues("error", feastURL).Observe(float64(durationMs))
		feastError.WithLabelValues(feastURL).Inc()

		return batchResult{featuresData: nil, err: err}
	}
	feastLatency.WithLabelValues("success", feastURL).Observe(float64(durationMs))

	entityFeaturePairs, err := fc.buildFeastFeaturesData(ctx, feastResponse, fc.featureTableSpec.Project, columns, entityIndexMap)
	if err != nil {
		return batchResult{featuresData: nil, columnTypes: nil, err: err}
	}

	var featuresData transTypes.ValueRows
	for _, data := range entityFeaturePairs {
		featuresData = append(featuresData, data.value)
	}

	if fc.cacheEnabled {
		if err := insertMultipleFeaturesToCache(fc.cache, entityFeaturePairs, fc.featureTableSpec.Project, fc.cacheTTL); err != nil {
			fc.logger.Error("insert_to_cache", zap.Any("error", err))
		}
	}

	return batchResult{featuresData: featuresData, columnTypes: entityFeaturePairs[0].columnTypes, err: nil}
}

func (fc *feastCall) buildFeastFeaturesData(ctx context.Context, feastResponse *feast.OnlineFeaturesResponse, projectName string, columns []string, entityIndexMap map[int]int) ([]entityFeaturePair, error) {
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
				if fc.valueMonitoringEnabled {
					v, err := getFloatValue(featVal)
					if err != nil {
						continue
					}
					feastFeatureSummary.WithLabelValues(column).Observe(v)
				}
			case serving.GetOnlineFeaturesResponse_NOT_FOUND, serving.GetOnlineFeaturesResponse_NULL_VALUE, serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE:
				defVal, ok := fc.defaultValues.GetDefaultValue(projectName, column)
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
			if fc.statusMonitoringEnabled {
				feastFeatureStatus.WithLabelValues(column, featureStatus.String()).Inc()
			}
		}
		data = append(data, entityFeaturePair{entity: entity, value: row, columnTypes: columnTypes})
	}

	return data, nil
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

	columns := getColumnNames(featureTableSpec)
	entityIndices := getEntityIndicesFromColumns(columns, featureTableSpec.Entities)

	numOfBatchBeforeCeil := float64(len(entityNotInCache)) / float64(fr.options.BatchSize)
	numOfBatch := int(math.Ceil(numOfBatchBeforeCeil))

	batchResultChan := make(chan batchResult, numOfBatch)

	for i := 0; i < numOfBatch; i++ {
		startIndex := i * fr.options.BatchSize
		endIndex := len(entityNotInCache)
		if endIndex > startIndex+fr.options.BatchSize {
			endIndex = startIndex + fr.options.BatchSize
		}
		batchedEntities := entityNotInCache[startIndex:endIndex]

		f, err := fr.newFeastCall(featureTableSpec, batchedEntities)
		if err != nil {
			return nil, err
		}

		hystrix.GoC(ctx, fr.options.FeastClientHystrixCommandName, func(ctx context.Context) error {
			batchResultChan <- f.do(ctx, features, columns, entityIndices)
			return nil
		}, func(ctx context.Context, err error) error {
			batchResultChan <- batchResult{featuresData: nil, err: err}
			return nil
		})
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
