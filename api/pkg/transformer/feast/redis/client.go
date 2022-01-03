package redis

import (
	"context"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/go-redis/redis/v8"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"google.golang.org/protobuf/types/known/durationpb"
)

type RedisEncoder struct {
	specs map[string]*spec.FeatureTableMetadata
}

type RedisClient struct {
	encoder   RedisEncoder
	pipeliner redis.Pipeliner
}

func NewRedisClient(redisStorage *spec.RedisStorage, featureTablesMetadata []*spec.FeatureTableMetadata) (RedisClient, error) {
	option := redisStorage.Option
	redisClient := redis.NewClient(&redis.Options{
		Addr:               redisStorage.GetRedisAddress(),
		MaxRetries:         int(option.MaxRetries),
		MinRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
		MaxRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
		DialTimeout:        getNullableDuration(option.DialTimeout),
		ReadTimeout:        getNullableDuration(option.ReadTimeout),
		WriteTimeout:       getNullableDuration(option.WriteTimeout),
		PoolSize:           int(option.PoolSize),
		MaxConnAge:         getNullableDuration(option.MaxConnAge),
		PoolTimeout:        getNullableDuration(option.PoolTimeout),
		IdleTimeout:        getNullableDuration(option.IdleTimeout),
		IdleCheckFrequency: getNullableDuration(option.IdleCheckFrequency),
		MinIdleConns:       int(option.MinIdleConnections),
	})

	return RedisClient{
		encoder:   newRedisEncoder(featureTablesMetadata),
		pipeliner: redisClient.Pipeline(),
	}, nil
}

func NewRedisClusterClient(redisClusterStorage *spec.RedisClusterStorage, featureTablesMetadata []*spec.FeatureTableMetadata) (RedisClient, error) {
	option := redisClusterStorage.Option
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              redisClusterStorage.GetRedisAddress(),
		MaxRetries:         int(option.MaxRetries),
		MinRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
		MaxRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
		DialTimeout:        getNullableDuration(option.DialTimeout),
		ReadTimeout:        getNullableDuration(option.ReadTimeout),
		WriteTimeout:       getNullableDuration(option.WriteTimeout),
		PoolSize:           int(option.PoolSize),
		MaxConnAge:         getNullableDuration(option.MaxConnAge),
		PoolTimeout:        getNullableDuration(option.PoolTimeout),
		IdleTimeout:        getNullableDuration(option.IdleTimeout),
		IdleCheckFrequency: getNullableDuration(option.IdleCheckFrequency),
		MinIdleConns:       int(option.MinIdleConnections),
	})

	return RedisClient{
		encoder:   newRedisEncoder(featureTablesMetadata),
		pipeliner: redisClient.Pipeline(),
	}, nil
}

func (r RedisClient) GetOnlineFeatures(ctx context.Context, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	encodedFeatureRequest, err := r.encoder.EncodeFeatureRequest(req)
	encodedEntities := encodedFeatureRequest.EncodedEntities
	encodedFeatures := encodedFeatureRequest.EncodedFeatures
	if err != nil {
		return nil, err
	}
	hmGetResults := make([]*redis.SliceCmd, len(encodedEntities))
	pipeline := r.pipeliner.Pipeline()
	for index, encodedEntity := range encodedEntities {
		hmGetResults[index] = pipeline.HMGet(ctx, encodedEntity, encodedFeatures...)
	}
	_, err = pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	redisHashMaps := make([][]interface{}, len(hmGetResults))
	for index, result := range hmGetResults {
		redisHashMap, err := result.Result()
		if err != nil {
			return nil, err
		}
		redisHashMaps[index] = redisHashMap
	}
	response, err := r.encoder.DecodeStoredRedisValue(redisHashMaps, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getNullableDuration(duration *durationpb.Duration) time.Duration {
	if duration == nil {
		return 0
	}
	return duration.AsDuration()
}
