package feast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	mErrors "github.com/caraml-dev/merlin/pkg/errors"
	hystrixpkg "github.com/caraml-dev/merlin/pkg/hystrix"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	transTypes "github.com/caraml-dev/merlin/pkg/transformer/types"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/cespare/xxhash"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"go.uber.org/zap"
)

type StorageClient interface {
	GetOnlineFeatures(ctx context.Context, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error)
}

type (
	URL     string
	Clients map[spec.ServingSource]StorageClient
)

type FeatureRetriever interface {
	RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error)
	RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error)
}

// FeastRetriever is feature retriever implementation for retrieving features from Feast
type FeastRetriever struct {
	feastClients      Clients
	entityExtractor   *EntityExtractor
	featureCache      *featureCache
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
	logger *zap.Logger,
) *FeastRetriever {
	defaultValues := compileDefaultValues(featureTableSpecs)

	hystrix.ConfigureCommand(options.FeastClientHystrixCommandName, hystrix.CommandConfig{
		Timeout:                hystrixpkg.DurationToInt(options.FeastTimeout, time.Millisecond),
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
		logger:            logger,
	}
}

// Options for the Feast transformer.
type Options struct {
	// Configuration of feast storage for direct retrieval
	StorageConfigs FeastStorageConfig `envconfig:"FEAST_STORAGE_CONFIGS" required:"true"`
	// Default feast source, there are two options REDIS and BIGTABLE
	DefaultFeastSource spec.ServingSource `envconfig:"DEFAULT_FEAST_SOURCE" required:"true"`

	// Flag to emit metric of features status retrieved from feast
	StatusMonitoringEnabled bool `envconfig:"FEAST_FEATURE_STATUS_MONITORING_ENABLED" default:"false"`
	// Flat to emit metric of feature value retrieved from feast
	ValueMonitoringEnabled bool `envconfig:"FEAST_FEATURE_VALUE_MONITORING_ENABLED" default:"false"`
	// Number of entities in one batch of feast call
	BatchSize int `envconfig:"FEAST_BATCH_SIZE" default:"50"`
	// Flag to enable cache of feast retrieval result
	CacheEnabled bool `envconfig:"FEAST_CACHE_ENABLED" default:"true"`
	// Duration of cache will be lived and used as response
	CacheTTL time.Duration `envconfig:"FEAST_CACHE_TTL" default:"60s"`
	// Size of cache that can be store
	CacheSizeInMB int `envconfig:"CACHE_SIZE_IN_MB" default:"100"`

	// Timeout of feast request
	FeastTimeout time.Duration `envconfig:"FEAST_TIMEOUT" default:"1s"`
	// Name of feast call command
	FeastClientHystrixCommandName string `envconfig:"FEAST_HYSTRIX_COMMAND_NAME" default:"feast_retrieval"`
	// Maximum concurrent requests when call feast
	FeastClientMaxConcurrentRequests int `envconfig:"FEAST_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	// Threshold of error percentage, once breach circuit will be open
	FeastClientRequestVolumeThreshold int `envconfig:"FEAST_HYSTRIX_REQUEST_VOLUME_THRESHOLD" default:"100"`
	// Sleep window is duration of rejecting calling feast once the circuit is open
	FeastClientSleepWindow int `envconfig:"FEAST_HYSTRIX_SLEEP_WINDOW" default:"1000"` // How long, in milliseconds, to wait after a circuit opens before testing for recovery
	// Threshold of number of request to model predictor
	FeastClientErrorPercentThreshold int `envconfig:"FEAST_HYSTRIX_ERROR_PERCENT_THRESHOLD" default:"25"`

	// Flag to enable feast keep alive
	FeastServingKeepAliveEnabled bool `envconfig:"FEAST_SERVING_KEEP_ALIVE_ENABLED" default:"false"`
	// Duration of interval between keep alive PING
	FeastServingKeepAliveTime time.Duration `envconfig:"FEAST_SERVING_KEEP_ALIVE_TIME" default:"60s"`
	// Duration of PING that considered as TIMEOUT
	FeastServingKeepAliveTimeout time.Duration `envconfig:"FEAST_SERVING_KEEP_ALIVE_TIMEOUT" default:"5s"`
	// Number of feast gRPC connection
	FeastGRPCConnCount int `envconfig:"FEAST_GRPC_CONN_COUNT" default:"10"`
}

func (fr *FeastRetriever) RetrieveFeatureOfEntityInRequest(ctx context.Context, requestJson transTypes.JSONObject) ([]*transTypes.FeatureTable, error) {
	ctx, span := tracer.Start(ctx, "feast.RetrieveFromRequest")
	defer span.End()

	sr := symbol.NewRegistryWithCompiledJSONPath(fr.entityExtractor.compiledJsonPath)
	sr.SetRawRequest(requestJson)

	return fr.RetrieveFeatureOfEntityInSymbolRegistry(ctx, sr)
}

func (fr *FeastRetriever) RetrieveFeatureOfEntityInSymbolRegistry(ctx context.Context, symbolRegistry symbol.Registry) ([]*transTypes.FeatureTable, error) {
	ctx, span := tracer.Start(ctx, "feast.RetrieveFromSymbolRegistry")
	defer span.End()

	// parallelize feast call per feature table
	resChan := make(chan callResult, len(fr.featureTableSpecs))
	for _, featureTableSpec := range fr.featureTableSpecs {
		go func(featureTableSpec *spec.FeatureTable) {
			featureTable, err := fr.getFeaturePerTable(ctx, symbolRegistry, featureTableSpec)
			resChan <- callResult{tableName: GetTableName(featureTableSpec), featureTable: featureTable, err: err}
		}(featureTableSpec)
	}

	// collect result
	feastFeatures := make([]*transTypes.FeatureTable, len(fr.featureTableSpecs))
	for i := 0; i < cap(resChan); i++ {
		res := <-resChan
		if res.err != nil {
			// cancel all other goroutine if one of the goroutine returning error.
			ctx.Done()
			return nil, res.err
		}
		feastFeatures[i] = res.featureTable.toFeatureTable(res.tableName)
	}

	return feastFeatures, nil
}

func (fr *FeastRetriever) getFeaturePerTable(ctx context.Context, symbolRegistry symbol.Registry, featureTableSpec *spec.FeatureTable) (*internalFeatureTable, error) {
	ctx, span := tracer.Start(ctx, "feast.getFeatureTable")
	span.SetAttributes(attribute.String("table.Name", GetTableName(featureTableSpec)))
	defer span.End()

	entities, err := fr.buildEntityRows(symbolRegistry, featureTableSpec.Entities)
	if err != nil {
		return nil, err
	}

	featureTable, err := fr.getFeatureTable(ctx, entities, featureTableSpec)
	if err != nil {
		return nil, err
	}
	return featureTable, nil
}

func (fr *FeastRetriever) buildEntityRows(symbolRegistry symbol.Registry, configEntities []*spec.Entity) ([]feast.Row, error) {
	var allSeries [][]*types.Value
	maxLength := 1

	for k, configEntity := range configEntities {
		vals, err := fr.entityExtractor.ExtractValuesFromSymbolRegistry(symbolRegistry, configEntity)
		if err != nil {
			return nil, fmt.Errorf("unable to extract entity %s: %w", configEntity.Name, err)
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

func (fr *FeastRetriever) getFeastClient(source spec.ServingSource) (StorageClient, error) {
	servingSource := source
	if servingSource == spec.ServingSource_UNKNOWN {
		servingSource = fr.options.DefaultFeastSource
	}

	client, ok := fr.feastClients[servingSource]
	if ok {
		return client, nil
	}

	return nil, errors.New("invalid feast serving url")
}

func (fr *FeastRetriever) getFeatureTable(ctx context.Context, entities []feast.Row, featureTableSpec *spec.FeatureTable) (*internalFeatureTable, error) {
	features := getFeatureNames(featureTableSpec)
	columns := getColumnNames(featureTableSpec)
	entitySet := getEntitySet(columns, featureTableSpec.Entities)

	var featureTable *internalFeatureTable
	entityNotInCache := entities
	if fr.options.CacheEnabled {
		featureTable, entityNotInCache = fr.featureCache.fetchFeatureTable(entities, columns, featureTableSpec.Project)
	}

	numOfBatchBeforeCeil := float64(len(entityNotInCache)) / float64(fr.options.BatchSize)
	numOfBatch := int(math.Ceil(numOfBatchBeforeCeil))
	batchResultChan := make(chan callResult, numOfBatch)
	fallbackChan := make(chan callResult, numOfBatch)

	for i := 0; i < numOfBatch; i++ {
		startIndex := i * fr.options.BatchSize
		endIndex := len(entityNotInCache)
		if endIndex > startIndex+fr.options.BatchSize {
			endIndex = startIndex + fr.options.BatchSize
		}
		batchedEntities := entityNotInCache[startIndex:endIndex]

		f, err := newCall(fr, featureTableSpec, columns, entitySet)
		if err != nil {
			return nil, err
		}

		hystrix.GoC(ctx, fr.options.FeastClientHystrixCommandName, func(ctx context.Context) error {
			batchResultChan <- f.do(ctx, batchedEntities, features)
			return nil
		}, func(ctx context.Context, err error) error {
			fallbackChan <- callResult{featureTable: nil, err: err}
			return nil
		})
	}

	// merge result from all batch, including the cached one
	for i := 0; i < numOfBatch; i++ {
		select {
		case res := <-batchResultChan:
			if res.err != nil {
				fr.logger.Error("feast retrieval error", zap.Any("error", res.err))
				return nil, handleFeastError(res.err)
			}

			if fr.options.CacheEnabled {
				if err := fr.featureCache.insertFeatureTable(res.featureTable, featureTableSpec.Project); err != nil {
					fr.logger.Error("insert_to_cache", zap.Any("error", err))
				}
			}

			// first result received
			if featureTable == nil {
				featureTable = res.featureTable
				continue
			}

			// merge subsequent result to featureTable
			err := featureTable.mergeFeatureTable(res.featureTable)
			if err != nil {
				return nil, err
			}

		case res := <-fallbackChan:
			fr.logger.Error("fallback error", zap.Any("error", res.err))
			return nil, handleFeastError(res.err)
		}
	}

	return featureTable, nil
}

func handleFeastError(err error) error {
	if errors.Is(err, hystrix.ErrTimeout) {
		return mErrors.NewDeadlineExceededError(err.Error())
	}
	if grpcErr, valid := status.FromError(err); valid && grpcErr.Code() == codes.DeadlineExceeded {
		return mErrors.NewDeadlineExceededError(grpcErr.Message())
	}
	return err
}

// getFeatureNames get list of feature name within a feature table spec
func getFeatureNames(config *spec.FeatureTable) []string {
	features := make([]string, len(config.Features))
	for idx, feature := range config.Features {
		features[idx] = feature.Name
	}
	return features
}

// getColumnNames get list of feature and entity name within a feature table spec
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

func getEntitySet(columns []string, entitiesConfig []*spec.Entity) map[string]bool {
	entitySet := make(map[string]bool, len(entitiesConfig))
	entitiesConfigMap := make(map[string]*spec.Entity)
	for _, entityConfig := range entitiesConfig {
		entitiesConfigMap[entityConfig.Name] = entityConfig
	}
	for _, column := range columns {
		if _, found := entitiesConfigMap[column]; found {
			entitySet[column] = true
		}
	}
	return entitySet
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
