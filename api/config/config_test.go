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
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/caraml-dev/mlp/api/pkg/webhooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	v1 "k8s.io/api/core/v1"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
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

func setupNewEnv(envMaps ...map[string]string) error {
	os.Clearenv()

	var err error
	for _, envMap := range envMaps {
		for key, val := range envMap {
			err = os.Setenv(key, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TestLoad(t *testing.T) {
	zeroSecond, _ := time.ParseDuration("0s")
	twoMinutes := 2 * time.Minute
	oneMinute := 1 * time.Minute

	tests := map[string]struct {
		filepaths []string
		env       map[string]string
		want      *Config
		wantErr   bool
	}{
		"multiple file": {
			filepaths: []string{"testdata/base-configs-1.yaml", "testdata/bigtable-config.yaml"},
			want: &Config{
				Environment:               "dev",
				Port:                      8080,
				LoggerDestinationURL:      "kafka:logger.destination:6668",
				MLObsLoggerDestinationURL: "mlobs:kafka.destination:6668",
				Sentry: sentry.Config{
					DSN:     "",
					Enabled: false,
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				NewRelic: newrelic.Config{
					Enabled:           true,
					AppName:           "merlin-test",
					Labels:            map[string]interface{}{},
					IgnoreStatusCodes: []int{400, 401, 402},
				},
				NumOfQueueWorkers:     2,
				SwaggerPath:           "swaggerpath.com",
				NamespaceLabelPrefix:  "caraml.dev/",
				DeploymentLabelPrefix: "caraml.com/",
				PyfuncGRPCOptions:     "{}",
				DbConfig: DatabaseConfig{
					Host:            "localhost",
					Port:            5432,
					User:            "merlin",
					Password:        "merlin",
					Database:        "merlin",
					MigrationPath:   "file://db-migrations",
					ConnMaxIdleTime: zeroSecond,
					ConnMaxLifetime: zeroSecond,
					MaxIdleConns:    0,
					MaxOpenConns:    0,
				},
				ClusterConfig: ClusterConfig{
					InClusterConfig:       false,
					EnvironmentConfigPath: "/app/cluster-env/environments.yaml",
					EnvironmentConfigs:    []*EnvironmentConfig{},
				},
				ImageBuilderConfig: ImageBuilderConfig{
					ClusterName: "test-cluster",
					GcpProject:  "test-project",
					BaseImage: BaseImageConfig{
						ImageName:           "ghcr.io/caraml-dev/merlin/merlin-pyfunc-base:0.0.0",
						DockerfilePath:      "pyfunc-server/docker/Dockerfile",
						BuildContextURI:     "git://github.com/caraml-dev/merlin.git#refs/tags/v0.0.0",
						BuildContextSubPath: "python",
					},
					PredictionJobBaseImage: BaseImageConfig{
						ImageName:           "ghcr.io/caraml-dev/merlin/merlin-pyspark-base:0.0.0",
						DockerfilePath:      "batch-predictor/docker/app.Dockerfile",
						BuildContextURI:     "git://github.com/caraml-dev/merlin.git#refs/tags/v0.0.0",
						BuildContextSubPath: "python",
						MainAppPath:         "/home/spark/merlin-spark-app/main.py",
					},
					BuildNamespace:         "caraml",
					DockerRegistry:         "test-docker.pkg.dev/test/caraml-registry",
					BuildTimeout:           "30m",
					KanikoImage:            "gcr.io/kaniko-project/executor:v1.21.0",
					KanikoServiceAccount:   "kaniko-merlin",
					KanikoPushRegistryType: "docker",
					KanikoAdditionalArgs:   []string{"--test=true", "--no-logs=false"},
					DefaultResources: ResourceRequestsLimits{
						Requests: Resource{
							CPU:    "1",
							Memory: "4Gi",
						},
						Limits: Resource{
							CPU:    "1",
							Memory: "4Gi",
						},
					},
					Retention: 48 * time.Hour,
					Tolerations: []v1.Toleration{
						{
							Key:      "purpose.caraml.com/batch",
							Value:    "true",
							Operator: v1.TolerationOpEqual,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
					NodeSelectors: map[string]string{
						"purpose.caraml.com/batch": "true",
					},
					MaximumRetry: 3,
					K8sConfig: &mlpcluster.K8sConfig{
						Cluster: &clientcmdapiv1.Cluster{
							Server:                   "https://127.0.0.1",
							CertificateAuthorityData: []byte("some_string"),
						},
						AuthInfo: &clientcmdapiv1.AuthInfo{
							Exec: &clientcmdapiv1.ExecConfig{
								APIVersion:         "some_api_version",
								Command:            "some_command",
								InteractiveMode:    clientcmdapiv1.IfAvailableExecInteractiveMode,
								ProvideClusterInfo: true,
							},
						},
						Name: "dev-server",
					},
					SafeToEvict:             false,
					SupportedPythonVersions: []string{"3.8.*", "3.9.*", "3.10.*"},
				},
				BatchConfig: BatchConfig{
					Tolerations: []v1.Toleration{
						{
							Key:      "purpose.caraml.com/batch",
							Value:    "true",
							Operator: v1.TolerationOpEqual,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
					NodeSelectors: map[string]string{
						"purpose.caraml.com/batch": "true",
					},
				},
				AuthorizationConfig: AuthorizationConfig{
					AuthorizationEnabled: true,
					KetoRemoteRead:       "http://mlp-keto-read:80",
					KetoRemoteWrite:      "http://mlp-keto-write:80",
					Caching: &InMemoryCacheConfig{
						Enabled:                     true,
						KeyExpirySeconds:            600,
						CacheCleanUpIntervalSeconds: 750,
					},
				},
				MlpAPIConfig: MlpAPIConfig{
					APIHost: "http://mlp.caraml.svc.local:8080",
				},
				FeatureToggleConfig: FeatureToggleConfig{
					MonitoringConfig: MonitoringConfig{
						MonitoringEnabled:    true,
						MonitoringBaseURL:    "https://test.io/merlin-overview-dashboard",
						MonitoringJobBaseURL: "https://test.io/batch-predictions-dashboard",
					},
					AlertConfig: AlertConfig{
						AlertEnabled: true,
						GitlabConfig: GitlabConfig{
							BaseURL:             "https://test.io/",
							DashboardRepository: "dashboards/merlin",
							DashboardBranch:     "master",
							AlertRepository:     "alerts/merlin",
							AlertBranch:         "master",
						},
						WardenConfig: WardenConfig{
							APIHost: "https://test.io/",
						},
					},
					ModelDeletionConfig: ModelDeletionConfig{
						Enabled: false,
					},
				},
				ReactAppConfig: ReactAppConfig{
					DocURL: []Documentation{
						{
							Label: "Merlin User Guide",
							Href:  "https://guide.io",
						},
						{
							Label: "Merlin Examples",
							Href:  "https://examples.io",
						},
					},
					DockerRegistries: "docker.io",
					Environment:      "dev",
					FeastCoreURL:     "https://feastcore.io",
					HomePage:         "/merlin",
					MerlinURL:        "https://test.io/api/merlin/v1",
					MlpURL:           "/api",
					OauthClientID:    "abc.apps.clientid.com",
					UPIDocumentation: "https://github.com/caraml-dev/universal-prediction-interface/blob/main/docs/api_markdown/caraml/upi/v1/index.md",
				},
				UI: UIConfig{
					StaticPath: "ui/build",
					IndexPath:  "index.html",
				},
				StandardTransformerConfig: StandardTransformerConfig{
					FeastServingURLs: []FeastServingURL{},
					FeastBigtableConfig: &FeastBigtableConfig{
						IsUsingDirectStorage: true,
						ServingURL:           "10.1.1.3",
						Project:              "gcp-project",
						Instance:             "instance",
						AppProfile:           "default",
						PoolSize:             3,
						KeepAliveInterval:    &twoMinutes,
						KeepAliveTimeout:     &oneMinute,
					},
					FeastGPRCConnCount: 10,
					FeastServingKeepAlive: &FeastServingKeepAliveConfig{
						Enabled: false,
						Time:    60 * time.Second,
						Timeout: time.Second,
					},
					ModelClientKeepAlive: &ModelClientKeepAliveConfig{
						Enabled: false,
						Time:    60 * time.Second,
						Timeout: 5 * time.Second,
					},
					ModelServerConnCount: 10,
					DefaultFeastSource:   spec.ServingSource_BIGTABLE,
					Jaeger: JaegerConfig{
						SamplerParam: "0.01",
						Disabled:     "true",
					},
					Kafka: KafkaConfig{
						CompressionType:     "none",
						MaxMessageSizeBytes: 1048588,
						ConnectTimeoutMS:    1000,
						SerializationFmt:    "protobuf",
						LingerMS:            100,
						NumPartitions:       24,
						ReplicationFactor:   3,
						AdditionalConfig:    "{}",
					},
					SimulatorFeastClientMaxConcurrentRequests: 100,
				},
				MlflowConfig: MlflowConfig{
					TrackingURL:         "https://mlflow.io",
					ArtifactServiceType: "gcs",
				},
				PyFuncPublisherConfig: PyFuncPublisherConfig{
					Kafka: KafkaConfig{
						Brokers:             "kafka:broker.destination:6668",
						Acks:                0,
						CompressionType:     "none",
						MaxMessageSizeBytes: 1048588,
						ConnectTimeoutMS:    1000,
						SerializationFmt:    "protobuf",
						LingerMS:            500,
						NumPartitions:       24,
						ReplicationFactor:   3,
						AdditionalConfig:    "{}",
					},
					SamplingRatioRate: 0.01,
				},
				InferenceServiceDefaults: InferenceServiceDefaults{
					UserContainerCPUDefaultLimit:          "100",
					UserContainerCPULimitRequestFactor:    0,
					UserContainerMemoryLimitRequestFactor: 2,
					DefaultEnvVarsWithoutCPULimits: []v1.EnvVar{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
				ObservabilityPublisher: ObservabilityPublisher{
					KafkaConsumer: KafkaConsumer{
						AdditionalConsumerConfig: map[string]string{},
					},
					DeploymentTimeout: 30 * time.Minute,
				},
				WebhooksConfig: webhooks.Config{
					Enabled: true,
					Config: map[webhooks.EventType][]webhooks.WebhookConfig{
						"on-model-deployed": {
							{
								Name:          "sync-webhooks",
								URL:           "http://127.0.0.1:8000/sync-webhook",
								Method:        "POST",
								FinalResponse: true,
							},
							{
								Name:   "async-webhooks",
								URL:    "http://127.0.0.1:8000/async-webhook",
								Method: "POST",
								Async:  true,
							},
						},
					},
				},
			},
		},
		"missing file": {
			filepaths: []string{"testdata/this-file-should-not-exist.yaml"},
			wantErr:   true,
		},
		"invalid duration format": {
			filepaths: []string{"testdata/invalid-duration-format.yaml"},
			wantErr:   true,
		},
		"invalid file format": {
			filepaths: []string{"testdata/invalid-file-format.yaml"},
			wantErr:   true,
		},
		"invalid type": {
			filepaths: []string{"testdata/invalid-type.yaml"},
			wantErr:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, setupNewEnv(tt.env))
			var emptyCfg Config
			got, err := Load(&emptyCfg, tt.filepaths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
