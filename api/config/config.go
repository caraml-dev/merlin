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

	DbConfig                  DatabaseConfig
	VaultConfig               VaultConfig
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
	OauthClientID     string         `envconfig:"REACT_APP_OAUTH_CLIENT_ID"`
	Environment       string         `envconfig:"REACT_APP_ENVIRONMENT"`
	SentryDSN         string         `envconfig:"REACT_APP_SENTRY_DSN"`
	DocURL            Documentations `envconfig:"REACT_APP_MERLIN_DOCS_URL"`
	AlertEnabled      bool           `envconfig:"REACT_APP_ALERT_ENABLED"`
	MonitoringEnabled bool           `envconfig:"REACT_APP_MONITORING_DASHBOARD_ENABLED"`
	HomePage          string         `envconfig:"REACT_APP_HOMEPAGE"`
	MerlinURL         string         `envconfig:"REACT_APP_MERLIN_API"`
	MlpURL            string         `envconfig:"REACT_APP_MLP_API"`
	FeastCoreURL      string         `envconfig:"REACT_APP_FEAST_CORE_API"`
	DockerRegistries  string         `envconfig:"REACT_APP_DOCKER_REGISTRIES"`
	MaxAllowedReplica int            `envconfig:"REACT_APP_MAX_ALLOWED_REPLICA" default:"20"`
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
}

type ImageBuilderConfig struct {
	ClusterName                  string `envconfig:"IMG_BUILDER_CLUSTER_NAME"`
	GcpProject                   string `envconfig:"IMG_BUILDER_GCP_PROJECT"`
	BuildContextURI              string `envconfig:"IMG_BUILDER_BUILD_CONTEXT_URI"`
	ContextSubPath               string `envconfig:"IMG_BUILDER_CONTEXT_SUB_PATH"`
	DockerfilePath               string `envconfig:"IMG_BUILDER_DOCKERFILE_PATH" default:"./Dockerfile"`
	BaseImage                    string `envconfig:"IMG_BUILDER_BASE_IMAGE"`
	PredictionJobBuildContextURI string `envconfig:"IMG_BUILDER_PREDICTION_JOB_BUILD_CONTEXT_URI"`
	PredictionJobContextSubPath  string `envconfig:"IMG_BUILDER_PREDICTION_JOB_CONTEXT_SUB_PATH"`
	PredictionJobDockerfilePath  string `envconfig:"IMG_BUILDER_PREDICTION_JOB_DOCKERFILE_PATH" default:"./Dockerfile"`
	PredictionJobBaseImage       string `envconfig:"IMG_BUILDER_PREDICTION_JOB_BASE_IMAGE"`
	BuildNamespace               string `envconfig:"IMG_BUILDER_NAMESPACE" default:"mlp"`
	DockerRegistry               string `envconfig:"IMG_BUILDER_DOCKER_REGISTRY"`
	BuildTimeout                 string `envconfig:"IMG_BUILDER_TIMEOUT" default:"10m"`
	KanikoImage                  string `envconfig:"IMG_BUILDER_KANIKO_IMAGE" default:"gcr.io/kaniko-project/executor:v1.6.0"`
	// How long to keep the image building job resource in the Kubernetes cluster. Default: 2 days (48 hours).
	Retention time.Duration       `envconfig:"IMG_BUILDER_RETENTION" default:"48h"`
	JobSpec   ImageBuilderJobSpec `envconfig:"IMG_BUILDER_JOB_SPEC"`
}

type ImageBuilderJobSpec struct {
	BackoffLimit            *int32                  `json:"backoffLimit,omitempty"`
	Volumes                 []v1.Volume             `json:"volumes,omitempty"`
	VolumeMounts            []v1.VolumeMount        `json:"volumeMounts,omitempty"`
	Env                     []v1.EnvVar             `json:"env,omitempty"`
	Resources               v1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations             []v1.Toleration         `json:"tolerations,omitempty"`
	NodeSelector            map[string]string       `json:"nodeSelectors,omitempty"`
	TTLSecondsAfterFinished *int32                  `json:"ttlSecondsAfterFinished,omitempty"`
}

func (spec *ImageBuilderJobSpec) Decode(value string) error {
	var jobSpec ImageBuilderJobSpec

	if err := json.Unmarshal([]byte(value), &jobSpec); err != nil {
		return err
	}
	*spec = jobSpec
	return nil
}

type VaultConfig struct {
	Address string `envconfig:"VAULT_ADDRESS"`
	Token   string `envconfig:"VAULT_TOKEN"`
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
	ImageName              string           `envconfig:"STANDARD_TRANSFORMER_IMAGE_NAME" required:"true"`
	DefaultFeastServingURL string           `envconfig:"DEFAULT_FEAST_SERVING_URL" required:"true"`
	FeastServingURLs       FeastServingURLs `envconfig:"FEAST_SERVING_URLS" required:"true"`
	FeastCoreURL           string           `envconfig:"FEAST_CORE_URL" required:"true"`
	FeastCoreAuthAudience  string           `envconfig:"FEAST_CORE_AUTH_AUDIENCE" required:"true"`
	EnableAuth             bool             `envconfig:"FEAST_AUTH_ENABLED" default:"false"`
	Jaeger                 JaegerConfig
}

type FeastServingURLs []FeastServingURL

type FeastServingURL struct {
	Host  string `json:"host"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
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
