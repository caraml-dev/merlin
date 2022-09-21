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
			desc:              "Success: valid single redis config",
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
			desc:              "Success: valid redis cluster config",
			redisConfigString: `{"is_redis_cluster": true,"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s","min_idle_conn":2}`,
			expectedRedisConfig: &FeastRedisConfig{
				IsRedisCluster: true,
				ServingURL:     "online-storage.merlin.dev",
				RedisAddresses: []string{
					"10.1.1.10", "10.1.1.11",
				},
				PoolSize:    4,
				MaxRetries:  1,
				DialTimeout: &tenSecDuration,
				MinIdleConn: 2,
			},
		},
		{
			desc:              "Invalid: pool_size 0",
			redisConfigString: `{"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 0,"max_retries": 1,"dial_timeout": "10s"}`,
			err:               fmt.Errorf(`envconfig.Process: assigning FEAST_REDIS_CONFIG to FeastRedisConfig: converting '{"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 0,"max_retries": 1,"dial_timeout": "10s"}' to type config.FeastRedisConfig. details: Key: 'FeastRedisConfig.PoolSize' Error:Field validation for 'PoolSize' failed on the 'gt' tag`),
		},
		{
			desc:              "Invalid: serving_url is not set",
			redisConfigString: `{"serving_url":"","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`,
			err:               fmt.Errorf(`envconfig.Process: assigning FEAST_REDIS_CONFIG to FeastRedisConfig: converting '{"serving_url":"","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}' to type config.FeastRedisConfig. details: ServingURL is required`),
		},
		{
			desc:              "Invalid: redis_addresses is not set",
			redisConfigString: `{"serving_url":"online-storage.merlin.dev","redis_addresses":[],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`,
			err:               fmt.Errorf(`envconfig.Process: assigning FEAST_REDIS_CONFIG to FeastRedisConfig: converting '{"serving_url":"online-storage.merlin.dev","redis_addresses":[],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}' to type config.FeastRedisConfig. details: Key: 'FeastRedisConfig.RedisAddresses' Error:Field validation for 'RedisAddresses' failed on the 'gt' tag`),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			os.Setenv("FEAST_REDIS_CONFIG", tC.redisConfigString) //nolint:errcheck

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
	oneMinuteDuration := Duration(time.Minute * 1)
	twoMinuteDuration := Duration(time.Minute * 2)
	testCases := []struct {
		desc                   string
		bigtableString         string
		expectedBigTableConfig *FeastBigtableConfig
		err                    error
	}{
		{
			desc:           "Success: valid env variable",
			bigtableString: `{"serving_url":"10.1.1.3","project":"sample","instance":"bt-instance","app_profile":"slave","pool_size":4,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}`,
			expectedBigTableConfig: &FeastBigtableConfig{
				ServingURL:        "10.1.1.3",
				Project:           "sample",
				Instance:          "bt-instance",
				AppProfile:        "slave",
				PoolSize:          4,
				KeepAliveInterval: &twoMinuteDuration,
				KeepAliveTimeout:  &oneMinuteDuration,
			},
		},
		{
			desc:           "Fail: serving_url is not set",
			bigtableString: `{"serving_url":""}`,
			err:            fmt.Errorf(`envconfig.Process: assigning FEAST_BIG_TABLE_CONFIG to FeastBigtableConfig: converting '{"serving_url":""}' to type config.FeastBigtableConfig. details: ServingURL is required`),
		},
		{
			desc:           "Fail: project is not set",
			bigtableString: `{"serving_url":"10.1.1.3","project":"","instance":"bt-instance","app_profile":"slave","pool_size":4,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}`,
			err:            fmt.Errorf(`envconfig.Process: assigning FEAST_BIG_TABLE_CONFIG to FeastBigtableConfig: converting '{"serving_url":"10.1.1.3","project":"","instance":"bt-instance","app_profile":"slave","pool_size":4,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}' to type config.FeastBigtableConfig. details: Project is required`),
		},
		{
			desc:           "Fail: instance is not set",
			bigtableString: `{"serving_url":"10.1.1.3","project":"sample","instance":"","app_profile":"slave","pool_size":4,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}`,
			err:            fmt.Errorf(`envconfig.Process: assigning FEAST_BIG_TABLE_CONFIG to FeastBigtableConfig: converting '{"serving_url":"10.1.1.3","project":"sample","instance":"","app_profile":"slave","pool_size":4,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}' to type config.FeastBigtableConfig. details: Instance is required`),
		},
		{
			desc:           "Fail: pool size is 0",
			bigtableString: `{"serving_url":"10.1.1.3","project":"sample","instance":"bt-instance","app_profile":"slave","pool_size":0,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}`,
			err:            fmt.Errorf(`envconfig.Process: assigning FEAST_BIG_TABLE_CONFIG to FeastBigtableConfig: converting '{"serving_url":"10.1.1.3","project":"sample","instance":"bt-instance","app_profile":"slave","pool_size":0,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}' to type config.FeastBigtableConfig. details: Key: 'FeastBigtableConfig.PoolSize' Error:Field validation for 'PoolSize' failed on the 'gt' tag`),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			os.Setenv("FEAST_BIG_TABLE_CONFIG", tC.bigtableString) //nolint:errcheck
			var cfg StandardTransformerConfig
			err := envconfig.Process("", &cfg)
			if err == nil {
				assert.Equal(t, tC.expectedBigTableConfig, cfg.FeastBigtableConfig)
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
			desc: "Complete redis cluster",
			redisConfig: &FeastRedisConfig{
				IsRedisCluster:       true,
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
			desc: "Complete single redis",
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
				Storage: &spec.OnlineStorage_Redis{
					Redis: &spec.RedisStorage{
						FeastServingUrl: "localhost",
						RedisAddress:    "10.1.1.2",
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
				IsRedisCluster: true,
				ServingURL:     "localhost",
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
	oneMinuteDuration := Duration(time.Minute * 1)
	twoMinuteDuration := Duration(time.Minute * 2)
	testCases := []struct {
		desc                   string
		bigtableConfig         *FeastBigtableConfig
		bigtableCredential     string
		expectedBigTableConfig *spec.OnlineStorage
	}{
		{
			desc: "Success",
			bigtableConfig: &FeastBigtableConfig{
				IsUsingDirectStorage: true,
				ServingURL:           "localhost",
				Project:              "gcp-project",
				Instance:             "instance",
				AppProfile:           "default",
				PoolSize:             3,
				KeepAliveInterval:    &twoMinuteDuration,
				KeepAliveTimeout:     &oneMinuteDuration,
			},
			bigtableCredential: `eyJrZXkiOiJ2YWx1ZSJ9`,
			expectedBigTableConfig: &spec.OnlineStorage{
				ServingType: spec.ServingType_DIRECT_STORAGE,
				Storage: &spec.OnlineStorage_Bigtable{
					Bigtable: &spec.BigTableStorage{
						FeastServingUrl: "localhost",
						Project:         "gcp-project",
						Instance:        "instance",
						AppProfile:      "default",
						Option: &spec.BigTableOption{
							GrpcConnectionPool: 3,
							KeepAliveInterval:  durationpb.New(time.Duration(twoMinuteDuration)),
							KeepAliveTimeout:   durationpb.New(time.Duration(oneMinuteDuration)),
							CredentialJson:     "eyJrZXkiOiJ2YWx1ZSJ9",
						},
					},
				},
			},
		},
		{
			desc: "Success: not using direct storage, and keep alive configurations are not set",
			bigtableConfig: &FeastBigtableConfig{
				IsUsingDirectStorage: false,
				ServingURL:           "localhost",
				Project:              "gcp-project",
				Instance:             "instance",
				AppProfile:           "default",
				PoolSize:             3,
			},
			expectedBigTableConfig: &spec.OnlineStorage{
				ServingType: spec.ServingType_FEAST_GRPC,
				Storage: &spec.OnlineStorage_Bigtable{
					Bigtable: &spec.BigTableStorage{
						FeastServingUrl: "localhost",
						Project:         "gcp-project",
						Instance:        "instance",
						AppProfile:      "default",
						Option: &spec.BigTableOption{
							GrpcConnectionPool: 3,
							KeepAliveInterval:  nil,
							KeepAliveTimeout:   nil,
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.bigtableConfig.ToFeastStorageWithCredential(tC.bigtableCredential)
			assert.Equal(t, tC.expectedBigTableConfig, got)
		})
	}
}

//nolint:errcheck
func setRequiredEnvironmentVariables() {
	os.Setenv("STANDARD_TRANSFORMER_IMAGE_NAME", "image:1")
	os.Setenv("DEFAULT_FEAST_SERVING_URL", "localhost")
	os.Setenv("FEAST_SERVING_URLS", `[]`)
	os.Setenv("FEAST_CORE_URL", "localhost")
	os.Setenv("FEAST_CORE_AUTH_AUDIENCE", "true")
	os.Setenv("DEFAULT_FEAST_SOURCE", "BIGTABLE")
	os.Setenv("SIMULATION_FEAST_BIGTABLE_URL", "online-serving-bt.dev")
	os.Setenv("SIMULATION_FEAST_REDIS_URL", "online-serving-redis.dev")
}
