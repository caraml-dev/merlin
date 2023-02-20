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
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
	v1 "k8s.io/api/core/v1"

	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/spec"
	internalValidator "github.com/gojek/merlin/pkg/validator"
	mlpcluster "github.com/gojek/mlp/api/pkg/cluster"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
)

const (
	MaxDeployedVersion = 2
)

type Config struct {
	Environment           string          `envconfig:"ENVIRONMENT" default:"dev"`
	Port                  int             `envconfig:"PORT" default:"8080"`
	LoggerDestinationURL  string          `envconfig:"LOGGER_DESTINATION_URL"`
	Sentry                sentry.Config   `envconfig:"SENTRY" split_words:"false"`
	NewRelic              newrelic.Config `envconfig:"NEWRELIC" split_words:"false" `
	EnvironmentConfigPath string          `envconfig:"DEPLOYMENT_CONFIG_PATH" required:"true"`
	NumOfQueueWorkers     int             `envconfig:"NUM_OF_WORKERS" default:"2"`
	SwaggerPath           string          `envconfig:"SWAGGER_PATH" default:"./swagger.yaml"`

	DbConfig                  DatabaseConfig
	ImageBuilderConfig        ImageBuilderConfig
	EnvironmentConfigs        []EnvironmentConfig
	AuthorizationConfig       AuthorizationConfig
	MlpAPIConfig              MlpAPIConfig
	FeatureToggleConfig       FeatureToggleConfig
	ReactAppConfig            ReactAppConfig
	UI                        UIConfig
	StandardTransformerConfig StandardTransformerConfig
	MlflowConfig              MlflowConfig
}

// UIConfig stores the configuration for the UI.
type UIConfig struct {
	StaticPath string `envconfig:"UI_STATIC_PATH" default:"ui/build"`
	IndexPath  string `envconfig:"UI_INDEX_PATH" default:"index.html"`
}

type ReactAppConfig struct {
	DocURL            Documentations `envconfig:"REACT_APP_MERLIN_DOCS_URL" json:"REACT_APP_MERLIN_DOCS_URL,omitempty"`
	DockerRegistries  string         `envconfig:"REACT_APP_DOCKER_REGISTRIES" json:"REACT_APP_DOCKER_REGISTRIES,omitempty"`
	Environment       string         `envconfig:"REACT_APP_ENVIRONMENT" json:"REACT_APP_ENVIRONMENT,omitempty"`
	FeastCoreURL      string         `envconfig:"REACT_APP_FEAST_CORE_API" json:"REACT_APP_FEAST_CORE_API,omitempty"`
	HomePage          string         `envconfig:"REACT_APP_HOMEPAGE" json:"REACT_APP_HOMEPAGE,omitempty"`
	MaxAllowedReplica int            `envconfig:"REACT_APP_MAX_ALLOWED_REPLICA" default:"20" json:"REACT_APP_MAX_ALLOWED_REPLICA,omitempty"`
	MerlinURL         string         `envconfig:"REACT_APP_MERLIN_API" json:"REACT_APP_MERLIN_API,omitempty"`
	MlpURL            string         `envconfig:"REACT_APP_MLP_API" json:"REACT_APP_MLP_API,omitempty"`
	OauthClientID     string         `envconfig:"REACT_APP_OAUTH_CLIENT_ID" json:"REACT_APP_OAUTH_CLIENT_ID,omitempty"`
	SentryDSN         string         `envconfig:"REACT_APP_SENTRY_DSN" json:"REACT_APP_SENTRY_DSN,omitempty"`
	UPIDocumentation  string         `envconfig:"REACT_APP_UPI_DOC_URL" json:"REACT_APP_UPI_DOC_URL,omitempty"`
	CPUCost           string         `envconfig:"REACT_APP_CPU_COST" json:"REACT_APP_CPU_COST,omitempty"`
	MemoryCost        string         `envconfig:"REACT_APP_MEMORY_COST" json:"REACT_APP_MEMORY_COST,omitempty"`
}

type BaseImageConfigs map[string]BaseImageConfig

// A struct containing configuration details for each base image type
type BaseImageConfig struct {
	// docker image name with path
	ImageName string `json:"imageName"`
	// Dockerfile Path within the build context
	DockerfilePath string `json:"dockerfilePath"`
	// GCS URL Containing build context
	BuildContextURI string `json:"buildContextURI"`
	// path to main file to run application
	MainAppPath string `json:"mainAppPath"`
}

// Decoder to decode the env variable which is a nested map into a list of BaseImageConfig
func (b *BaseImageConfigs) Decode(value string) error {
	var configList BaseImageConfigs

	if err := json.Unmarshal([]byte(value), &configList); err != nil {
		return err
	}
	*b = configList
	return nil
}

type Documentations []Documentation

type Documentation struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

func (docs *Documentations) Decode(value string) error {
	var listOfDoc Documentations

	if err := json.Unmarshal([]byte(value), &listOfDoc); err != nil {
		return err
	}
	*docs = listOfDoc
	return nil
}

type DatabaseConfig struct {
	Host          string `envconfig:"DATABASE_HOST" required:"true"`
	Port          int    `envconfig:"DATABASE_PORT" default:"5432"`
	User          string `envconfig:"DATABASE_USER" required:"true"`
	Password      string `envconfig:"DATABASE_PASSWORD" required:"true"`
	Database      string `envconfig:"DATABASE_NAME" default:"mlp"`
	MigrationPath string `envconfig:"DATABASE_MIGRATIONS_PATH" default:"file://db-migrations"`

	ConnMaxIdleTime time.Duration `envconfig:"DATABASE_CONN_MAX_IDLE_TIME" default:"0s"`
	ConnMaxLifetime time.Duration `envconfig:"DATABASE_CONN_MAX_LIFETIME" default:"0s"`
	MaxIdleConns    int           `envconfig:"DATABASE_MAX_IDLE_CONNS" default:"0"`
	MaxOpenConns    int           `envconfig:"DATABASE_MAX_OPEN_CONNS" default:"0"`
}

type ImageBuilderConfig struct {
	ClusterName                  string           `envconfig:"IMG_BUILDER_CLUSTER_NAME"`
	GcpProject                   string           `envconfig:"IMG_BUILDER_GCP_PROJECT"`
	BuildContextURI              string           `envconfig:"IMG_BUILDER_BUILD_CONTEXT_URI"`
	ContextSubPath               string           `envconfig:"IMG_BUILDER_CONTEXT_SUB_PATH"`
	DockerfilePath               string           `envconfig:"IMG_BUILDER_DOCKERFILE_PATH" default:"./Dockerfile"`
	BaseImages                   BaseImageConfigs `envconfig:"IMG_BUILDER_BASE_IMAGES"`
	PredictionJobBuildContextURI string           `envconfig:"IMG_BUILDER_PREDICTION_JOB_BUILD_CONTEXT_URI"`
	PredictionJobContextSubPath  string           `envconfig:"IMG_BUILDER_PREDICTION_JOB_CONTEXT_SUB_PATH"`
	PredictionJobDockerfilePath  string           `envconfig:"IMG_BUILDER_PREDICTION_JOB_DOCKERFILE_PATH" default:"./Dockerfile"`
	PredictionJobBaseImages      BaseImageConfigs `envconfig:"IMG_BUILDER_PREDICTION_JOB_BASE_IMAGES"`
	BuildNamespace               string           `envconfig:"IMG_BUILDER_NAMESPACE" default:"mlp"`
	DockerRegistry               string           `envconfig:"IMG_BUILDER_DOCKER_REGISTRY"`
	BuildTimeout                 string           `envconfig:"IMG_BUILDER_TIMEOUT" default:"10m"`
	KanikoImage                  string           `envconfig:"IMG_BUILDER_KANIKO_IMAGE" default:"gcr.io/kaniko-project/executor:v1.6.0"`
	KanikoServiceAccount         string           `envconfig:"IMG_BUILDER_KANIKO_SERVICE_ACCOUNT"`
	// How long to keep the image building job resource in the Kubernetes cluster. Default: 2 days (48 hours).
	Retention     time.Duration        `envconfig:"IMG_BUILDER_RETENTION" default:"48h"`
	Tolerations   Tolerations          `envconfig:"IMG_BUILDER_TOLERATIONS"`
	NodeSelectors DictEnv              `envconfig:"IMG_BUILDER_NODE_SELECTORS"`
	MaximumRetry  int32                `envconfig:"IMG_BUILDER_MAX_RETRY" default:"3"`
	K8sConfig     mlpcluster.K8sConfig `envconfig:"IMG_BUILDER_K8S_CONFIG"`
}

type Tolerations []v1.Toleration

func (spec *Tolerations) Decode(value string) error {
	var tolerations Tolerations

	if err := json.Unmarshal([]byte(value), &tolerations); err != nil {
		return err
	}
	*spec = tolerations
	return nil
}

type DictEnv map[string]string

func (d *DictEnv) Decode(value string) error {
	var dict DictEnv

	if err := json.Unmarshal([]byte(value), &dict); err != nil {
		return err
	}
	*d = dict
	return nil
}

type AuthorizationConfig struct {
	AuthorizationEnabled   bool   `envconfig:"AUTHORIZATION_ENABLED" default:"true"`
	AuthorizationServerURL string `envconfig:"AUTHORIZATION_SERVER_URL" default:"http://localhost:4466"`
}

type FeatureToggleConfig struct {
	MonitoringConfig MonitoringConfig
	AlertConfig      AlertConfig
}

type MonitoringConfig struct {
	MonitoringEnabled    bool   `envconfig:"MONITORING_DASHBOARD_ENABLED" default:"false"`
	MonitoringBaseURL    string `envconfig:"MONITORING_DASHBOARD_BASE_URL"`
	MonitoringJobBaseURL string `envconfig:"MONITORING_DASHBOARD_JOB_BASE_URL"`
}

type AlertConfig struct {
	AlertEnabled bool `envconfig:"ALERT_ENABLED" default:"false"`
	GitlabConfig GitlabConfig
	WardenConfig WardenConfig
}

type GitlabConfig struct {
	BaseURL             string `envconfig:"GITLAB_BASE_URL"`
	Token               string `envconfig:"GITLAB_TOKEN"`
	DashboardRepository string `envconfig:"GITLAB_DASHBOARD_REPOSITORY"`
	DashboardBranch     string `envconfig:"GITLAB_DASHBOARD_BRANCH" default:"master"`
	AlertRepository     string `envconfig:"GITLAB_ALERT_REPOSITORY"`
	AlertBranch         string `envconfig:"GITLAB_ALERT_BRANCH" default:"master"`
}

type WardenConfig struct {
	APIHost string `envconfig:"WARDEN_API_HOST"`
}

type MlpAPIConfig struct {
	APIHost       string `envconfig:"MLP_API_HOST" required:"true"`
	EncryptionKey string `envconfig:"MLP_API_ENCRYPTION_KEY" required:"true"`
}

type StandardTransformerConfig struct {
	ImageName             string               `envconfig:"STANDARD_TRANSFORMER_IMAGE_NAME" required:"true"`
	FeastServingURLs      FeastServingURLs     `envconfig:"FEAST_SERVING_URLS" required:"true"`
	FeastCoreURL          string               `envconfig:"FEAST_CORE_URL" required:"true"`
	FeastCoreAuthAudience string               `envconfig:"FEAST_CORE_AUTH_AUDIENCE" required:"true"`
	EnableAuth            bool                 `envconfig:"FEAST_AUTH_ENABLED" default:"false"`
	FeastRedisConfig      *FeastRedisConfig    `envconfig:"FEAST_REDIS_CONFIG"`
	FeastBigtableConfig   *FeastBigtableConfig `envconfig:"FEAST_BIG_TABLE_CONFIG"`
	// Base64 Service Account
	BigtableCredential string             `envconfig:"FEAST_BIGTABLE_CREDENTIAL"`
	DefaultFeastSource spec.ServingSource `envconfig:"DEFAULT_FEAST_SOURCE" default:"BIGTABLE"`
	Jaeger             JaegerConfig
	SimulationFeast    SimulationFeastConfig
}

// SimulationFeastConfig feast config that aimed to be used only for simulation of standard transformer
type SimulationFeastConfig struct {
	FeastRedisURL    string `envconfig:"SIMULATION_FEAST_REDIS_URL" required:"true"`
	FeastBigtableURL string `envconfig:"SIMULATION_FEAST_BIGTABLE_URL" required:"true"`
}

// ToFeastStorageConfigsForSimulation convert standard transformer config to feast storage config that will be used for transformer simulation
// the difference with ToFeastStorageConfigs, this method will overwrite serving url by using serving url that can be access outside of model k8s cluster and serving type is grpc (direct storage is not supported)
func (stc *StandardTransformerConfig) ToFeastStorageConfigsForSimulation() feast.FeastStorageConfig {
	storageCfg := stc.ToFeastStorageConfigs()
	servingURLs := map[spec.ServingSource]string{
		spec.ServingSource_REDIS:    stc.SimulationFeast.FeastRedisURL,
		spec.ServingSource_BIGTABLE: stc.SimulationFeast.FeastBigtableURL,
	}
	for sourceType, storage := range storageCfg {
		storage.ServingType = spec.ServingType_FEAST_GRPC
		if sourceType == spec.ServingSource_UNKNOWN {
			sourceType = stc.DefaultFeastSource
		}
		servingURL := servingURLs[sourceType]
		switch storage.Storage.(type) {
		case *spec.OnlineStorage_RedisCluster:
			redisCluster := storage.GetRedisCluster()
			redisCluster.FeastServingUrl = servingURL
		case *spec.OnlineStorage_Redis:
			redis := storage.GetRedis()
			redis.FeastServingUrl = servingURL
		case *spec.OnlineStorage_Bigtable:
			bigtable := storage.GetBigtable()
			bigtable.FeastServingUrl = servingURL
		}
	}
	return storageCfg
}

// ToFeastStorageConfigs convert standard transformer config into feast storage config
func (stc *StandardTransformerConfig) ToFeastStorageConfigs() feast.FeastStorageConfig {
	feastStorageConfig := feast.FeastStorageConfig{}
	validate, _ := internalValidator.NewValidator()

	// need to validate redis and big table config, because of `FeastRedisConfig` and `FeastBigtableConfig` wont be null when environment variables not set
	// this is due to bug in envconfig library https://github.com/kelseyhightower/envconfig/issues/113
	if stc.FeastRedisConfig != nil && validate.Struct(stc.FeastRedisConfig) == nil {
		feastStorageConfig[spec.ServingSource_REDIS] = stc.FeastRedisConfig.ToFeastStorage()
	}
	if stc.FeastBigtableConfig != nil && validate.Struct(stc.FeastBigtableConfig) == nil {
		feastStorageConfig[spec.ServingSource_BIGTABLE] = stc.FeastBigtableConfig.ToFeastStorageWithCredential(stc.BigtableCredential)
	}
	return feastStorageConfig
}

type FeastServingURLs []FeastServingURL

type FeastServingURL struct {
	Host       string `json:"host"`
	Label      string `json:"label"`
	Icon       string `json:"icon"`
	SourceType string `json:"source_type"`
}

func (u *FeastServingURLs) Decode(value string) error {
	var urls FeastServingURLs
	if err := json.Unmarshal([]byte(value), &urls); err != nil {
		return err
	}
	*u = urls
	return nil
}

func (u *FeastServingURLs) URLs() []string {
	urls := []string{}
	for _, url := range *u {
		urls = append(urls, url.Host)
	}
	return urls
}

type JaegerConfig struct {
	AgentHost    string `envconfig:"JAEGER_AGENT_HOST"`
	AgentPort    string `envconfig:"JAEGER_AGENT_PORT"`
	SamplerType  string `envconfig:"JAEGER_SAMPLER_TYPE" default:"probabilistic"`
	SamplerParam string `envconfig:"JAEGER_SAMPLER_PARAM" default:"0.01"`
	Disabled     string `envconfig:"JAEGER_DISABLED" default:"true"`
}

type MlflowConfig struct {
	TrackingURL string `envconfig:"MLFLOW_TRACKING_URL" required:"true"`
}

func InitConfigEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	cfg.EnvironmentConfigs = initEnvironmentConfigs(cfg.EnvironmentConfigPath)
	return &cfg, nil
}
