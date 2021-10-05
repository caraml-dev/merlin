package feast

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cespare/xxhash"
	feast "github.com/feast-dev/feast/sdk/go"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type featureCache struct {
	cache cache.Cache
	ttl   time.Duration
}

func newFeatureCache(ttl time.Duration, sizeInMB int) *featureCache {
	return &featureCache{
		cache: cache.NewInMemoryCache(sizeInMB),
		ttl:   ttl,
	}
}

type CacheKey struct {
	Entity         feast.Row
	Project        string
	ColumnNameHash uint64
}

type CacheValue struct {
	ValueRow   types.ValueRow
	ValueTypes []feastTypes.ValueType_Enum
}

var (
	feastCacheRetrievalCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_cache_retrieval_count",
		Help:      "Retrieve feature from cache",
	})

	feastCacheHitCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_cache_hit_count",
		Help:      "Cache is hitted",
	})
)

// fetchFeatureTable fetch features of several entities from cache scoped by its project and return it as a feature table
func (fc *featureCache) fetchFeatureTable(entities []feast.Row, columnNames []string, project string) (*internalFeatureTable, []feast.Row) {
	var entityNotInCache []feast.Row
	var entityInCache []feast.Row
	var featuresFromCache types.ValueRows

	// initialize empty value types
	columnTypes := make([]feastTypes.ValueType_Enum, len(columnNames))

	columnNameHash := computeHash(columnNames)
	for _, entity := range entities {
		key := CacheKey{Entity: entity, Project: project, ColumnNameHash: columnNameHash}
		keyByte, err := json.Marshal(key)
		if err != nil {
			entityNotInCache = append(entityNotInCache, entity)
			continue
		}

		feastCacheRetrievalCount.Inc()
		val, err := fc.cache.Fetch(keyByte)
		if err != nil {
			entityNotInCache = append(entityNotInCache, entity)
			continue
		}

		feastCacheHitCount.Inc()
		var cacheValue CacheValue
		if err := json.Unmarshal(val, &cacheValue); err != nil {
			entityNotInCache = append(entityNotInCache, entity)
			continue
		}

		columnTypes = mergeColumnTypes(columnTypes, cacheValue.ValueTypes)
		if cacheValue.ValueRow, err = castValueRow(cacheValue.ValueRow, columnTypes); err != nil {
			continue
		}

		entityInCache = append(entityInCache, entity)
		featuresFromCache = append(featuresFromCache, cacheValue.ValueRow)
	}

	return &internalFeatureTable{
		columnNames: columnNames,
		columnTypes: columnTypes,
		valueRows:   featuresFromCache,
		entities:    entityInCache,
	}, entityNotInCache
}

// insertFeatureTable insert a feature tables containing list of entities and their features into cache scoped by the project
func (fc *featureCache) insertFeatureTable(featureTable *internalFeatureTable, project string) error {
	var errorMsgs []string

	for idx, entity := range featureTable.entities {
		if err := fc.insertFeaturesOfEntity(entity, featureTable.columnNames, project, featureTable.valueRows[idx], featureTable.columnTypes); err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("(value: %v, with message: %v)", featureTable.valueRows[idx], err.Error()))
		}
	}
	if len(errorMsgs) > 0 {
		compiledErrorMsgs := strings.Join(errorMsgs, ",")
		return fmt.Errorf("error inserting to cached: %s", compiledErrorMsgs)
	}
	return nil
}

// insertFeaturesOfEntity insert features values of a given entity scoped by its project
func (fc *featureCache) insertFeaturesOfEntity(entity feast.Row, columnNames []string, project string, value types.ValueRow, valueTypes []feastTypes.ValueType_Enum) error {
	key := CacheKey{
		Entity:         entity,
		Project:        project,
		ColumnNameHash: computeHash(columnNames),
	}
	keyByte, err := json.Marshal(key)
	if err != nil {
		return err
	}

	cacheValue := CacheValue{
		ValueRow:   value,
		ValueTypes: valueTypes,
	}
	dataByte, err := json.Marshal(cacheValue)
	if err != nil {
		return err
	}
	return fc.cache.Insert(keyByte, dataByte, fc.ttl)
}

func castValueRow(row types.ValueRow, columnTypes []feastTypes.ValueType_Enum) (types.ValueRow, error) {
	for idx, val := range row {
		castedVal, err := castIntValue(val, columnTypes[idx])
		if err != nil {
			return row, err
		}
		row[idx] = castedVal
	}
	return row, nil
}

// castIntValue cast cache value of integer type to its correct type
// It's necessary since we store value as json and any integer value will be converted to float64
func castIntValue(val interface{}, valType feastTypes.ValueType_Enum) (interface{}, error) {
	switch valType {
	case feastTypes.ValueType_INT64:
		return converter.ToInt64(val)
	case feastTypes.ValueType_INT32:
		return converter.ToInt32(val)
	default:
		return val, nil
	}
}

func computeHash(columns []string) uint64 {
	return xxhash.Sum64String(strings.Join(columns, "-"))
}
