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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
	sigyaml "sigs.k8s.io/yaml"

	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	internalValidator "github.com/caraml-dev/merlin/pkg/validator"
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/ory/viper"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	MaxDeployedVersion = 2
)

type Config struct {
	Environment           string `default:"dev"`
	Port                  int    `default:"8080"`
	LoggerDestinationURL  string
	Sentry                sentry.Config
	NewRelic              newrelic.Config
	EnvironmentConfigPath string `validate:"required"`
	NumOfQueueWorkers     int    `default:"2"`
	SwaggerPath           string `default:"./swagger.yaml"`

	DeploymentLabelPrefix string `default:"gojek.com/"`
	PyfuncGRPCOptions     string `default:"{}"`

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
	StaticPath string `default:"ui/build"`
	IndexPath  string `default:"index.html"`
}

type ReactAppConfig struct {
	DocURL            Documentations `json:"REACT_APP_MERLIN_DOCS_URL,omitempty"`
	DockerRegistries  string         `json:"REACT_APP_DOCKER_REGISTRIES,omitempty"`
	Environment       string         `json:"REACT_APP_ENVIRONMENT,omitempty"`
	FeastCoreURL      string         `json:"REACT_APP_FEAST_CORE_API,omitempty"`
	HomePage          string         `json:"REACT_APP_HOMEPAGE,omitempty"`
	MaxAllowedReplica int            `default:"20" json:"REACT_APP_MAX_ALLOWED_REPLICA,omitempty"`
	MerlinURL         string         `json:"REACT_APP_MERLIN_API,omitempty"`
	MlpURL            string         `json:"REACT_APP_MLP_API,omitempty"`
	OauthClientID     string         `json:"REACT_APP_OAUTH_CLIENT_ID,omitempty"`
	SentryDSN         string         `json:"REACT_APP_SENTRY_DSN,omitempty"`
	UPIDocumentation  string         `json:"REACT_APP_UPI_DOC_URL,omitempty"`
	CPUCost           string         `json:"REACT_APP_CPU_COST,omitempty"`
	MemoryCost        string         `json:"REACT_APP_MEMORY_COST,omitempty"`
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
	Host          string `validate:"required"`
	Port          int    `default:"5432"`
	User          string `validate:"required"`
	Password      string `validate:"required"`
	Database      string `default:"mlp"`
	MigrationPath string `default:"file://db-migrations"`

	ConnMaxIdleTime time.Duration `default:"0s"`
	ConnMaxLifetime time.Duration `default:"0s"`
	MaxIdleConns    int           `default:"0"`
	MaxOpenConns    int           `default:"0"`
}

// Resource contains the Kubernetes resource request and limits
type Resource struct {
	CPU    string `validate:"required"`
	Memory string `validate:"required"`
}

// ResourceRequestsLimits contains the Kubernetes resource request and limits for kaniko
type ResourceRequestsLimits struct {
	Requests Resource `validate:"required"`
	Limits   Resource `validate:"required"`
}

func (r *ResourceRequestsLimits) Decode(value string) error {
	var resourceRequestsLimits ResourceRequestsLimits
	var err error

	if err := json.Unmarshal([]byte(value), &resourceRequestsLimits); err != nil {
		return err
	}
	*r = resourceRequestsLimits

	// Validate that the quantity representation is valid
	_, err = resourcev1.ParseQuantity(r.Requests.CPU)
	if err != nil {
		return err
	}
	_, err = resourcev1.ParseQuantity(r.Requests.Memory)
	if err != nil {
		return err
	}
	_, err = resourcev1.ParseQuantity(r.Limits.CPU)
	if err != nil {
		return err
	}
	_, err = resourcev1.ParseQuantity(r.Limits.Memory)
	if err != nil {
		return err
	}

	return nil
}

type ImageBuilderConfig struct {
	ClusterName                  string
	GcpProject                   string
	BuildContextURI              string
	ContextSubPath               string
	DockerfilePath               string `default:"./Dockerfile"`
	BaseImages                   BaseImageConfigs
	PredictionJobBuildContextURI string
	PredictionJobContextSubPath  string
	PredictionJobDockerfilePath  string `default:"./Dockerfile"`
	PredictionJobBaseImages      BaseImageConfigs
	BuildNamespace               string `default:"mlp"`
	DockerRegistry               string
	BuildTimeout                 string `default:"10m"`
	KanikoImage                  string `default:"gcr.io/kaniko-project/executor:v1.6.0"`
	KanikoServiceAccount         string
	Resources                    ResourceRequestsLimits
	// How long to keep the image building job resource in the Kubernetes cluster. Default: 2 days (48 hours).
	Retention     time.Duration `default:"48h"`
	Tolerations   Tolerations
	NodeSelectors DictEnv
	MaximumRetry  int32 `default:"3"`
	K8sConfig     mlpcluster.K8sConfig
	SafeToEvict   bool `default:"false"`
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
	AuthorizationEnabled   bool                 `default:"true"`
	AuthorizationServerURL string               `default:"http://localhost:4466"`
	Caching                *InMemoryCacheConfig `Validate:"required_if=AuthorizationEnabled True"`
}

type InMemoryCacheConfig struct {
	Enabled                     bool
	KeyExpirySeconds            int  `Validate:"required_if=Enabled True" default:"600"`
	CacheCleanUpIntervalSeconds int  `Validate:"required_if=Enabled True" default:"900"`
}

type FeatureToggleConfig struct {
	MonitoringConfig MonitoringConfig
	AlertConfig      AlertConfig
}

type MonitoringConfig struct {
	MonitoringEnabled    bool `default:"false"`
	MonitoringBaseURL    string
	MonitoringJobBaseURL string
}

type AlertConfig struct {
	AlertEnabled bool `default:"false"`
	GitlabConfig GitlabConfig
	WardenConfig WardenConfig
}

type GitlabConfig struct {
	BaseURL             string
	Token               string
	DashboardRepository string
	DashboardBranch     string `default:"master"`
	AlertRepository     string
	AlertBranch         string `default:"master"`
}

type WardenConfig struct {
	APIHost string
}

type MlpAPIConfig struct {
	APIHost       string `validate:"required"`
}

// FeastServingKeepAliveConfig config for feast serving grpc keepalive
type FeastServingKeepAliveConfig struct {
	// Enable the client grpc keepalive
	Enabled bool `default:"false"`
	// Duration of time no activity until client try to PING gRPC server
	Time time.Duration `default:"60s"`
	// Duration of time client waits if no activity connection will be closed
	Timeout time.Duration `default:"1s"`
}

// ModelClientKeepAliveConfig config for merlin model predictor grpc keepalive
type ModelClientKeepAliveConfig struct {
	// Enable the client grpc keepalive
	Enabled bool `default:"false"`
	// Duration of time no activity until client try to PING gRPC server
	Time time.Duration `default:"60s"`
	// Duration of time client waits if no activity connection will be closed
	Timeout time.Duration `default:"5s"`
}

type StandardTransformerConfig struct {
	ImageName             string           `validate:"required"`
	FeastServingURLs      FeastServingURLs `validate:"required"`
	FeastCoreURL          string           `validate:"required"`
	FeastCoreAuthAudience string           `validate:"required"`
	EnableAuth            bool             `default:"false"`
	FeastRedisConfig      *FeastRedisConfig
	FeastBigtableConfig   *FeastBigtableConfig
	FeastGPRCConnCount    int                  `default:"10"`
	FeastServingKeepAlive *FeastServingKeepAliveConfig
	ModelClientKeepAlive  *ModelClientKeepAliveConfig
	ModelServerConnCount  int `default:"10"`
	// Base64 Service Account
	BigtableCredential string
	DefaultFeastSource spec.ServingSource `default:"BIGTABLE"`
	Jaeger             JaegerConfig
	SimulationFeast    SimulationFeastConfig
	Kafka              KafkaConfig
}

// Kafka configuration for publishing prediction log
type KafkaConfig struct {
	Topic               string
	Brokers             string
	CompressionType     string `default:"none"`
	MaxMessageSizeBytes int    `default:"1048588"`
	ConnectTimeoutMS    int    `default:"1000"`
	SerializationFmt    string `default:"protobuf"`
}

// SimulationFeastConfig feast config that aimed to be used only for simulation of standard transformer
type SimulationFeastConfig struct {
	FeastRedisURL    string `validate:"required"`
	FeastBigtableURL string `validate:"required"`
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

	// need to Validate redis and big table config, because of `FeastRedisConfig` and `FeastBigtableConfig` wont be null when environment variables not set
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
	AgentHost    string
	AgentPort    string
	SamplerType  string `default:"probabilistic"`
	SamplerParam string `default:"0.01"`
	Disabled     string `default:"true"`
}

type MlflowConfig struct {
	TrackingURL string `validate:"required"`
	ArtifactServiceType string `validate:"required"`
}

func InitConfigEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	cfg.EnvironmentConfigs = InitEnvironmentConfigs(cfg.EnvironmentConfigPath)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) Validate() error {
	v := validator.New()
	err := v.Struct(cfg)
	if err != nil {
		return err
	}
	// Validate pyfunc server keep alive config, it must be string in json format
	var pyfuncGRPCOpts json.RawMessage
	if err := json.Unmarshal([]byte(cfg.PyfuncGRPCOptions), &pyfuncGRPCOpts); err != nil {
		return err
	}
	return nil
}

// Load creates a Config object from default config values, config files and environment variables.
// Load accepts config files as the argument. JSON and YAML format are both supported.
//
// If multiple config files are provided, the subsequent config files will override the config
// values from the config files loaded earlier.
//
// These config files will override the default config values (refer to setDefaultValues function)
// and can be overridden by the values from environment variables. Nested keys in the config
// can be set from environment variable name separed by "_". For instance the config value for
// "DbConfig.Port" can be overridden by environment variable name "DBCONFIG_PORT". Note that
// all environment variable names must be upper case.
//
// If no config file is provided, only the default config values and config values from environment
// varibales will be loaded.
//
// Refer to example.yaml for an example of config file.
func Load(filepaths ...string) (*Config, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))

	// Load default config values
	setDefaultValues(v)

	// Load config values from the provided config files
	for _, f := range filepaths {
		v.SetConfigFile(f)
		err := v.MergeInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config from file '%s': %w", f, err)
		}
	}

	// Load config values from environment variables.
	// Nested keys in the config is represented by variable name separated by '_'.
	// For example, DbConfig.Host can be set from environment variable DBCONFIG_HOST.
	v.SetEnvKeyReplacer(strings.NewReplacer("::", "_"))
	v.AutomaticEnv()

	config := &Config{}

	// Unmarshal config values into the config object.
	// Add StringToQuantityHookFunc() to the default DecodeHook in order to parse quantity string
	// into quantity object. Refs:
	// https://github.com/spf13/viper/blob/493643fd5e4b44796124c05d59ee04ba5f809e19/viper.go#L1003-L1005
	// https://github.com/mitchellh/mapstructure/blob/9e1e4717f8567d7ead72d070d064ad17d444a67e/decode_hooks_test.go#L128
	err := v.Unmarshal(config, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			StringToQuantityHookFunc(),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config values: %w", err)
	}
	config, err = loadImageBuilderConfig(config, v.AllSettings())
	if err != nil {
		return nil, fmt.Errorf("Failed to load imagebuilderconfig.k8sconfig, err %s", err)
	}

	return config, nil
}

func loadImageBuilderConfig(config *Config, v map[string]interface{}) (*Config, error) {
	// NOTE: This section is added to parse any fields in ImageBuilderConfig.K8sConfig that does not
	// have yaml tags.
	// For example `certificate-authority-data` is not unmarshalled
	// by vipers unmarshal method.

	clusterConfig, ok := v["imagebuilderconfig"]
	if !ok {
		return config, nil
	}
	contents := clusterConfig.(map[string]interface{})
	imageBuilderK8sCfg, ok := contents["k8sconfig"]
	if !ok {
		return config, nil
	}
	// convert back to byte string
	var byteForm []byte
	byteForm, err := yaml.Marshal(imageBuilderK8sCfg)
	if err != nil {
		return nil, err
	}
	// use sigyaml.Unmarshal to convert to json object then unmarshal

	k8sConfig := mlpcluster.K8sConfig{}
	if err := sigyaml.Unmarshal(byteForm, &k8sConfig); err != nil {
		return nil, err
	}
	config.ImageBuilderConfig.K8sConfig = k8sConfig
	return config, nil
}

// setDefaultValues for all keys in Viper config. We need to set values for all keys so that
// we can always use environment variables to override the config keys. In Viper v1, if the
// keys do not have default values, and the key does not appear in the config file, it cannot
// be overridden by environment variables, unless each key is called with BindEnv.
// https://github.com/spf13/viper/issues/188
// https://github.com/spf13/viper/issues/761
func setDefaultValues(v *viper.Viper) {
	v.SetDefault("Environment", "dev")
	v.SetDefault("Port", "8080")
	v.SetDefault("NumOfQueueWorkers", "2")
	v.SetDefault("SwaggerPath", "./swagger.yaml")

	v.SetDefault("DeploymentLabelPrefix", "gojek.com/")
	v.SetDefault("PyfuncGRPCOptions", "{}")

	v.SetDefault("DatabaseConfig::Port", "false")
	v.SetDefault("AuthConfig::URL", "")

	//v.SetDefault("DbConfig::Host", "localhost")
	v.SetDefault("DbConfig::Port", "5432")
	//v.SetDefault("DbConfig::User", "")
	//v.SetDefault("DbConfig::Password", "")
	v.SetDefault("DbConfig::Database", "mlp")
	v.SetDefault("DbConfig::MigrationPath", "file://db-migrations")
	v.SetDefault("DbConfig::ConnMaxIdleTime", "0s")
	v.SetDefault("DbConfig::ConnMaxLifetime", "0s")
	v.SetDefault("DbConfig::MaxIdleConns", "0")
	v.SetDefault("DbConfig::MaxOpenConns", "0")

	v.SetDefault("ImageBuilderConfig::DockerfilePath", "./Dockerfile")
	v.SetDefault("ImageBuilderConfig::PredictionJobDockerfilePath", "./Dockerfile")
	v.SetDefault("ImageBuilderConfig::BuildNamespace", "mlp")
	v.SetDefault("ImageBuilderConfig::BuildTimeout", "10m")
	v.SetDefault("ImageBuilderConfig::KanikoImage", "gcr.io/kaniko-project/executor:v1.6.0")
	v.SetDefault("ImageBuilderConfig::Retention", "48h")
	v.SetDefault("ImageBuilderConfig::MaximumRetry", "3")
	v.SetDefault("ImageBuilderConfig::SafeToEvict", "false")

	v.SetDefault("AuthorizationConfig::AuthorizationEnabled", "true")
	v.SetDefault("AuthorizationConfig::AuthorizationServerURL", "http://localhost:4466")

	v.SetDefault("FeatureToggleConfig::MonitoringConfig::MonitoringEnabled", "false")
	v.SetDefault("FeatureToggleConfig::AlertConfig::AlertEnabled", "false")
	v.SetDefault("FeatureToggleConfig::AlertConfig::GitlabConfig::DashboardBranch", "master")
	v.SetDefault("FeatureToggleConfig::AlertConfig::GitlabConfig::AlertBranch", "master")

	v.SetDefault("ReactAppConfig::MaxAllowedReplica", "20")

	v.SetDefault("UI::StaticPath", "ui/build")
	v.SetDefault("UI::IndexPath", "index.html")

	v.SetDefault("StandardTransformerConfig::EnableAuth", "false")
	v.SetDefault("StandardTransformerConfig::FeastBigtableConfig::IsUsingDirectStorage", "false")
	v.SetDefault("StandardTransformerConfig::FeastServingKeepAlive::Enabled", "false")
	v.SetDefault("StandardTransformerConfig::FeastServingKeepAlive::Time", "60s")
	v.SetDefault("StandardTransformerConfig::FeastServingKeepAlive::Timeout", "1s")
	v.SetDefault("StandardTransformerConfig::ModelClientKeepAlive::Timeout", "false")
	v.SetDefault("StandardTransformerConfig::ModelClientKeepAlive::Timeout", "60s")
	v.SetDefault("StandardTransformerConfig::ModelClientKeepAlive::Timeout", "5s")
	v.SetDefault("StandardTransformerConfig::BigtableCredential", "")
	v.SetDefault("StandardTransformerConfig::Jaeger::SamplerType", "probabilistic")
	v.SetDefault("StandardTransformerConfig::Jaeger::SamplerParam", "0.01")
	v.SetDefault("StandardTransformerConfig::Jaeger::Disabled", "true")
	v.SetDefault("StandardTransformerConfig::SimulationFeast::FeastRedisURL", "")
	v.SetDefault("StandardTransformerConfig::SimulationFeast::FeastBigtableURL", "")
	v.SetDefault("StandardTransformerConfig::Kafka::CompressionType", "none")
	v.SetDefault("StandardTransformerConfig::Kafka::MaxMessageSizeBytes", "1048588")
	v.SetDefault("StandardTransformerConfig::Kafka::ConnectTimeoutMS", "1000")
	v.SetDefault("StandardTransformerConfig::Kafka::SerializationFmt", "protobuf")

	v.SetDefault("Sentry::Enabled", "false")
	v.SetDefault("Sentry::DSN", "")
}

// Quantity is an alias for resource.Quantity
type Quantity resource.Quantity

// StringToQuantityHookFunc converts string to quantity type. This function is required since
// viper uses mapstructure to unmarshal values.
// https://github.com/spf13/viper#unmarshaling
// https://pkg.go.dev/github.com/mitchellh/mapstructure#DecodeHookFunc
func StringToQuantityHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(Quantity{}) {
			return data, nil
		}

		// Convert it by parsing
		q, err := resource.ParseQuantity(data.(string))
		if err != nil {
			return nil, err
		}

		return Quantity(q), nil
	}
}
