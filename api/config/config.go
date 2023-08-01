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

	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	internalValidator "github.com/caraml-dev/merlin/pkg/validator"
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/ory/viper"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
)

const (
	viperKeyDelimiter = "::"

	MaxDeployedVersion = 2
)

type Config struct {
	Environment          string `validate:"required" default:"dev"`
	Port                 int    `validate:"required" default:"8080"`
	LoggerDestinationURL string
	Sentry               sentry.Config
	NewRelic             newrelic.Config
	NumOfQueueWorkers    int    `validate:"required" default:"2"`
	SwaggerPath          string `validate:"required" default:"./swagger.yaml"`

	DeploymentLabelPrefix string `validate:"required" default:"gojek.com/"`
	PyfuncGRPCOptions     string `validate:"required" default:"{}"`

	DbConfig                  DatabaseConfig     `validate:"required"`
	ClusterConfig             ClusterConfig      `validate:"required"`
	ImageBuilderConfig        ImageBuilderConfig `validate:"required"`
	AuthorizationConfig       AuthorizationConfig
	MlpAPIConfig              MlpAPIConfig `validate:"required"`
	FeatureToggleConfig       FeatureToggleConfig
	ReactAppConfig            ReactAppConfig
	UI                        UIConfig
	StandardTransformerConfig StandardTransformerConfig
	MlflowConfig              MlflowConfig
}

// UIConfig stores the configuration for the UI.
type UIConfig struct {
	StaticPath string `validate:"required" default:"ui/build"`
	IndexPath  string `validate:"required" default:"index.html"`
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
	ImageName string `validate:"required" json:"imageName"`
	// Dockerfile Path within the build context
	DockerfilePath string `validate:"required" json:"dockerfilePath"`
	// GCS URL Containing build context
	BuildContextURI string `validate:"required" json:"buildContextURI"`
	// path to main file to run application
	MainAppPath string `validate:"required" json:"mainAppPath"`
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
	Port          int    `validate:"required" default:"5432"`
	User          string `validate:"required"`
	Password      string `validate:"required"`
	Database      string `validate:"required" default:"mlp"`
	MigrationPath string `validate:"required" default:"file://db-migrations"`

	ConnMaxIdleTime time.Duration `default:"0s"`
	ConnMaxLifetime time.Duration `default:"0s"`
	MaxIdleConns    int
	MaxOpenConns    int
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

// ClusterConfig contains the cluster controller information.
// Supported features are in cluster configuration and Kubernetes client CA certificates.
type ClusterConfig struct {
	// InClusterConfig is a flag if the service account is provided in Kubernetes
	// and has the relevant credentials to handle all cluster operations.
	InClusterConfig bool

	// EnvironmentConfigPath refers to a path that contains EnvironmentConfigs
	EnvironmentConfigPath string `validate:"required_without=InClusterConfig"`
	EnvironmentConfigs    []*EnvironmentConfig
}

type ImageBuilderConfig struct {
	ClusterName                 string `validate:"required"`
	GcpProject                  string
	ContextSubPath              string
	DockerfilePath              string           `validate:"required" default:"./Dockerfile"`
	BaseImages                  BaseImageConfigs `validate:"required"`
	PredictionJobContextSubPath string
	PredictionJobDockerfilePath string           `validate:"required" default:"./Dockerfile"`
	PredictionJobBaseImages     BaseImageConfigs `validate:"required"`
	BuildNamespace              string           `validate:"required" default:"mlp"`
	DockerRegistry              string           `validate:"required"`
	BuildTimeout                string           `validate:"required" default:"10m"`
	KanikoImage                 string           `validate:"required" default:"gcr.io/kaniko-project/executor:v1.6.0"`
	KanikoServiceAccount        string
	Resources                   ResourceRequestsLimits `validate:"required"`
	// How long to keep the image building job resource in the Kubernetes cluster. Default: 2 days (48 hours).
	Retention     time.Duration `validate:"required" default:"48h"`
	Tolerations   Tolerations
	NodeSelectors map[string]string
	MaximumRetry  int32                 `validate:"required" default:"3"`
	K8sConfig     *mlpcluster.K8sConfig `validate:"required" default:"-"`
	SafeToEvict   bool                  `default:"false"`
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

type AuthorizationConfig struct {
	AuthorizationEnabled bool                 `default:"true"`
	KetoRemoteRead       string               `default:"http://localhost:4466"`
	KetoRemoteWrite      string               `default:"http://localhost:4467"`
	Caching              *InMemoryCacheConfig `validate:"required_if=AuthorizationEnabled True"`
}

type InMemoryCacheConfig struct {
	Enabled                     bool
	KeyExpirySeconds            int `validate:"required_if=Enabled True" default:"600"`
	CacheCleanUpIntervalSeconds int `validate:"required_if=Enabled True" default:"900"`
}

type FeatureToggleConfig struct {
	MonitoringConfig    MonitoringConfig
	AlertConfig         AlertConfig
	ModelDeletionConfig ModelDeletionConfig
}

type MonitoringConfig struct {
	MonitoringEnabled    bool   `default:"false"`
	MonitoringBaseURL    string `validate:"required_if=Enabled True"`
	MonitoringJobBaseURL string `validate:"required_if=Enabled True"`
}

type AlertConfig struct {
	AlertEnabled bool `default:"false"`
	GitlabConfig GitlabConfig
	WardenConfig WardenConfig
}

type ModelDeletionConfig struct {
	Enabled bool `default:"false"`
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
	APIHost string `validate:"required"`
}

// FeastServingKeepAliveConfig config for feast serving grpc keepalive
type FeastServingKeepAliveConfig struct {
	// Enable the client grpc keepalive
	Enabled bool `default:"false"`
	// Duration of time no activity until client try to PING gRPC server
	Time time.Duration `validate:"required_if=Enabled True" default:"60s"`
	// Duration of time client waits if no activity connection will be closed
	Timeout time.Duration `validate:"required_if=Enabled True" default:"1s"`
}

// ModelClientKeepAliveConfig config for merlin model predictor grpc keepalive
type ModelClientKeepAliveConfig struct {
	// Enable the client grpc keepalive
	Enabled bool `default:"false"`
	// Duration of time no activity until client try to PING gRPC server
	Time time.Duration `validate:"required_if=Enabled True" default:"60s"`
	// Duration of time client waits if no activity connection will be closed
	Timeout time.Duration `validate:"required_if=Enabled True" default:"5s"`
}

type StandardTransformerConfig struct {
	ImageName             string `validate:"required"`
	FeastServingURLs      FeastServingURLs
	FeastCoreURL          string               `validate:"required"`
	FeastCoreAuthAudience string               `validate:"required"`
	EnableAuth            bool                 `default:"false"`
	FeastRedisConfig      *FeastRedisConfig    `default:"-"`
	FeastBigtableConfig   *FeastBigtableConfig `default:"-"`
	FeastGPRCConnCount    int                  `validate:"required" default:"10"`
	FeastServingKeepAlive *FeastServingKeepAliveConfig
	ModelClientKeepAlive  *ModelClientKeepAliveConfig
	ModelServerConnCount  int `validate:"required" default:"10"`
	// Base64 Service Account
	BigtableCredential string
	DefaultFeastSource spec.ServingSource    `validate:"required" default:"2"`
	Jaeger             JaegerConfig          `validate:"required"`
	SimulationFeast    SimulationFeastConfig `validate:"required"`
	Kafka              KafkaConfig           `validate:"required"`
}

// KafkaConfig configuration for publishing prediction log
type KafkaConfig struct {
	Topic               string
	Brokers             string `validate:"required"`
	CompressionType     string `validate:"required" default:"none"`
	MaxMessageSizeBytes int    `validate:"required" default:"1048588"`
	ConnectTimeoutMS    int    `validate:"required" default:"1000"`
	SerializationFmt    string `validate:"required" default:"protobuf"`
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
	CollectorURL string `validate:"required"`
	SamplerParam string `validate:"required" default:"0.01"`
	Disabled     string `validate:"required" default:"true"`
}

type MlflowConfig struct {
	TrackingURL         string `validate:"required"`
	ArtifactServiceType string `validate:"required"`
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
// can be set from environment variable name separated by "_". For instance the config value for
// "DbConfig.Port" can be overridden by environment variable name "DBCONFIG_PORT". Note that
// all environment variable names must be upper case.
//
// If no config file is provided, only the default config values and config values from environment
// variables will be loaded.
//
// Refer to example.yaml for an example of config file.
func Load(spec interface{}, filepaths ...string) (*Config, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter(viperKeyDelimiter))

	err := reflectViperConfig("", spec, v)
	if err != nil {
		return nil, fmt.Errorf("failed to read default config via reflection: %w", err)
	}

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
	v.SetEnvKeyReplacer(strings.NewReplacer(viperKeyDelimiter, "_"))
	v.AutomaticEnv()

	// Unmarshal config values into the config object.
	// Add StringToQuantityHookFunc() to the default DecodeHook in order to parse quantity string
	// into quantity object. Refs:
	// https://github.com/spf13/viper/blob/493643fd5e4b44796124c05d59ee04ba5f809e19/viper.go#L1003-L1005
	// https://github.com/mitchellh/mapstructure/blob/9e1e4717f8567d7ead72d070d064ad17d444a67e/decode_hooks_test.go#L128
	err = v.Unmarshal(spec, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			StringToQuantityHookFunc(),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config values: %w", err)
	}

	config, ok := spec.(*Config)
	if !ok {
		return nil, fmt.Errorf("failed to parse config values into Config object")
	}
	config.ImageBuilderConfig.K8sConfig, err = loadImageBuilderConfig(v.AllSettings())
	if err != nil {
		return nil, fmt.Errorf("failed to load imagebuilderconfig.k8sconfig, err %w", err)
	}

	return config, nil
}

func reflectViperConfig(prefix string, spec interface{}, v *viper.Viper) error {
	s := reflect.ValueOf(spec)
	s = s.Elem()
	typeOfSpec := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)

		viperKey := ftype.Name
		// Nested struct tags
		if prefix != "" {
			viperKey = fmt.Sprintf("%s%s%s", prefix, viperKeyDelimiter, ftype.Name)
		}
		value := ftype.Tag.Get("default")
		if value == "-" {
			continue
		}
		v.SetDefault(viperKey, value)
		// Create dynamic map using reflection
		if ftype.Type.Kind() == reflect.Map {
			mapValue := reflect.MakeMapWithSize(ftype.Type, 0)
			v.SetDefault(viperKey, mapValue)
		}

		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					// nil pointer to a non-struct: leave it alone
					break
				}
				// nil pointer to struct: create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		if f.Kind() == reflect.Struct {
			// Capture information about the config parent prefix
			parentPrefix := prefix
			if !ftype.Anonymous {
				parentPrefix = viperKey
			}

			// Use recursion to resolve nested config
			nestedPtr := f.Addr().Interface()
			err := reflectViperConfig(parentPrefix, nestedPtr, v)
			if err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

func loadImageBuilderConfig(v map[string]interface{}) (*mlpcluster.K8sConfig, error) {
	// NOTE: This section is added to parse any fields in ImageBuilderConfig.K8sConfig that does not
	// have yaml tags.
	// For example `certificate-authority-data` is not unmarshalled
	// by vipers unmarshal method.
	clusterConfig, ok := v["imagebuilderconfig"]
	if !ok {
		return nil, nil
	}
	contents := clusterConfig.(map[string]interface{})
	imageBuilderK8sCfg, ok := contents["k8sconfig"]
	if !ok {
		return nil, nil
	}
	// convert back to byte string
	var byteForm []byte
	byteForm, err := yaml.Marshal(imageBuilderK8sCfg)
	if err != nil {
		return nil, err
	}
	k8sConfig := mlpcluster.K8sConfig{}
	if err := yaml.Unmarshal(byteForm, &k8sConfig); err != nil {
		return nil, err
	}
	return &k8sConfig, nil
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
