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
	baseFilePath := "./testdata/config-1.yaml"
	redisCfgFilePath := "./testdata/redis-config.yaml"
	bigtableCfgFilePath := "./testdata/bigtable-config.yaml"
	simulationRedisURL := "online-redis-serving.dev"
	simulationBigtableURL := "online-bt-serving.dev"
	testCases := []struct {
		desc                   string
		redisConfigFilePath    *string
		bigtableConfigFilePath *string
		simulationRedisURL     *string
		simulationBigtableURL  *string
		feastStorageCfg        feast.FeastStorageConfig

		bigtableCredential string
	}{
		{
			desc:                   "redis config and big table config set",
			redisConfigFilePath:    &redisCfgFilePath,
			bigtableConfigFilePath: &bigtableCfgFilePath,
			simulationRedisURL:     &simulationRedisURL,
			simulationBigtableURL:  &simulationBigtableURL,
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
			redisConfigFilePath:   &redisCfgFilePath,
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
			desc:                   "redis config not set and big table config set",
			bigtableConfigFilePath: &bigtableCfgFilePath,
			simulationRedisURL:     &simulationRedisURL,
			simulationBigtableURL:  &simulationBigtableURL,
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
			filePaths := []string{baseFilePath}
			if tC.redisConfigFilePath != nil {
				filePaths = append(filePaths, *tC.redisConfigFilePath)
			}
			if tC.bigtableConfigFilePath != nil {
				filePaths = append(filePaths, *tC.bigtableConfigFilePath)
			}
			if tC.simulationBigtableURL != nil {
				os.Setenv("STANDARDTRANSFORMERCONFIG_SIMULATIONFEAST_FEASTBIGTABLEURL", *tC.simulationBigtableURL)
			}
			if tC.simulationRedisURL != nil {
				os.Setenv("STANDARDTRANSFORMERCONFIG_SIMULATIONFEAST_FEASTREDISURL", *tC.simulationRedisURL)
			}
			os.Setenv("STANDARDTRANSFORMERCONFIG_BIGTABLECREDENTIAL", tC.bigtableCredential)
			var emptyCfg Config
			cfg, err := Load(&emptyCfg, filePaths...)
			require.NoError(t, err)
			got := cfg.StandardTransformerConfig.ToFeastStorageConfigsForSimulation()
			assert.Equal(t, tC.feastStorageCfg, got)
		})
	}
}

//nolint:errcheck
func TestStandardTransformerConfig_ToFeastStorageConfigs(t *testing.T) {
	baseFilePath := "./testdata/config-1.yaml"
	redisCfgFilePath := "./testdata/redis-config.yaml"
	bigtableCfgFilePath := "./testdata/bigtable-config.yaml"
	simulationRedisURL := "online-redis-serving.dev"
	simulationBigtableURL := "online-bt-serving.dev"
	testCases := []struct {
		desc                   string
		redisConfigFilePath    *string
		bigtableConfigFilePath *string
		feastStorageCfg        feast.FeastStorageConfig
		bigtableCredential     string
	}{
		{
			desc:                   "redis config and big table config set",
			redisConfigFilePath:    &redisCfgFilePath,
			bigtableConfigFilePath: &bigtableCfgFilePath,
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
			desc:                "redis config set and big table config not set",
			redisConfigFilePath: &redisCfgFilePath,
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
			desc:                   "redis config not set and big table config set",
			bigtableConfigFilePath: &bigtableCfgFilePath,
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
			filePaths := []string{baseFilePath}
			if tC.redisConfigFilePath != nil {
				filePaths = append(filePaths, *tC.redisConfigFilePath)
			}
			if tC.bigtableConfigFilePath != nil {
				filePaths = append(filePaths, *tC.bigtableConfigFilePath)
			}
			os.Setenv("STANDARDTRANSFORMERCONFIG_BIGTABLECREDENTIAL", tC.bigtableCredential)
			os.Setenv("STANDARDTRANSFORMERCONFIG_SIMULATIONFEAST_FEASTBIGTABLEURL", simulationBigtableURL)
			os.Setenv("STANDARDTRANSFORMERCONFIG_SIMULATIONFEAST_FEASTREDISURL", simulationRedisURL)
			var emptyCfg Config
			cfg, err := Load(&emptyCfg, filePaths...)
			require.NoError(t, err)
			got := cfg.StandardTransformerConfig.ToFeastStorageConfigs()
			assert.Equal(t, tC.feastStorageCfg, got)
		})
	}
}
