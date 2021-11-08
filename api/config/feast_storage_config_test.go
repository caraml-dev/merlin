package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestFeastRedisConfig(t *testing.T) {
	tenSecDuration := Duration(time.Second * 10)
	testCases := []struct {
		desc                string
		redisConfigString   string
		expectedRedisConfig *FeastRedisConfig
		err                 error
	}{
		{
			desc:              "Success: valid redisconfig",
			redisConfigString: `{"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`,
			expectedRedisConfig: &FeastRedisConfig{
				ServingURL: "online-storage.merlin.dev",
				RedisAddresses: []string{
					"10.1.1.10", "10.1.1.11",
				},
				PoolSize:    4,
				MaxRetries:  1,
				DialTimeout: &tenSecDuration,
			},
		},
		{
			desc:              "Invalid: pool_size 0",
			redisConfigString: `{"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 0,"max_retries": 1,"dial_timeout": "10s"}`,
			err:               fmt.Errorf(`envconfig.Process: assigning FEAST_REDIS_CONFIG to FeastRedisConfig: converting '{"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 0,"max_retries": 1,"dial_timeout": "10s"}' to type config.FeastRedisConfig. details: Key: 'FeastRedisConfig.PoolSize' Error:Field validation for 'PoolSize' failed on the 'gt' tag`),
		},
		{
			desc:              "Invalid: serving_url is not ser",
			redisConfigString: `{"serving_url":"","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`,
			err:               fmt.Errorf(`envconfig.Process: assigning FEAST_REDIS_CONFIG to FeastRedisConfig: converting '{"serving_url":"","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}' to type config.FeastRedisConfig. details: ServingURL is required`),
		},
		{
			desc:              "Invalid: redis_addresses is not ser",
			redisConfigString: `{"serving_url":"online-storage.merlin.dev","redis_addresses":[],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`,
			err:               fmt.Errorf(`envconfig.Process: assigning FEAST_REDIS_CONFIG to FeastRedisConfig: converting '{"serving_url":"online-storage.merlin.dev","redis_addresses":[],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}' to type config.FeastRedisConfig. details: Key: 'FeastRedisConfig.RedisAddresses' Error:Field validation for 'RedisAddresses' failed on the 'gt' tag`),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			os.Setenv("FEAST_REDIS_CONFIG", tC.redisConfigString)
			var cfg StandardTransformerConfig
			err := envconfig.Process("", &cfg)
			if err == nil {
				assert.Equal(t, tC.expectedRedisConfig, cfg.FeastRedisConfig)
			} else {
				assert.EqualError(t, err, tC.err.Error())
			}
		})
	}
}

func TestFeastBigtableConfig(t *testing.T) {
	testCases := []struct {
		desc                   string
		bigtableString         string
		expectedBigTableConfig *FeastBigTableConfig
		err                    error
	}{
		{
			desc:           "Success: valid env variable",
			bigtableString: `{"serving_url":"10.1.1.3"}`,
			expectedBigTableConfig: &FeastBigTableConfig{
				ServingURL: "10.1.1.3",
			},
		},
		{
			desc:           "Fail: serving_url is not set",
			bigtableString: `{"serving_url":""}`,
			err:            fmt.Errorf(`envconfig.Process: assigning FEAST_BIG_TABLE_CONFIG to FeastBigTableConfig: converting '{"serving_url":""}' to type config.FeastBigTableConfig. details: ServingURL is required`),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			os.Setenv("FEAST_BIG_TABLE_CONFIG", tC.bigtableString)
			var cfg StandardTransformerConfig
			err := envconfig.Process("", &cfg)
			if err == nil {
				assert.Equal(t, tC.expectedBigTableConfig, cfg.FeastBigTableConfig)
			} else {
				assert.EqualError(t, err, tC.err.Error())
			}
		})
	}
}

func TestRedisConfig_ToFeastStorage(t *testing.T) {
	timeDuration := Duration(time.Second * 20)
	testCases := []struct {
		desc                 string
		redisConfig          *FeastRedisConfig
		expectedFeastStorage *spec.OnlineStorage
	}{
		{
			desc: "Complete",
			redisConfig: &FeastRedisConfig{
				IsUsingDirectStorage: true,
				ServingURL:           "localhost",
				RedisAddresses: []string{
					"10.1.1.2", "10.1.1.3",
				},
				PoolSize:           5,
				MaxRetries:         0,
				MinRetryBackoff:    &timeDuration,
				DialTimeout:        &timeDuration,
				ReadTimeout:        &timeDuration,
				WriteTimeout:       &timeDuration,
				MaxConnAge:         &timeDuration,
				PoolTimeout:        &timeDuration,
				IdleTimeout:        &timeDuration,
				IdleCheckFrequency: &timeDuration,
			},
			expectedFeastStorage: &spec.OnlineStorage{
				ServingType: spec.ServingType_DIRECT_STORAGE,
				Storage: &spec.OnlineStorage_RedisCluster{
					RedisCluster: &spec.RedisClusterStorage{
						FeastServingUrl: "localhost",
						RedisAddress: []string{
							"10.1.1.2", "10.1.1.3",
						},
						Option: &spec.RedisOption{
							MaxRetries:         0,
							PoolSize:           5,
							MinRetryBackoff:    durationpb.New(time.Duration(timeDuration)),
							DialTimeout:        durationpb.New(time.Duration(timeDuration)),
							ReadTimeout:        durationpb.New(time.Duration(timeDuration)),
							WriteTimeout:       durationpb.New(time.Duration(timeDuration)),
							MaxConnAge:         durationpb.New(time.Duration(timeDuration)),
							PoolTimeout:        durationpb.New(time.Duration(timeDuration)),
							IdleTimeout:        durationpb.New(time.Duration(timeDuration)),
							IdleCheckFrequency: durationpb.New(time.Duration(timeDuration)),
						},
					},
				},
			},
		},
		{
			desc: "Without set time related option",
			redisConfig: &FeastRedisConfig{
				ServingURL: "localhost",
				RedisAddresses: []string{
					"10.1.1.2", "10.1.1.3",
				},
				PoolSize:   5,
				MaxRetries: 2,
			},
			expectedFeastStorage: &spec.OnlineStorage{
				Storage: &spec.OnlineStorage_RedisCluster{
					RedisCluster: &spec.RedisClusterStorage{
						FeastServingUrl: "localhost",
						RedisAddress: []string{
							"10.1.1.2", "10.1.1.3",
						},
						Option: &spec.RedisOption{
							MaxRetries: 2,
							PoolSize:   5,
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.redisConfig.ToFeastStorage()
			assert.Equal(t, tC.expectedFeastStorage, got)
		})
	}
}

func TestBigtableConfig_ToFeastConfig(t *testing.T) {
	testCases := []struct {
		desc                   string
		bigtableConfig         *FeastBigTableConfig
		expectedBigTableConfig *spec.OnlineStorage
	}{
		{
			desc: "Success",
			bigtableConfig: &FeastBigTableConfig{
				IsUsingDirectStorage: true,
				ServingURL:           "localhost",
			},
			expectedBigTableConfig: &spec.OnlineStorage{
				ServingType: spec.ServingType_DIRECT_STORAGE,
				Storage: &spec.OnlineStorage_Bigtable{
					Bigtable: &spec.BigTableStorage{
						FeastServingUrl: "localhost",
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.bigtableConfig.ToFeastStorage()
			assert.Equal(t, tC.expectedBigTableConfig, got)
		})
	}
}

func setRequiredEnvironmentVariables() {
	os.Setenv("STANDARD_TRANSFORMER_IMAGE_NAME", "image:1")
	os.Setenv("DEFAULT_FEAST_SERVING_URL", "localhost")
	os.Setenv("FEAST_SERVING_URLS", `[]`)
	os.Setenv("FEAST_CORE_URL", "localhost")
	os.Setenv("FEAST_CORE_AUTH_AUDIENCE", "true")
	os.Setenv("FEAST_BIG_TABLE_CONFIG", `{"serving_url":"10.1.1.3"}`)
	os.Setenv("DEFAULT_FEAST_SOURCE", "BIGTABLE")
}
