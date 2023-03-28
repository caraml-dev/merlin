// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestFeastServingURLs_URLs(t *testing.T) {
	tests := []struct {
		name string
		u    *FeastServingURLs
		want []string
	}{
		{
			name: "",
			u: &FeastServingURLs{
				{Host: "localhost:6566"},
				{Host: "localhost:6567"},
			},
			want: []string{"localhost:6566", "localhost:6567"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.URLs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FeastServingURLs.URLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:errcheck
func TestStandardTransformerConfig_ToFeastStorageConfigsForSimulation(t *testing.T) {
	redisCfg := `{"is_redis_cluster": true,"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`
	bigtableCfg := `{"serving_url":"10.1.1.3","project":"gcp-project","is_using_direct_storage":true,"instance":"instance","app_profile":"default","pool_size":3,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}`
	simulationRedisURL := "online-redis-serving.dev"
	simulationBigtableURL := "online-bt-serving.dev"
	testCases := []struct {
		desc                  string
		redisConfig           *string
		bigtableConfig        *string
		simulationRedisURL    *string
		simulationBigtableURL *string
		feastStorageCfg       feast.FeastStorageConfig

		bigtableCredential string
	}{
		{
			desc:                  "redis config and big table config set",
			redisConfig:           &redisCfg,
			bigtableConfig:        &bigtableCfg,
			simulationRedisURL:    &simulationRedisURL,
			simulationBigtableURL: &simulationBigtableURL,
			feastStorageCfg: feast.FeastStorageConfig{
				spec.ServingSource_REDIS: &spec.OnlineStorage{
					ServingType: spec.ServingType_FEAST_GRPC,
					Storage: &spec.OnlineStorage_RedisCluster{
						RedisCluster: &spec.RedisClusterStorage{
							FeastServingUrl: "online-redis-serving.dev",
							RedisAddress:    []string{"10.1.1.10", "10.1.1.11"},
							Option: &spec.RedisOption{
								PoolSize:    4,
								MaxRetries:  1,
								DialTimeout: durationpb.New(time.Second * 10),
							},
						},
					},
				},
				spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
					ServingType: spec.ServingType_FEAST_GRPC,
					Storage: &spec.OnlineStorage_Bigtable{
						Bigtable: &spec.BigTableStorage{
							FeastServingUrl: "online-bt-serving.dev",
							Project:         "gcp-project",
							Instance:        "instance",
							AppProfile:      "default",
							Option: &spec.BigTableOption{
								GrpcConnectionPool: 3,
								KeepAliveInterval:  durationpb.New(time.Minute * 2),
								KeepAliveTimeout:   durationpb.New(time.Minute * 1),
								CredentialJson:     "eyJrZXkiOiJ2YWx1ZSJ9",
							},
						},
					},
				},
			},
			bigtableCredential: `eyJrZXkiOiJ2YWx1ZSJ9`,
		},
		{
			desc:                  "redis config set and big table config not set",
			redisConfig:           &redisCfg,
			simulationRedisURL:    &simulationRedisURL,
			simulationBigtableURL: &simulationBigtableURL,
			feastStorageCfg: feast.FeastStorageConfig{
				spec.ServingSource_REDIS: &spec.OnlineStorage{
					ServingType: spec.ServingType_FEAST_GRPC,
					Storage: &spec.OnlineStorage_RedisCluster{
						RedisCluster: &spec.RedisClusterStorage{
							FeastServingUrl: "online-redis-serving.dev",
							RedisAddress:    []string{"10.1.1.10", "10.1.1.11"},
							Option: &spec.RedisOption{
								PoolSize:    4,
								MaxRetries:  1,
								DialTimeout: durationpb.New(time.Second * 10),
							},
						},
					},
				},
			},
		},
		{
			desc:                  "redis config not set and big table config set",
			bigtableConfig:        &bigtableCfg,
			simulationRedisURL:    &simulationRedisURL,
			simulationBigtableURL: &simulationBigtableURL,
			feastStorageCfg: feast.FeastStorageConfig{
				spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
					ServingType: spec.ServingType_FEAST_GRPC,
					Storage: &spec.OnlineStorage_Bigtable{
						Bigtable: &spec.BigTableStorage{
							FeastServingUrl: "online-bt-serving.dev",
							Project:         "gcp-project",
							Instance:        "instance",
							AppProfile:      "default",
							Option: &spec.BigTableOption{
								GrpcConnectionPool: 3,
								KeepAliveInterval:  durationpb.New(time.Minute * 2),
								KeepAliveTimeout:   durationpb.New(time.Minute * 1),
							},
						},
					},
				},
			},
		},
		{
			desc:                  "redis config and big table config not set",
			simulationRedisURL:    &simulationRedisURL,
			simulationBigtableURL: &simulationBigtableURL,
			feastStorageCfg:       feast.FeastStorageConfig{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			if tC.redisConfig != nil {
				os.Setenv("FEAST_REDIS_CONFIG", *tC.redisConfig)
			}
			if tC.bigtableConfig != nil {
				os.Setenv("FEAST_BIG_TABLE_CONFIG", *tC.bigtableConfig)
			}
			if tC.simulationBigtableURL != nil {
				os.Setenv("SIMULATION_FEAST_BIGTABLE_URL", *tC.simulationBigtableURL)
			}
			if tC.simulationRedisURL != nil {
				os.Setenv("SIMULATION_FEAST_REDIS_URL", *tC.simulationRedisURL)
			}
			os.Setenv("FEAST_BIGTABLE_CREDENTIAL", tC.bigtableCredential)

			var cfg StandardTransformerConfig
			err := envconfig.Process("", &cfg)
			require.NoError(t, err)
			got := cfg.ToFeastStorageConfigsForSimulation()
			assert.Equal(t, tC.feastStorageCfg, got)
		})
	}
}

//nolint:errcheck
func TestStandardTransformerConfig_ToFeastStorageConfigs(t *testing.T) {
	redisCfg := `{"is_redis_cluster": true,"serving_url":"online-storage.merlin.dev","redis_addresses":["10.1.1.10", "10.1.1.11"],"pool_size": 4,"max_retries": 1,"dial_timeout": "10s"}`
	bigtableCfg := `{"serving_url":"10.1.1.3","project":"gcp-project","is_using_direct_storage":true,"instance":"instance","app_profile":"default","pool_size":3,"keep_alive_interval":"2m","keep_alive_timeout":"1m"}`
	simulationRedisURL := "online-redis-serving.dev"
	simulationBigtableURL := "online-bt-serving.dev"
	testCases := []struct {
		desc               string
		redisConfig        *string
		bigtableConfig     *string
		feastStorageCfg    feast.FeastStorageConfig
		bigtableCredential string
	}{
		{
			desc:           "redis config and big table config set",
			redisConfig:    &redisCfg,
			bigtableConfig: &bigtableCfg,
			feastStorageCfg: feast.FeastStorageConfig{
				spec.ServingSource_REDIS: &spec.OnlineStorage{
					ServingType: spec.ServingType_FEAST_GRPC,
					Storage: &spec.OnlineStorage_RedisCluster{
						RedisCluster: &spec.RedisClusterStorage{
							FeastServingUrl: "online-storage.merlin.dev",
							RedisAddress:    []string{"10.1.1.10", "10.1.1.11"},
							Option: &spec.RedisOption{
								PoolSize:    4,
								MaxRetries:  1,
								DialTimeout: durationpb.New(time.Second * 10),
							},
						},
					},
				},
				spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
					ServingType: spec.ServingType_DIRECT_STORAGE,
					Storage: &spec.OnlineStorage_Bigtable{
						Bigtable: &spec.BigTableStorage{
							FeastServingUrl: "10.1.1.3",
							Project:         "gcp-project",
							Instance:        "instance",
							AppProfile:      "default",
							Option: &spec.BigTableOption{
								GrpcConnectionPool: 3,
								KeepAliveInterval:  durationpb.New(time.Minute * 2),
								KeepAliveTimeout:   durationpb.New(time.Minute * 1),
								CredentialJson:     "eyJrZXkiOiJ2YWx1ZSJ9",
							},
						},
					},
				},
			},
			bigtableCredential: `eyJrZXkiOiJ2YWx1ZSJ9`,
		},
		{
			desc:        "redis config set and big table config not set",
			redisConfig: &redisCfg,
			feastStorageCfg: feast.FeastStorageConfig{
				spec.ServingSource_REDIS: &spec.OnlineStorage{
					ServingType: spec.ServingType_FEAST_GRPC,
					Storage: &spec.OnlineStorage_RedisCluster{
						RedisCluster: &spec.RedisClusterStorage{
							FeastServingUrl: "online-storage.merlin.dev",
							RedisAddress:    []string{"10.1.1.10", "10.1.1.11"},
							Option: &spec.RedisOption{
								PoolSize:    4,
								MaxRetries:  1,
								DialTimeout: durationpb.New(time.Second * 10),
							},
						},
					},
				},
			},
		},
		{
			desc:           "redis config not set and big table config set",
			bigtableConfig: &bigtableCfg,
			feastStorageCfg: feast.FeastStorageConfig{
				spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
					ServingType: spec.ServingType_DIRECT_STORAGE,
					Storage: &spec.OnlineStorage_Bigtable{
						Bigtable: &spec.BigTableStorage{
							FeastServingUrl: "10.1.1.3",
							Project:         "gcp-project",
							Instance:        "instance",
							AppProfile:      "default",
							Option: &spec.BigTableOption{
								GrpcConnectionPool: 3,
								KeepAliveInterval:  durationpb.New(time.Minute * 2),
								KeepAliveTimeout:   durationpb.New(time.Minute * 1),
							},
						},
					},
				},
			},
		},
		{
			desc:            "redis config and big table config not set",
			feastStorageCfg: feast.FeastStorageConfig{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Clearenv()
			setRequiredEnvironmentVariables()
			if tC.redisConfig != nil {
				os.Setenv("FEAST_REDIS_CONFIG", *tC.redisConfig)
			}
			if tC.bigtableConfig != nil {
				os.Setenv("FEAST_BIG_TABLE_CONFIG", *tC.bigtableConfig)
			}
			os.Setenv("FEAST_BIGTABLE_CREDENTIAL", tC.bigtableCredential)
			os.Setenv("SIMULATION_FEAST_BIGTABLE_URL", simulationBigtableURL)
			os.Setenv("SIMULATION_FEAST_REDIS_URL", simulationRedisURL)
			var cfg StandardTransformerConfig
			err := envconfig.Process("", &cfg)
			require.NoError(t, err)
			got := cfg.ToFeastStorageConfigs()
			assert.Equal(t, tC.feastStorageCfg, got)
		})
	}
}
