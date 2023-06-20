package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestFeastRedisConfig(t *testing.T) {
	baseFilePath := "./testdata/config-1.yaml"
	tenSecDuration := time.Second * 10
	testCases := []struct {
		desc                string
		redisCfgFilePath    string
		env                 map[string]string
		expectedRedisConfig *FeastRedisConfig
		err                 error
	}{
		{
			desc:             "Success: valid single redis config",
			redisCfgFilePath: "./testdata/valid-single-redis-config.yaml",
			env:              map[string]string{},
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
			desc:             "Success: valid redis cluster config",
			redisCfgFilePath: "./testdata/valid-single-redis-cluster-config.yaml",
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
			desc:             "Invalid: pool_size 0",
			redisCfgFilePath: "./testdata/invalid-redis-config-no-pool-size.yaml",
			err:              fmt.Errorf("Key: 'FeastRedisConfig.PoolSize' Error:Field validation for 'PoolSize' failed on the 'gt' tag"),
		},
		{
			desc:             "Invalid: serving_url is not set",
			redisCfgFilePath: "./testdata/invalid-redis-config-no-serving-url.yaml",
			err:              fmt.Errorf("Key: 'FeastRedisConfig.ServingURL' Error:Field validation for 'ServingURL' failed on the 'required' tag"),
		},
		{
			desc:             "Invalid: redis_addresses is not set",
			redisCfgFilePath: "./testdata/invalid-redis-config-no-redis-addresses.yaml",
			err:              fmt.Errorf("Key: 'FeastRedisConfig.RedisAddresses' Error:Field validation for 'RedisAddresses' failed on the 'required' tag"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			filePaths := []string{baseFilePath}
			filePaths = append(filePaths, tC.redisCfgFilePath)

			cfg, err := Load(filePaths...)
			assert.Nil(t, err)

			v := validator.New()
			err = v.Struct(cfg.StandardTransformerConfig.FeastRedisConfig)
			if tC.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, tC.expectedRedisConfig, cfg.StandardTransformerConfig.FeastRedisConfig)
			} else {
				assert.NotNil(t, err)
				assert.EqualError(t, err, tC.err.Error())
			}
		})
	}
}

func TestFeastBigtableConfig(t *testing.T) {
	baseFilePath := "./testdata/config-1.yaml"
	oneMinuteDuration := time.Minute * 1
	twoMinuteDuration := time.Minute * 2
	testCases := []struct {
		desc                   string
		bigtableCfgFilePath    string
		expectedBigTableConfig *FeastBigtableConfig
		err                    error
	}{
		{
			desc:                "Success: valid big table config",
			bigtableCfgFilePath: "./testdata/bigtable-config.yaml",
			expectedBigTableConfig: &FeastBigtableConfig{
				IsUsingDirectStorage: true,
				ServingURL:           "10.1.1.3",
				Project:              "gcp-project",
				Instance:             "instance",
				AppProfile:           "default",
				PoolSize:             3,
				KeepAliveInterval:    &twoMinuteDuration,
				KeepAliveTimeout:     &oneMinuteDuration,
			},
		},
		{
			desc:                "Fail: serving_url is not set",
			bigtableCfgFilePath: "./testdata/invalid-bigtable-config-no-serving-url.yaml",
			err:                 fmt.Errorf("Key: 'FeastBigtableConfig.ServingURL' Error:Field validation for 'ServingURL' failed on the 'required' tag"),
		},
		{
			desc:                "Fail: project is not set",
			bigtableCfgFilePath: "./testdata/invalid-bigtable-config-no-project.yaml",
			err:                 fmt.Errorf("Key: 'FeastBigtableConfig.Project' Error:Field validation for 'Project' failed on the 'required' tag"),
		},
		{
			desc:                "Fail: instance is not set",
			bigtableCfgFilePath: "./testdata/invalid-bigtable-config-no-instance.yaml",
			err:                 fmt.Errorf("Key: 'FeastBigtableConfig.Instance' Error:Field validation for 'Instance' failed on the 'required' tag"),
		},
		{
			desc:                "Fail: pool size is 0",
			bigtableCfgFilePath: "./testdata/invalid-bigtable-config-no-pool-size.yaml",
			err:                 fmt.Errorf("Key: 'FeastBigtableConfig.PoolSize' Error:Field validation for 'PoolSize' failed on the 'gt' tag"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			filePaths := []string{baseFilePath}
			filePaths = append(filePaths, tC.bigtableCfgFilePath)

			cfg, err := Load(filePaths...)
			assert.Nil(t, err)

			v := validator.New()
			err = v.Struct(cfg.StandardTransformerConfig.FeastBigtableConfig)
			if tC.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, tC.expectedBigTableConfig, cfg.StandardTransformerConfig.FeastBigtableConfig)
			} else {
				assert.NotNil(t, err)
				assert.EqualError(t, err, tC.err.Error())
			}
		})
	}
}

func TestRedisConfig_ToFeastStorage(t *testing.T) {
	timeDuration := time.Second * 20
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
	oneMinuteDuration := time.Minute * 1
	twoMinuteDuration := time.Minute * 2
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
	os.Setenv("STANDARDTRANSFORMERCONFIG_IMAGENAME", "image:1")
	os.Setenv("DEFAULT_FEAST_SERVING_URL", "localhost")
	os.Setenv("STANDARDTRANSFORMERCONFIG_FEASTSERVINGURLS", `[]`)
	os.Setenv("STANDARDTRANSFORMERCONFIG_FEASTCOREURL", "localhost")
	os.Setenv("STANDARDTRANSFORMERCONFIG_FEASTCOREAUTHAUDIENCE", "true")
	os.Setenv("STANDARDTRANSFORMERCONFIG_DEFAULTFEASTSOURCE", "BIGTABLE")
	os.Setenv("STANDARDTRANSFORMERCONFIG_SIMULATIONFEAST_FEASTBIGTABLEURL", "online-serving-bt.dev")
	os.Setenv("STANDARDTRANSFORMERCONFIG_SIMULATIONFEAST_FEASTREDISURL", "online-serving-redis.dev")
	os.Setenv("STANDARDTRANSFORMERCONFIG_KAFKA_BROKERS", "kafka-brokers:9999")
}
