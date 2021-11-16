package feast

import (
	"time"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"google.golang.org/protobuf/types/known/durationpb"
)

type FeastStorageConfig map[spec.ServingSource]*spec.OnlineStorage

// RedisOverwriteConfig is redis configuration that will overwrite existing storage config that specify by merlin API
type RedisOverwriteConfig struct {
	RedisDirectStorageEnabled *bool          `envconfig:"FEAST_REDIS_DIRECT_STORAGE_ENABLED"`
	PoolSize                  *int32         `envconfig:"FEAST_REDIS_POOL_SIZE"`
	ReadTimeout               *time.Duration `envconfig:"FEAST_REDIS_READ_TIMEOUT"`
	WriteTimeout              *time.Duration `envconfig:"FEAST_REDIS_WRITE_TIMEOUT"`
}

// overwrite feast options with values that specified by user through transformer environment variables
func OverwriteFeastOptionsConfig(opts Options, redisConfig RedisOverwriteConfig) Options {
	feastOpts := opts
	for _, storage := range opts.StorageConfigs {
		switch storage.Storage.(type) {
		case *spec.OnlineStorage_Redis:
			storage.ServingType = getServingType(redisConfig.RedisDirectStorageEnabled, storage.ServingType)
			redisStorage := storage.GetRedis()
			overwriteRedisOption(redisStorage.Option, redisConfig)
		case *spec.OnlineStorage_RedisCluster:
			storage.ServingType = getServingType(redisConfig.RedisDirectStorageEnabled, storage.ServingType)
			redisClusterStorage := storage.GetRedisCluster()
			overwriteRedisOption(redisClusterStorage.Option, redisConfig)
		}
	}
	return feastOpts
}

func getServingType(userEnableDirectStorage *bool, currentServingType spec.ServingType) spec.ServingType {
	if userEnableDirectStorage == nil {
		return currentServingType
	}
	if *userEnableDirectStorage {
		return spec.ServingType_DIRECT_STORAGE
	}
	return spec.ServingType_FEAST_GRPC
}

func overwriteRedisOption(opts *spec.RedisOption, redisOverwriteConfig RedisOverwriteConfig) {
	if redisOverwriteConfig.PoolSize != nil {
		opts.PoolSize = *redisOverwriteConfig.PoolSize
	}
	if redisOverwriteConfig.ReadTimeout != nil {
		opts.ReadTimeout = durationpb.New(*redisOverwriteConfig.ReadTimeout)
	}
	if redisOverwriteConfig.WriteTimeout != nil {
		opts.WriteTimeout = durationpb.New(*redisOverwriteConfig.WriteTimeout)
	}
}
