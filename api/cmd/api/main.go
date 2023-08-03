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
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mlflowDelete "github.com/caraml-dev/mlp/api/pkg/client/mlflow"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/api"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/database"
	"github.com/caraml-dev/merlin/log"
	mlflow "github.com/caraml-dev/merlin/mlflow"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/gitlab"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/queue/work"
	"github.com/caraml-dev/merlin/service"
	"github.com/caraml-dev/merlin/storage"
	"github.com/caraml-dev/merlin/warden"
	"github.com/caraml-dev/mlp/api/pkg/authz/enforcer"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
)

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

type configFlags []string

func (c *configFlags) String() string {
	return strings.Join(*c, ",")
}

func (c *configFlags) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func main() {
	ctx := context.Background()

	var configFlags configFlags
	flag.Var(&configFlags, "config", "Path to a configuration file. This flag can be specified multiple "+
		"times to load multiple configurations.")
	flag.Parse()

	if len(configFlags) < 1 {
		log.Panicf("Must specify at least one config path using -config")
	}

	var emptyCfg config.Config
	cfg, err := config.Load(&emptyCfg, configFlags...)
	if err != nil {
		log.Panicf("Failed initializing config: %v", err)
	}

	// cfg.ClusterConfig.EnvironmentConfigs = config.InitEnvironmentConfigs(cfg.ClusterConfig.EnvironmentConfigPath)
	environmentConfigs, err := config.InitEnvironmentConfigs(cfg.ClusterConfig.EnvironmentConfigPath)
	if err != nil {
		log.Panicf("Failed initializing environment configs: %v", err)
	}

	cfg.ClusterConfig.EnvironmentConfigs = environmentConfigs
	if cfg.ClusterConfig.InClusterConfig && len(cfg.ClusterConfig.EnvironmentConfigs) > 1 {
		log.Panicf("There should only be one cluster if in cluster credentials are used")
	}

	if err := cfg.Validate(); err != nil {
		log.Panicf("Failed validating config: %s", err)
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

	// Init db
	db, err := database.InitDB(&cfg.DbConfig)
	if err != nil {
		log.Panicf(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Panicf(err.Error())
	}
	defer func() {
		err := sqlDB.Close()
		if err != nil {
			log.Errorf("Error closing connection: %w", err)
		}
	}()

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

	apiRouter, err := api.NewRouter(dependencies.apiContext)
	if err != nil {
		log.Fatalf("Unable to initialize /v1 router: %v", err)
	}
	mount(router, "/v1/internal", healthcheck.NewHandler())
	mount(router, "/v1", apiRouter)
	mount(router, "/metrics", promhttp.Handler())
	mount(router, "/debug", newPprofRouter())

	router.Path("/swagger.yaml").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.SwaggerPath)
	})

	uiEnv := uiEnvHandler{
		ReactAppConfig: &cfg.ReactAppConfig,

		AlertEnabled: cfg.FeatureToggleConfig.AlertConfig.AlertEnabled,

		DefaultFeastServingSource: cfg.StandardTransformerConfig.DefaultFeastSource.String(),
		FeastServingURLs:          cfg.StandardTransformerConfig.FeastServingURLs,

		MonitoringEnabled:              cfg.FeatureToggleConfig.MonitoringConfig.MonitoringEnabled,
		MonitoringPredictionJobBaseURL: cfg.FeatureToggleConfig.MonitoringConfig.MonitoringJobBaseURL,

		ModelDeletionEnabled: cfg.FeatureToggleConfig.ModelDeletionConfig.Enabled,
	}

	uiHomePage := fmt.Sprintf("/%s", strings.TrimPrefix(cfg.ReactAppConfig.HomePage, "/"))

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
			errCh <- fmt.Errorf("server failed: %w", err)
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

	if err := models.InitKubernetesLabeller(cfg.DeploymentLabelPrefix, cfg.Environment); err != nil {
		log.Panicf("invalid deployment label prefix (%s): %s", cfg.DeploymentLabelPrefix, err)
	}

	webServiceBuilder, predJobBuilder, imageBuilderJanitor := initImageBuilder(cfg)

	clusterControllers := initClusterControllers(cfg)
	modelServiceDeployment := initModelServiceDeployment(cfg, webServiceBuilder, clusterControllers, db)
	versionEndpointService := initVersionEndpointService(cfg, webServiceBuilder, clusterControllers, db, coreClient, dispatcher)
	modelEndpointService := initModelEndpointService(cfg, db)

	batchControllers := initBatchControllers(cfg, db, mlpAPIClient)
	batchDeployment := initBatchDeployment(cfg, db, batchControllers, predJobBuilder)
	predictionJobService := initPredictionJobService(cfg, batchControllers, predJobBuilder, db, dispatcher)
	logService := initLogService(cfg)
	// use "mlp" as product name for enforcer so that same policy can be reused by other components
	enforcerCfg := enforcer.NewEnforcerBuilder().KetoEndpoints(cfg.AuthorizationConfig.KetoRemoteRead,
		cfg.AuthorizationConfig.KetoRemoteWrite)
	if cfg.AuthorizationConfig.Caching.Enabled {
		enforcerCfg = enforcerCfg.WithCaching(
			cfg.AuthorizationConfig.Caching.KeyExpirySeconds,
			cfg.AuthorizationConfig.Caching.CacheCleanUpIntervalSeconds,
		)
	}
	authEnforcer, err := enforcerCfg.Build()
	if err != nil {
		log.Panicf("unable to initialize authorization enforcer %v", err)
	}

	projectsService, err := service.NewProjectsService(mlpAPIClient)
	if err != nil {
		log.Panicf("unable to initialize Projects service: %v", err)
	}
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

	mlflowDeleteService, err := mlflowDelete.NewMlflowService(http.DefaultClient, mlflowDelete.Config{
		TrackingURL:         mlflowConfig.TrackingURL,
		ArtifactServiceType: mlflowConfig.ArtifactServiceType,
	})
	if err != nil {
		log.Panicf("failed initializing mlflow delete package: %v", err)
	}

	transformerService := service.NewTransformerService(cfg.StandardTransformerConfig)
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
		TransformerService:        transformerService,
		MlflowDeleteService:       mlflowDeleteService,

		AuthorizationEnabled:      cfg.AuthorizationConfig.AuthorizationEnabled,
		FeatureToggleConfig:       cfg.FeatureToggleConfig,
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
