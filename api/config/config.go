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
	"github.com/kelseyhightower/envconfig"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
)

const (
	MaxDeployedVersion = 2
)

type Config struct {
	Environment          string `envconfig:"ENVIRONMENT" default:"dev"`
	Port                 int    `envconfig:"PORT" default:"8080"`
	LoggerDestinationURL string `envconfig:"LOGGER_DESTINATION_URL"`

	DbConfig DatabaseConfig
	Sentry   sentry.Config   `split_words:"false" envconfig:"SENTRY"`
	NewRelic newrelic.Config `split_words:"false" envconfig:"NEWRELIC"`

	ImageBuilderConfig ImageBuilderConfig
	VaultConfig        VaultConfig

	EnvironmentConfigPath string `envconfig:"DEPLOYMENT_CONFIG_PATH" required:"true"`
	EnvironmentConfigs    []EnvironmentConfig
	AuthorizationConfig   AuthorizationConfig

	MlpAPIConfig MlpAPIConfig

	FeatureToggleConfig FeatureToggleConfig

	ReactAppConfig ReactAppConfig

	UI UIConfig
}

// UIConfig stores the configuration for the UI.
type UIConfig struct {
	StaticPath string `envconfig:"UI_STATIC_PATH" default:"ui/build"`
	IndexPath  string `envconfig:"UI_INDEX_PATH" default:"index.html"`
}

type ReactAppConfig struct {
	OauthClientID     string `envconfig:"REACT_APP_OAUTH_CLIENT_ID"`
	Environment       string `envconfig:"REACT_APP_ENVIRONMENT"`
	SentryDSN         string `envconfig:"REACT_APP_SENTRY_DSN"`
	DocURL            string `envconfig:"REACT_APP_MERLIN_DOCS_URL"`
	AlertEnabled      bool   `envconfig:"REACT_APP_ALERT_ENABLED"`
	MonitoringEnabled bool   `envconfig:"REACT_APP_MONITORING_DASHBOARD_ENABLED"`
	HomePage          string `envconfig:"REACT_APP_HOMEPAGE"`
	MerlinURL         string `envconfig:"REACT_APP_MERLIN_API"`
	MlpURL            string `envconfig:"REACT_APP_MLP_API"`
	DockerRegistries  string `envconfig:"REACT_APP_DOCKER_REGISTRIES"`
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

func InitConfigEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	cfg.EnvironmentConfigs = initEnvironmentConfigs(cfg.EnvironmentConfigPath)
	return &cfg, nil
}
