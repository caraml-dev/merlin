package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/gojek/merlin/pkg/transformer/spec"
	internalValidator "github.com/gojek/merlin/pkg/validator"
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
	IsRedisCluster       bool      `json:"is_redis_cluster"`
	IsUsingDirectStorage bool      `json:"is_using_direct_storage"`
	ServingURL           string    `json:"serving_url" validate:"required"`
	RedisAddresses       []string  `json:"redis_addresses" validate:"required,gt=0"`
	PoolSize             int32     `json:"pool_size" validate:"gt=0"`
	MaxRetries           int32     `json:"max_retries"`
	MinRetryBackoff      *Duration `json:"min_retry_backoff"`
	DialTimeout          *Duration `json:"dial_timeout"`
	ReadTimeout          *Duration `json:"read_timeout"`
	WriteTimeout         *Duration `json:"write_timeout"`
	MaxConnAge           *Duration `json:"max_conn_age"`
	PoolTimeout          *Duration `json:"pool_timeout"`
	IdleTimeout          *Duration `json:"idle_timeout"`
	IdleCheckFrequency   *Duration `json:"idle_check_frequency"`
	MinIdleConn          int32     `json:"min_idle_conn"`
}

func (redisCfg *FeastRedisConfig) Decode(value string) error {
	var cfg FeastRedisConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return err
	}
	validate := internalValidator.NewValidator()
	if err := validate.Struct(cfg); err != nil {
		translatedErr := err.(validator.ValidationErrors)[0].Translate(internalValidator.EN)
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
	IsUsingDirectStorage bool      `json:"is_using_direct_storage" default:"false"`
	ServingURL           string    `json:"serving_url" validate:"required"`
	Project              string    `json:"project" validate:"required"`
	Instance             string    `json:"instance" validate:"required"`
	AppProfile           string    `json:"app_profile"`
	PoolSize             int32     `json:"pool_size" validate:"gt=0"`
	KeepAliveInterval    *Duration `json:"keep_alive_interval"`
	KeepAliveTimeout     *Duration `json:"keep_alive_timeout"`
}

func (bigtableCfg *FeastBigtableConfig) Decode(value string) error {
	var cfg FeastBigtableConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return err
	}
	validate := internalValidator.NewValidator()
	if err := validate.Struct(cfg); err != nil {
		translatedErr := err.(validator.ValidationErrors)[0].Translate(internalValidator.EN)
		return fmt.Errorf(translatedErr)
	}
	*bigtableCfg = cfg
	return nil
}

func (bigtableCfg *FeastBigtableConfig) ToFeastStorage() *spec.OnlineStorage {
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
				},
			},
		},
	}
}

func getDurationProtoFromTime(timeDuration *Duration) *durationpb.Duration {
	if timeDuration == nil {
		return nil
	}
	return durationpb.New(time.Duration(*timeDuration))
}
