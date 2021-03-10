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

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"

	"github.com/gojek/merlin/api"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/cronjob"
	"github.com/gojek/merlin/gitlab"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlflow"
	"github.com/gojek/merlin/queue"
	"github.com/gojek/merlin/service"
	"github.com/gojek/merlin/storage"
	"github.com/gojek/merlin/warden"
)

func main() {
	ctx := context.Background()

	cfg, err := config.InitConfigEnv()
	if err != nil {
		log.Panicf("Failed initializing config: %v", err)
	}

	// Initializing Sentry client
	cfg.Sentry.Labels = map[string]string{"environment": cfg.Environment}
	if err := sentry.InitSentry(cfg.Sentry); err != nil {
		log.Errorf("Failed to initialize sentry client: %s", err)
	}
	defer sentry.Close()

	// Initializing NewRelic
	if err := newrelic.InitNewRelic(cfg.NewRelic); err != nil {
		log.Errorf("Failed to initialize newrelic: %s", err)
	}
	defer newrelic.Shutdown(5 * time.Second)

	db, dbDeferFunc := initDB(cfg.DbConfig)
	defer dbDeferFunc()

	runDBMigration(db, cfg.DbConfig.MigrationPath)

	appCtx := initAppContext(ctx, cfg, db)
	tracker, err := cronjob.NewTracker(appCtx.ProjectsService,
		appCtx.ModelsService,
		storage.NewPredictionJobStorage(db),
		storage.NewDeploymentStorage(db))
	if err != nil {
		log.Panicf("unable to create tracker %v", err)
	}
	tracker.Start()

	router := mux.NewRouter()

	mount(router, "/v1/internal", healthcheck.NewHandler())
	mount(router, "/v1", api.NewRouter(appCtx))
	mount(router, "/metrics", promhttp.Handler())

	reactConfig := cfg.ReactAppConfig
	uiEnv := uiEnvHandler{
		OauthClientID:    reactConfig.OauthClientID,
		Environment:      reactConfig.Environment,
		SentryDSN:        reactConfig.SentryDSN,
		DocURL:           reactConfig.DocURL,
		HomePage:         reactConfig.HomePage,
		MerlinURL:        reactConfig.MerlinURL,
		MlpURL:           reactConfig.MlpURL,
		FeastCoreURL:     reactConfig.FeastCoreURL,
		DockerRegistries: reactConfig.DockerRegistries,

		MonitoringEnabled:              cfg.FeatureToggleConfig.MonitoringConfig.MonitoringEnabled,
		MonitoringPredictionJobBaseURL: cfg.FeatureToggleConfig.MonitoringConfig.MonitoringJobBaseURL,

		AlertEnabled: cfg.FeatureToggleConfig.AlertConfig.AlertEnabled,
	}

	uiHomePage := reactConfig.HomePage
	if !strings.HasPrefix(uiHomePage, "/") {
		uiHomePage = "/" + uiEnv.HomePage
	}
	router.Path(uiHomePage + "/env.js").HandlerFunc(uiEnv.handler)

	ui := uiHandler{staticPath: cfg.UI.StaticPath, indexPath: cfg.UI.IndexPath}
	router.PathPrefix(uiHomePage).Handler(http.StripPrefix(uiHomePage, ui))

	log.Infof("listening at port :%d", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), cors.AllowAll().Handler(router))
}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}

func initAppContext(ctx context.Context, cfg *config.Config, db *gorm.DB) api.AppContext {
	mlpAPIClient := initMLPAPIClient(ctx, cfg.MlpAPIConfig)
	coreClient := initFeastCoreClient(cfg.StandardTransformerConfig.FeastCoreURL)

	vaultClient := initVault(cfg.VaultConfig)
	webServiceBuilder, predJobBuilder := initImageBuilder(cfg, vaultClient)

	modelEndpointService := initModelEndpointService(cfg, vaultClient, db)
	versionEndpointService := initVersionEndpointService(cfg, webServiceBuilder, vaultClient, db)
	predictionJobService := initPredictionJobService(cfg, mlpAPIClient, predJobBuilder, vaultClient, db)
	logService := initLogService(cfg, vaultClient)

	// use "mlp" as product name for enforcer so that same policy can be reused by excalibur
	authEnforcer, err := enforcer.NewEnforcerBuilder().
		URL(cfg.AuthorizationConfig.AuthorizationServerURL).
		Product("mlp").
		Build()
	if err != nil {
		log.Panicf("unable to initialize authorization enforcer %v", err)
	}

	projectsService := service.NewProjectsService(mlpAPIClient)
	modelsService := service.NewModelsService(db, mlpAPIClient)
	versionsService := service.NewVersionsService(db, mlpAPIClient)
	environmentService := initEnvironmentService(cfg, db)
	secretService := service.NewSecretService(mlpAPIClient)

	gitlabConfig := cfg.FeatureToggleConfig.AlertConfig.GitlabConfig
	gitlabClient, err := gitlab.NewClient(gitlabConfig.BaseURL, gitlabConfig.Token)
	if err != nil {
		log.Panicf("unable to initialize GitLab client: %v", err)
	}

	wardenConfig := cfg.FeatureToggleConfig.AlertConfig.WardenConfig
	wardenClient := warden.NewClient(nil, wardenConfig.APIHost)

	modelEndpointAlertService := service.NewModelEndpointAlertService(
		storage.NewAlertStorage(db), gitlabClient, wardenClient,
		gitlabConfig.DashboardRepository, gitlabConfig.DashboardBranch,
		gitlabConfig.AlertRepository, gitlabConfig.AlertBranch)

	mlflowConfig := cfg.MlflowConfig
	mlflowClient := mlflow.NewClient(mlflowConfig.TrackingURL)
	dispatcher := queue.NewDispatcher(queue.Config{
		NumWorkers: 2,
		Db:         db,
	})
	dispatcher.RegisterJob("test", func(j *queue.Job) error {
		log.Debugf("Job name %s", j.Name)
		return nil
	})
	dispatcher.StartWorkers()
	return api.AppContext{
		EnvironmentService:        environmentService,
		ProjectsService:           projectsService,
		ModelsService:             modelsService,
		ModelEndpointsService:     modelEndpointService,
		VersionsService:           versionsService,
		EndpointsService:          versionEndpointService,
		PredictionJobService:      predictionJobService,
		LogService:                logService,
		SecretService:             secretService,
		ModelEndpointAlertService: modelEndpointAlertService,
		AuthorizationEnabled:      cfg.AuthorizationConfig.AuthorizationEnabled,
		MonitoringConfig:          cfg.FeatureToggleConfig.MonitoringConfig,
		AlertEnabled:              cfg.FeatureToggleConfig.AlertConfig.AlertEnabled,
		DB:                        db,
		Enforcer:                  authEnforcer,
		FeastCoreClient:           coreClient,
		MlflowClient:              mlflowClient,
		Dispatcher:                dispatcher,
	}
}
