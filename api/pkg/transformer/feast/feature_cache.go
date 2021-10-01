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
	Entity          feast.Row
	Project         string
	FeatureNameHash uint64
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
func (fc *featureCache) fetchFeatureTable(entities []feast.Row, featureNames []string, project string) (*internalFeatureTable, []feast.Row) {
	var entityNotInCache []feast.Row
	var entityInCache []feast.Row
	var featuresFromCache types.ValueRows

	// initialize empty value types
	columnTypes := make([]feastTypes.ValueType_Enum, len(featureNames))

	hashedFeatureNames := computeHash(featureNames)
	for _, entity := range entities {
		key := CacheKey{Entity: entity, Project: project, FeatureNameHash: hashedFeatureNames}
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

		entityInCache = append(entityInCache, entity)
		featuresFromCache = append(featuresFromCache, cacheValue.ValueRow)
		columnTypes = mergeColumnTypes(columnTypes, cacheValue.ValueTypes)
	}

	return &internalFeatureTable{
		columnNames: featureNames,
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
func (fc *featureCache) insertFeaturesOfEntity(entity feast.Row, featureNames []string, project string, value types.ValueRow, valueTypes []feastTypes.ValueType_Enum) error {
	key := CacheKey{
		Entity:          entity,
		Project:         project,
		FeatureNameHash: computeHash(featureNames),
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

func computeHash(names []string) uint64 {
	return xxhash.Sum64String(strings.Join(names, "-"))
}
