package redis

import (
	"context"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/go-redis/redis/v8"
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
		OnConnect: func(_ context.Context, _ *redis.Conn) error {
			redisNewConn.WithLabelValues().Inc()
			return nil
		},
	})

	redisClient.AddHook(&redisHook{})
	go recordRedisConnMetric(redisClient, nil)

	return newClient(redisClient.Pipeline(), featureTablesMetadata)
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
		OnConnect: func(_ context.Context, _ *redis.Conn) error {
			redisNewConn.WithLabelValues().Inc()
			return nil
		},
	})

	redisClient.AddHook(&redisHook{})
	go recordRedisConnMetric(nil, redisClient)

	return newClient(redisClient.Pipeline(), featureTablesMetadata)
}

func recordRedisConnMetric(client *redis.Client, clusterClient *redis.ClusterClient) {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var poolStats *redis.PoolStats
			if client != nil {
				poolStats = client.PoolStats()
			} else {
				poolStats = clusterClient.PoolStats()
			}

			if poolStats == nil {
				continue
			}

			redisConnPoolStats.WithLabelValues(hitConnStats).Set(float64(poolStats.Hits))
			redisConnPoolStats.WithLabelValues(missConnStats).Set(float64(poolStats.Misses))
			redisConnPoolStats.WithLabelValues(timeoutConnStats).Set(float64(poolStats.Timeouts))
			redisConnPoolStats.WithLabelValues(idleConnStats).Set(float64(poolStats.IdleConns))
			redisConnPoolStats.WithLabelValues(staleConnStats).Set(float64(poolStats.StaleConns))
			redisConnPoolStats.WithLabelValues(totalConnStats).Set(float64(poolStats.TotalConns))
		}
	}
}

func newClient(pipeliner redis.Pipeliner, featureTablesMetadata []*spec.FeatureTableMetadata) (RedisClient, error) {
	return RedisClient{
		encoder:   newRedisEncoder(featureTablesMetadata),
		pipeliner: pipeliner,
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
	pipeline := r.pipeliner

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
