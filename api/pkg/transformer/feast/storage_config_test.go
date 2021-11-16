package feast

import (
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestOverwriteFeastOptionsConfig(t *testing.T) {
	boolTrue := true
	boolFalse := false
	poolSize := int32(4)
	twoMinutes := time.Minute * 2
	fourMinutes := time.Minute * 4

	testCases := []struct {
		desc         string
		opts         Options
		redisConfig  RedisOverwriteConfig
		expectedOpts Options
	}{
		{
			desc: "Redis cluster do not overwrite",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolTrue,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Redis cluster overwrite serving type",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_FEAST_GRPC,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolTrue,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Redis cluster overwrite serving type, pool size",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolFalse,
				PoolSize:                  &poolSize,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_FEAST_GRPC,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     4,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Redis cluster overwrite serving type, pool size, read timeout, write timeout",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolFalse,
				PoolSize:                  &poolSize,
				ReadTimeout:               &twoMinutes,
				WriteTimeout:              &fourMinutes,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_FEAST_GRPC,
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize:     4,
									ReadTimeout:  durationpb.New(time.Minute * 2),
									WriteTimeout: durationpb.New(time.Minute * 4),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Single redis do not overwrite",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolTrue,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Single redis overwrite serving type",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_FEAST_GRPC,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolTrue,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Single redis overwrite serving type, pool size",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolFalse,
				PoolSize:                  &poolSize,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_FEAST_GRPC,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     4,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Single redis overwrite serving type, pool size, read timeout, write timeout",
			opts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_DIRECT_STORAGE,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     10,
									ReadTimeout:  durationpb.New(time.Minute * 1),
									WriteTimeout: durationpb.New(time.Minute * 2),
								},
							},
						},
					},
				},
			},
			redisConfig: RedisOverwriteConfig{
				RedisDirectStorageEnabled: &boolFalse,
				PoolSize:                  &poolSize,
				ReadTimeout:               &twoMinutes,
				WriteTimeout:              &fourMinutes,
			},
			expectedOpts: Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						ServingType: spec.ServingType_FEAST_GRPC,
						Storage: &spec.OnlineStorage_Redis{
							Redis: &spec.RedisStorage{
								FeastServingUrl: "localhost:6866",
								RedisAddress:    "10.1.1.2",
								Option: &spec.RedisOption{
									PoolSize:     4,
									ReadTimeout:  durationpb.New(time.Minute * 2),
									WriteTimeout: durationpb.New(time.Minute * 4),
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := OverwriteFeastOptionsConfig(tC.opts, tC.redisConfig)
			assert.Equal(t, tC.expectedOpts, got)
		})
	}
}
