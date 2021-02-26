package feast

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
)

func fetchFeaturesFromCache(cache Cache, entities []feast.Row) (FeaturesData, []feast.Row) {
	var entityNotInCache []feast.Row
	var featuresFromCache FeaturesData
	for _, entity := range entities {
		keyByte, err := json.Marshal(entity)
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
		var cacheData FeatureData
		if err := json.Unmarshal(val, &cacheData); err != nil {
			entityNotInCache = append(entityNotInCache, entity)
			continue
		}
		featuresFromCache = append(featuresFromCache, cacheData)
	}
	return featuresFromCache, entityNotInCache
}

func insertFeaturesToCache(cache Cache, data entityFeaturePair, ttl time.Duration) error {
	keyByte, err := json.Marshal(data.key)
	if err != nil {
		return err
	}
	dataByte, err := json.Marshal(data.value)
	if err != nil {
		return err
	}
	return cache.Insert(keyByte, dataByte, ttl)
}

func insertMultipleFeaturesToCache(cache Cache, cacheData []entityFeaturePair, ttl time.Duration) error {
	var errorMsgs []string
	for _, data := range cacheData {
		if err := insertFeaturesToCache(cache, data, ttl); err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("(value: %v, with message: %v)", data.value, err.Error()))
		}
	}
	if len(errorMsgs) > 0 {
		compiledErrorMsgs := strings.Join(errorMsgs, ",")
		return fmt.Errorf("error inserting to cached: %s", compiledErrorMsgs)
	}
	return nil
}
