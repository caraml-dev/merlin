package feast

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
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

func fetchFeaturesFromCache(cache cache.Cache, entities []feast.Row, project string) (types.ValueRows, []feast.Row) {
	var entityNotInCache []feast.Row
	var featuresFromCache types.ValueRows
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
		var cacheData types.ValueRow
		if err := json.Unmarshal(val, &cacheData); err != nil {
			entityNotInCache = append(entityNotInCache, entity)
			continue
		}
		featuresFromCache = append(featuresFromCache, cacheData)
	}
	return featuresFromCache, entityNotInCache
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

	dataByte, err := json.Marshal(data.value)
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
