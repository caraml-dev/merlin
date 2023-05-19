package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	internalValidator "github.com/caraml-dev/merlin/pkg/validator"
	"github.com/go-playground/validator"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		timeDuration := time.Duration(value)
		*d = Duration(timeDuration)
		return nil
	case string:
		var err error
		timeDuration, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(timeDuration)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// FeastRedisConfig is redis storage configuration for standard transformer whether it is single or cluster
type FeastRedisConfig struct {
	// Flag to indicate whether redis configuration is for single redis or redis cluster
	IsRedisCluster bool `json:"is_redis_cluster"`
	// Flag to indicate whether feast serving using direct REDIS storage or using feast grpc
	IsUsingDirectStorage bool `json:"is_using_direct_storage"`
	// Feast GRPC serving URL that use redis
	ServingURL string `json:"serving_url" validate:"required"`
	// Address of redis
	RedisAddresses []string `json:"redis_addresses" validate:"required,gt=0"`
	// Number of maximum pool size of redis connections
	PoolSize int32 `json:"pool_size" validate:"gt=0"`
	// Max retries if redis command returning error
	MaxRetries int32 `json:"max_retries"`
	// Backoff duration before attempt retry
	MinRetryBackoff *time.Duration `json:"min_retry_backoff"`
	// Maximum duration before timeout to establishing new redis connection
	DialTimeout *time.Duration `json:"dial_timeout"`
	// Maximum duration before timeout to read from redis
	ReadTimeout *time.Duration `json:"read_timeout"`
	// Maximum duration before timeout to write to redis
	WriteTimeout *time.Duration `json:"write_timeout"`
	// Maximum age of a connection before mark it as stale and destroy the connection
	MaxConnAge *time.Duration `json:"max_conn_age"`
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	PoolTimeout *time.Duration `json:"pool_timeout"`
	// Amount of time after which client closes idle connections
	IdleTimeout *time.Duration `json:"idle_timeout"`
	// Frequency of idle checks
	IdleCheckFrequency *time.Duration `json:"idle_check_frequency"`
	// Minimum number of idle connections
	MinIdleConn int32 `json:"min_idle_conn"`
}

func (redisCfg *FeastRedisConfig) Decode(value string) error {
	var cfg FeastRedisConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return err
	}
	validate, _ := internalValidator.NewValidator()
	if err := validate.Struct(cfg); err != nil {
		translatedErr := err.(validator.ValidationErrors)[0].Translate(internalValidator.EN) // nolint:errorlint
		return fmt.Errorf(translatedErr)
	}
	*redisCfg = cfg
	return nil
}

func (redisCfg *FeastRedisConfig) ToFeastStorage() *spec.OnlineStorage {
	servingType := spec.ServingType_FEAST_GRPC
	if redisCfg.IsUsingDirectStorage {
		servingType = spec.ServingType_DIRECT_STORAGE
	}

	redisOpts := &spec.RedisOption{
		PoolSize:           redisCfg.PoolSize,
		MaxRetries:         redisCfg.MaxRetries,
		MinRetryBackoff:    getDurationProtoFromTime(redisCfg.MinRetryBackoff),
		DialTimeout:        getDurationProtoFromTime(redisCfg.DialTimeout),
		ReadTimeout:        getDurationProtoFromTime(redisCfg.ReadTimeout),
		WriteTimeout:       getDurationProtoFromTime(redisCfg.WriteTimeout),
		MaxConnAge:         getDurationProtoFromTime(redisCfg.MaxConnAge),
		PoolTimeout:        getDurationProtoFromTime(redisCfg.PoolTimeout),
		IdleTimeout:        getDurationProtoFromTime(redisCfg.IdleTimeout),
		IdleCheckFrequency: getDurationProtoFromTime(redisCfg.IdleCheckFrequency),
		MinIdleConnections: redisCfg.MinIdleConn,
	}
	if redisCfg.IsRedisCluster {
		return &spec.OnlineStorage{
			ServingType: servingType,
			Storage: &spec.OnlineStorage_RedisCluster{
				RedisCluster: &spec.RedisClusterStorage{
					FeastServingUrl: redisCfg.ServingURL,
					RedisAddress:    redisCfg.RedisAddresses,
					Option:          redisOpts,
				},
			},
		}
	}

	return &spec.OnlineStorage{
		ServingType: servingType,
		Storage: &spec.OnlineStorage_Redis{
			Redis: &spec.RedisStorage{
				FeastServingUrl: redisCfg.ServingURL,
				RedisAddress:    redisCfg.RedisAddresses[0],
				Option:          redisOpts,
			},
		},
	}
}

// FeastBigtableConfig is bigtable storage configuration for standard transformer
type FeastBigtableConfig struct {
	// Flag to indicate whether feast serving using direct BIGTABLE storage or using feast grpc
	IsUsingDirectStorage bool `json:"is_using_direct_storage" default:"false"`
	// Feast GRPC serving URL that use bigtable
	ServingURL string `json:"serving_url" validate:"required"`
	// GCP Project where the bigtable instance is located
	Project string `json:"project" validate:"required"`
	// Name of bigtable instance
	Instance string `json:"instance" validate:"required"`
	// Bigtable app profile
	AppProfile string `json:"app_profile"`
	// Number of grpc connnection created
	PoolSize int32 `json:"pool_size" validate:"gt=0"`
	// Interval to send keep-alive packet to established connection
	KeepAliveInterval *time.Duration `json:"keep_alive_interval"`
	// Duration before connection is marks as not healthy and close the connection afterward
	KeepAliveTimeout *time.Duration `json:"keep_alive_timeout"`
}

func (bigtableCfg *FeastBigtableConfig) Decode(value string) error {
	var cfg FeastBigtableConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return err
	}
	validate, _ := internalValidator.NewValidator()
	if err := validate.Struct(cfg); err != nil {
		translatedErr := err.(validator.ValidationErrors)[0].Translate(internalValidator.EN) // nolint:errorlint
		return fmt.Errorf(translatedErr)
	}
	*bigtableCfg = cfg
	return nil
}

func (bigtableCfg *FeastBigtableConfig) ToFeastStorageWithCredential(credential string) *spec.OnlineStorage {
	servingType := spec.ServingType_FEAST_GRPC
	if bigtableCfg.IsUsingDirectStorage {
		servingType = spec.ServingType_DIRECT_STORAGE
	}

	return &spec.OnlineStorage{
		ServingType: servingType,
		Storage: &spec.OnlineStorage_Bigtable{
			Bigtable: &spec.BigTableStorage{
				FeastServingUrl: bigtableCfg.ServingURL,
				Project:         bigtableCfg.Project,
				Instance:        bigtableCfg.Instance,
				AppProfile:      bigtableCfg.AppProfile,
				Option: &spec.BigTableOption{
					GrpcConnectionPool: bigtableCfg.PoolSize,
					KeepAliveInterval:  getDurationProtoFromTime(bigtableCfg.KeepAliveInterval),
					KeepAliveTimeout:   getDurationProtoFromTime(bigtableCfg.KeepAliveTimeout),
					CredentialJson:     credential,
				},
			},
		},
	}
}

func getDurationProtoFromTime(timeDuration *time.Duration) *durationpb.Duration {
	if timeDuration == nil {
		return nil
	}
	return durationpb.New(*timeDuration)
}
