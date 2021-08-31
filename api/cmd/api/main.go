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
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	"github.com/gojek/merlin/gitlab"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlflow"
	"github.com/gojek/merlin/queue"
	"github.com/gojek/merlin/queue/work"
	"github.com/gojek/merlin/service"
	"github.com/gojek/merlin/storage"
	"github.com/gojek/merlin/warden"
)

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

func main() {
	ctx := context.Background()

	cfg, err := config.InitConfigEnv()
	if err != nil {
		log.Panicf("Failed initializing config: %v", err)
	}

	fmt.Printf("imageBuilding job config %+v", cfg.ImageBuilderConfig)

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

	dispatcher := queue.NewDispatcher(queue.Config{
		NumWorkers: cfg.NumOfQueueWorkers,
		Db:         db,
	})
	defer dispatcher.Stop()

	dependencies := buildDependencies(ctx, cfg, db, dispatcher)

	registerQueueJob(dispatcher, dependencies.modelDeployment, dependencies.batchDeployment)
	dispatcher.Start()

	if err := initCronJob(dependencies, db); err != nil {
		log.Panicf("Failed to initialize cron jobs: %s", err)
	}

	router := mux.NewRouter()

	mount(router, "/v1/internal", healthcheck.NewHandler())
	mount(router, "/v1", api.NewRouter(dependencies.apiContext))
	mount(router, "/metrics", promhttp.Handler())
	mount(router, "/debug", newPprofRouter())

	reactConfig := cfg.ReactAppConfig
	uiEnv := uiEnvHandler{
		OauthClientID:     reactConfig.OauthClientID,
		Environment:       reactConfig.Environment,
		SentryDSN:         reactConfig.SentryDSN,
		DocURL:            reactConfig.DocURL,
		HomePage:          reactConfig.HomePage,
		MerlinURL:         reactConfig.MerlinURL,
		MlpURL:            reactConfig.MlpURL,
		DockerRegistries:  reactConfig.DockerRegistries,
		MaxAllowedReplica: reactConfig.MaxAllowedReplica,

		DefaultFeastServingURL: cfg.StandardTransformerConfig.DefaultFeastServingURL,
		FeastServingURLs:       cfg.StandardTransformerConfig.FeastServingURLs,
		FeastCoreURL:           reactConfig.FeastCoreURL,

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
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: cors.AllowAll().Handler(router),
	}

	stopCh := setupSignalHandler()
	errCh := make(chan error, 1)
	go func() {
		// Don't forward ErrServerClosed as that indicates we're already shutting down.
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server failed: %+v", err)
		}
	}()

	// Exit as soon as we see a shutdown signal or the server failed.
	select {
	case <-stopCh:
	case err := <-errCh:
		log.Errorf("error %+v", err)
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Errorf("failed to shutdown HTTP server")
	}
}

func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}

func newPprofRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Path("/pprof/cmdline").HandlerFunc(http.HandlerFunc(pprof.Cmdline))
	r.Path("/pprof/profile").HandlerFunc(http.HandlerFunc(pprof.Profile))
	r.Path("/pprof/symbol").HandlerFunc(http.HandlerFunc(pprof.Symbol))
	r.Path("/pprof/trace").HandlerFunc(http.HandlerFunc(pprof.Trace))
	r.Path("/pprof/allocs").Handler(pprof.Handler("allocs"))
	r.Path("/pprof/block").Handler(pprof.Handler("block"))
	r.Path("/pprof/goroutine").Handler(pprof.Handler("goroutine"))
	r.Path("/pprof/heap").Handler(pprof.Handler("heap"))
	r.Path("/pprof/mutex").Handler(pprof.Handler("mutex"))
	r.Path("/pprof/threadcreate").Handler(pprof.Handler("threadcreate"))
	r.Path("/pprof/").HandlerFunc(http.HandlerFunc(pprof.Index))
	return r
}

func registerQueueJob(consumer queue.Consumer, modelServiceDepl *work.ModelServiceDeployment, batchDepl *work.BatchDeployment) {
	consumer.RegisterJob(service.ModelServiceDeployment, modelServiceDepl.Deploy)
	consumer.RegisterJob(service.BatchDeployment, batchDepl.Deploy)
}

func buildDependencies(ctx context.Context, cfg *config.Config, db *gorm.DB, dispatcher *queue.Dispatcher) deps {
	mlpAPIClient := initMLPAPIClient(ctx, cfg.MlpAPIConfig)
	coreClient := initFeastCoreClient(cfg.StandardTransformerConfig.FeastCoreURL, cfg.StandardTransformerConfig.FeastCoreAuthAudience, cfg.StandardTransformerConfig.EnableAuth)

	vaultClient := initVault(cfg.VaultConfig)
	webServiceBuilder, predJobBuilder, imageBuilderJanitor := initImageBuilder(cfg, vaultClient)

	modelEndpointService := initModelEndpointService(cfg, vaultClient, db)

	clusterControllers := initClusterControllers(cfg, vaultClient)
	modelServiceDeployment := initModelServiceDeployment(cfg, webServiceBuilder, clusterControllers, db)
	versionEndpointService := initVersionEndpointService(cfg, webServiceBuilder, clusterControllers, db, dispatcher)

	batchControllers := initBatchControllers(cfg, vaultClient, db, mlpAPIClient)
	batchDeployment := initBatchDeployment(cfg, db, batchControllers, predJobBuilder)
	predictionJobService := initPredictionJobService(cfg, batchControllers, predJobBuilder, db, dispatcher)
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
		gitlabConfig.AlertRepository, gitlabConfig.AlertBranch,
		cfg.FeatureToggleConfig.MonitoringConfig.MonitoringBaseURL)

	mlflowConfig := cfg.MlflowConfig
	mlflowClient := mlflow.NewClient(mlflowConfig.TrackingURL)
	apiContext := api.AppContext{
		DB:       db,
		Enforcer: authEnforcer,

		EnvironmentService:        environmentService,
		ProjectsService:           projectsService,
		ModelsService:             modelsService,
		ModelEndpointsService:     modelEndpointService,
		VersionsService:           versionsService,
		EndpointsService:          versionEndpointService,
		LogService:                logService,
		PredictionJobService:      predictionJobService,
		SecretService:             secretService,
		ModelEndpointAlertService: modelEndpointAlertService,

		AuthorizationEnabled: cfg.AuthorizationConfig.AuthorizationEnabled,
		AlertEnabled:         cfg.FeatureToggleConfig.AlertConfig.AlertEnabled,
		MonitoringConfig:     cfg.FeatureToggleConfig.MonitoringConfig,

		StandardTransformerConfig: cfg.StandardTransformerConfig,

		FeastCoreClient: coreClient,
		MlflowClient:    mlflowClient,
	}
	return deps{
		apiContext:          apiContext,
		modelDeployment:     modelServiceDeployment,
		batchDeployment:     batchDeployment,
		imageBuilderJanitor: imageBuilderJanitor,
	}
}
