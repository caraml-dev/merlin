package feast

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/types"
)

type CacheKey struct {
	Entity  feast.Row
	Project string
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

func fetchFeaturesFromCache(cache cache.Cache, entities []feast.Row, project string) (types.ValueRows, []feastTypes.ValueType_Enum, []feast.Row) {
	var entityNotInCache []feast.Row
	var featuresFromCache types.ValueRows
	var columnTypes []feastTypes.ValueType_Enum
	for _, entity := range entities {
		key := CacheKey{Entity: entity, Project: project}
		keyByte, err := json.Marshal(key)
		if err != nil {
			entityNotInCache = append(entityNotInCache, entity)
			continue
		}

		feastCacheRetrievalCount.Inc()
		val, err := cache.Fetch(keyByte)
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
		featuresFromCache = append(featuresFromCache, cacheValue.ValueRow)
		columnTypes = mergeColumnTypes(columnTypes, cacheValue.ValueTypes)
	}
	return featuresFromCache, columnTypes, entityNotInCache
}

func insertFeaturesToCache(cache cache.Cache, data entityFeaturePair, project string, ttl time.Duration) error {
	key := CacheKey{
		Entity:  data.entity,
		Project: project,
	}
	keyByte, err := json.Marshal(key)
	if err != nil {
		return err
	}

	value := CacheValue{
		ValueRow:   data.value,
		ValueTypes: data.columnTypes,
	}
	dataByte, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cache.Insert(keyByte, dataByte, ttl)
}

func insertMultipleFeaturesToCache(cache cache.Cache, cacheData []entityFeaturePair, project string, ttl time.Duration) error {
	var errorMsgs []string
	for _, data := range cacheData {
		if err := insertFeaturesToCache(cache, data, project, ttl); err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("(value: %v, with message: %v)", data.value, err.Error()))
		}
	}
	if len(errorMsgs) > 0 {
		compiledErrorMsgs := strings.Join(errorMsgs, ",")
		return fmt.Errorf("error inserting to cached: %s", compiledErrorMsgs)
	}
	return nil
}
