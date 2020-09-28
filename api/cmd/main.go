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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"golang.org/x/oauth2/google"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/gojek/mlp/pkg/authz/enforcer"
	"github.com/gojek/mlp/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/pkg/instrumentation/sentry"

	"github.com/gojek/merlin/api"
	"github.com/gojek/merlin/batch"
	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/cronjob"
	"github.com/gojek/merlin/gitlab"
	"github.com/gojek/merlin/imagebuilder"
	"github.com/gojek/merlin/istio"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service"
	"github.com/gojek/merlin/storage"
	"github.com/gojek/merlin/vault"
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

	databaseURL := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.DbConfig.Host,
		cfg.DbConfig.Port,
		cfg.DbConfig.User,
		cfg.DbConfig.Database,
		cfg.DbConfig.Password)

	db, err := gorm.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
	defer db.Close()

	// Migrating database
	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(cfg.DbConfig.MigrationPath, "postgres", driver)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}

	// Create an HTTP client for mlp-api client with Google default credential
	googleClient, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		panic(err)
	}

	mlpApiClient := mlp.NewAPIClient(googleClient, cfg.MlpApiConfig.ApiHost, cfg.MlpApiConfig.EncryptionKey)

	vaultClient := initVault(cfg)
	webServiceBuilder, predJobBuilder := initImageBuilder(cfg, vaultClient)

	modelEndpointService := initModelEndpointService(cfg, vaultClient, db)
	versionEndpointService := initVersionEndpointService(cfg, webServiceBuilder, vaultClient, db)
	predictionJobService := initPredictionJobService(cfg, mlpApiClient, predJobBuilder, vaultClient, db)
	logService := initLogService(cfg, vaultClient)

	// use "mlp" as product name for enforcer so that same policy can be reused by excalibur
	authEnforcer, err := enforcer.NewEnforcerBuilder().
		URL(cfg.AuthorizationConfig.AuthorizationServerUrl).
		Product("mlp").
		Build()
	if err != nil {
		log.Panicf("unable to initialize authorization enforcer %v", err)
	}

	projectsService := service.NewProjectsService(mlpApiClient)
	modelsService := service.NewModelsService(db, mlpApiClient)
	versionsService := service.NewVersionsService(db, mlpApiClient)
	environmentService := initEnvironmentService(cfg, db)
	secretService := service.NewSecretService(mlpApiClient)

	gitlabConfig := cfg.FeatureToggleConfig.AlertConfig.GitlabConfig
	gitlabClient, err := gitlab.NewClient(gitlabConfig.BaseURL, gitlabConfig.Token)
	if err != nil {
		log.Panicf("unable to initialize GitLab client: %v", err)
	}

	wardenConfig := cfg.FeatureToggleConfig.AlertConfig.WardenConfig
	wardenClient := warden.NewClient(nil, wardenConfig.ApiHost)

	modelEndpointAlertService := service.NewModelEndpointAlertService(
		storage.NewAlertStorage(db), gitlabClient, wardenClient,
		gitlabConfig.DashboardRepository, gitlabConfig.DashboardBranch,
		gitlabConfig.AlertRepository, gitlabConfig.AlertBranch)

	tracker, err := cronjob.NewTracker(projectsService,
		modelsService,
		storage.NewPredictionJobStorage(db),
		storage.NewDeploymentStorage(db))
	if err != nil {
		log.Panicf("unable to create tracker %v", err)
	}
	tracker.Start()

	appCtx := api.AppContext{
		EnvironmentService: environmentService,

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
	}

	router := mux.NewRouter()

	mount(router, "/v1/internal", healthcheck.NewHandler())
	mount(router, "/v1", api.NewRouter(appCtx))
	mount(router, "/metrics", promhttp.Handler())

	reactConfig := cfg.ReactAppConfig
	uiEnv := uiEnvHandler{
		OauthClientID:     reactConfig.OauthClientID,
		Environment:       reactConfig.Environment,
		SentryDSN:         reactConfig.SentryDSN,
		DocURL:            reactConfig.DocURL,
		AlertEnabled:      reactConfig.AlertEnabled,
		MonitoringEnabled: reactConfig.MonitoringEnabled,
		HomePage:          reactConfig.HomePage,
		MerlinURL:         reactConfig.MerlinURL,
		MlpURL:            reactConfig.MlpURL,
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

type uiEnvHandler struct {
	OauthClientID     string `json:"REACT_APP_OAUTH_CLIENT_ID,omitempty"`
	Environment       string `json:"REACT_APP_ENVIRONMENT,omitempty"`
	SentryDSN         string `json:"REACT_APP_SENTRY_DSN,omitempty"`
	DocURL            string `json:"REACT_APP_MERLIN_DOCS_URL,omitempty"`
	AlertEnabled      bool   `json:"REACT_APP_ALERT_ENABLED"`
	MonitoringEnabled bool   `json:"REACT_APP_MONITORING_DASHBOARD_ENABLED"`
	HomePage          string `json:"REACT_APP_HOMEPAGE,omitempty"`
	MerlinURL         string `json:"REACT_APP_MERLIN_API,omitempty"`
	MlpURL            string `json:"REACT_APP_MLP_API,omitempty"`
}

func (h uiEnvHandler) handler(w http.ResponseWriter, r *http.Request) {
	envJSON, err := json.Marshal(h)
	if err != nil {
		envJSON = []byte("{}")
	}
	fmt.Fprintf(w, "window.env = %s;", envJSON)
}

// uiHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type uiHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h uiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func initPredictionJobService(cfg *config.Config, mlpApiClient mlp.APIClient, builder imagebuilder.ImageBuilder, vaultClient vault.VaultClient, db *gorm.DB) service.PredictionJobService {
	controllers := make(map[string]batch.Controller)
	predictionJobStorage := storage.NewPredictionJobStorage(db)
	for _, env := range cfg.EnvironmentConfigs {
		if !env.IsPredictionJobEnabled {
			continue
		}

		clusterName := env.Cluster
		clusterSecret, err := vaultClient.GetClusterSecret(clusterName)
		if err != nil {
			log.Panicf("unable to get cluster secret of cluster: %s %v", clusterName, err)
		}

		restConfig := &rest.Config{
			Host: clusterSecret.Endpoint,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: false,
				CAData:   []byte(clusterSecret.CaCert),
				CertData: []byte(clusterSecret.ClientCert),
				KeyData:  []byte(clusterSecret.ClientKey),
			},
		}

		sparkClient := versioned.NewForConfigOrDie(restConfig)
		kubeClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			log.Panicf("unable to create kubernetes client: %v", err)
		}

		manifestManager := batch.NewManifestManager(kubeClient)
		envMetadata := cluster.Metadata{
			ClusterName: env.Cluster,
			GcpProject:  env.GcpProject,
		}

		ctl := batch.NewController(predictionJobStorage, mlpApiClient, sparkClient, kubeClient, manifestManager, envMetadata)
		stopCh := make(chan struct{})
		go ctl.Run(stopCh)

		controllers[env.Name] = ctl
	}

	return service.NewPredictionJobService(controllers, builder, predictionJobStorage, clock.RealClock{}, cfg.Environment)
}

func initEnvironmentService(cfg *config.Config, db *gorm.DB) service.EnvironmentService {
	// Synchronize environment in db and config
	// Add new environment if not available
	// Update cluster name and is_default param
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Panicf("rolled back")
		}
	}()

	envSvc, err := service.NewEnvironmentService(tx)
	if err != nil {
		log.Panicf("unable to initialize environment service: %v", err)
	}

	for _, envCfg := range cfg.EnvironmentConfigs {
		var isDefault *bool = nil
		if envCfg.IsDefault {
			isDefault = &envCfg.IsDefault
		}

		var isDefaultPredictionJob *bool = nil
		if envCfg.IsDefaultPredictionJob {
			isDefaultPredictionJob = &envCfg.IsDefaultPredictionJob
		}

		cfg := config.ParseDeploymentConfig(envCfg)

		env, err := envSvc.GetEnvironment(envCfg.Name)
		if err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Panicf("unable to get environment %s: %v", envCfg.Name, err)
			}

			// Create new environment
			log.Infof("adding environment %s: cluster: %s, is_default: %v", envCfg.Name, envCfg.Cluster, envCfg.IsDefault)
			env = &models.Environment{
				Name:       envCfg.Name,
				Cluster:    envCfg.Cluster,
				IsDefault:  isDefault,
				Region:     envCfg.Region,
				GcpProject: envCfg.GcpProject,
				MaxCpu:     envCfg.MaxCpu,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    cfg.MinReplica,
					MaxReplica:    cfg.MaxReplica,
					CpuRequest:    cfg.CpuRequest,
					MemoryRequest: cfg.MemoryRequest,
				},
				IsDefaultPredictionJob: isDefaultPredictionJob,
				IsPredictionJobEnabled: envCfg.IsPredictionJobEnabled,
			}

			if envCfg.IsPredictionJobEnabled {
				env.DefaultPredictionJobResourceRequest = &models.PredictionJobResourceRequest{
					DriverCpuRequest:      envCfg.PredictionJobConfig.DriverCpuRequest,
					DriverMemoryRequest:   envCfg.PredictionJobConfig.DriverMemoryRequest,
					ExecutorReplica:       envCfg.PredictionJobConfig.ExecutorReplica,
					ExecutorCpuRequest:    envCfg.PredictionJobConfig.ExecutorCpuRequest,
					ExecutorMemoryRequest: envCfg.PredictionJobConfig.ExecutorMemoryRequest,
				}
			}

		} else {
			// Update
			log.Infof("updating environment %s: cluster: %s, is_default: %v", envCfg.Name, envCfg.Cluster, envCfg.IsDefault)

			env.Cluster = envCfg.Cluster
			env.IsDefault = isDefault
			env.Region = envCfg.Region
			env.GcpProject = envCfg.GcpProject
			env.MaxCpu = envCfg.MaxCpu
			env.MaxMemory = envCfg.MaxMemory
			env.DefaultResourceRequest = &models.ResourceRequest{
				MinReplica:    cfg.MinReplica,
				MaxReplica:    cfg.MaxReplica,
				CpuRequest:    cfg.CpuRequest,
				MemoryRequest: cfg.MemoryRequest,
			}
			env.IsDefaultPredictionJob = isDefaultPredictionJob
			env.IsPredictionJobEnabled = envCfg.IsPredictionJobEnabled

			if envCfg.IsPredictionJobEnabled {
				env.DefaultPredictionJobResourceRequest = &models.PredictionJobResourceRequest{
					DriverCpuRequest:      envCfg.PredictionJobConfig.DriverCpuRequest,
					DriverMemoryRequest:   envCfg.PredictionJobConfig.DriverMemoryRequest,
					ExecutorReplica:       envCfg.PredictionJobConfig.ExecutorReplica,
					ExecutorCpuRequest:    envCfg.PredictionJobConfig.ExecutorCpuRequest,
					ExecutorMemoryRequest: envCfg.PredictionJobConfig.ExecutorMemoryRequest,
				}
			}
		}

		_, err = envSvc.Save(env)
		if err != nil {
			log.Panicf("unable to update environment %s: %v", env.Name, err)
		}
	}

	// Ensure at least 1 environment is set as default
	defaultEnv, err := envSvc.GetDefaultEnvironment()
	if err != nil {
		log.Panicf("No default environment is set")
	}

	// Ensure at least 1 environment is set as default for prediction job
	predJobDefaultEnv, err := envSvc.GetDefaultPredictionJobEnvironment()
	if err != nil {
		log.Panicf("No default environment for prediction job is set")
	}

	log.Infof("%s is set as default environment for webservice", defaultEnv.Name)
	log.Infof("%s is set as default environment for prediction job", predJobDefaultEnv.Name)

	// environment_name column in model_endpoints and version_endpoints are null
	// the first time db migration is performed from 0.1.0 to 0.2.0
	// These code is to populate the null environment_name with default environment
	err = tx.Model(&models.VersionEndpoint{}).
		Where("environment_name is null").
		Select("environment_name").
		Updates(map[string]interface{}{"environment_name": defaultEnv.Name}).
		Error
	if err != nil {
		log.Panicf("unable to update missing environment name %v", err)
	}

	err = tx.Model(&models.ModelEndpoint{}).
		Where("environment_name is null").
		Select("environment_name").
		Updates(map[string]interface{}{"environment_name": defaultEnv.Name}).
		Error
	if err != nil {
		log.Panicf("unable to update missing environment name %v", err)
	}

	err = tx.Commit().Error
	if err != nil {
		log.Panicf("Unable to commit update to environment table %v", err)
	}
	svc, _ := service.NewEnvironmentService(db)
	return svc
}

func initModelEndpointService(cfg *config.Config, vaultClient vault.VaultClient, db *gorm.DB) service.ModelEndpointsService {
	istioClients := make(map[string]istio.Client)
	for _, env := range cfg.EnvironmentConfigs {
		clusterName := env.Cluster
		clusterSecret, err := vaultClient.GetClusterSecret(clusterName)
		if err != nil {
			log.Panicf("unable to get cluster secret of cluster: %s %v", clusterName, err)
		}

		istioClient, err := istio.NewClient(istio.Config{
			ClusterHost:       clusterSecret.Endpoint,
			ClusterCACert:     clusterSecret.CaCert,
			ClusterClientCert: clusterSecret.ClientCert,
			ClusterClientKey:  clusterSecret.ClientKey,
		})
		if err != nil {
			log.Panicf("unable to initialize cluster controller %v", err)
		}

		istioClients[env.Name] = istioClient
	}

	return service.NewModelEndpointsService(istioClients, db, cfg.Environment)
}

func initVersionEndpointService(cfg *config.Config, builder imagebuilder.ImageBuilder, vaultClient vault.VaultClient, db *gorm.DB) service.EndpointsService {
	controllers := make(map[string]cluster.Controller)
	for _, env := range cfg.EnvironmentConfigs {
		clusterName := env.Cluster
		clusterSecret, err := vaultClient.GetClusterSecret(clusterName)
		if err != nil {
			log.Panicf("unable to get cluster secret of cluster: %s %v", clusterName, err)
		}

		ctl, err := cluster.NewController(cluster.ClusterConfig{
			Host:       clusterSecret.Endpoint,
			CACert:     clusterSecret.CaCert,
			ClientCert: clusterSecret.ClientCert,
			ClientKey:  clusterSecret.ClientKey,

			ClusterName: clusterName,
			GcpProject:  env.GcpProject,
		}, config.ParseDeploymentConfig(env))
		if err != nil {
			log.Panicf("unable to initialize cluster controller %v", err)
		}

		controllers[env.Name] = ctl
	}

	return service.NewEndpointService(controllers, builder, storage.NewVersionEndpointStorage(db),
		storage.NewDeploymentStorage(db), cfg.Environment,
		cfg.FeatureToggleConfig.MonitoringConfig)
}

func initVault(cfg *config.Config) vault.VaultClient {
	vaultConfig := &vault.Config{
		Address: cfg.VaultConfig.Address,
		Token:   cfg.VaultConfig.Token,
	}
	vaultClient, err := vault.NewVaultClient(vaultConfig)
	if err != nil {
		log.Panicf("unable to initialize vault")
	}
	return vaultClient
}

func initImageBuilder(cfg *config.Config, vaultClient vault.VaultClient) (webserviceBuilder imagebuilder.ImageBuilder, predJobBuilder imagebuilder.ImageBuilder) {
	imgBuilderClusterSecret, err := vaultClient.GetClusterSecret(cfg.ImageBuilderConfig.ClusterName)
	if err != nil {
		log.Panicf("unable to retrieve secret for cluster %s from vault", cfg.ImageBuilderConfig.ClusterName)
	}

	restConfig := &rest.Config{
		Host: imgBuilderClusterSecret.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   []byte(imgBuilderClusterSecret.CaCert),
			CertData: []byte(imgBuilderClusterSecret.ClientCert),
			KeyData:  []byte(imgBuilderClusterSecret.ClientKey),
		},
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Panicf("unable to initialize image builder")
	}

	timeout, err := time.ParseDuration(cfg.ImageBuilderConfig.BuildTimeout)
	if err != nil {
		log.Panicf("unable to parse image builder timeout to time.Duration %s", cfg.ImageBuilderConfig.BuildTimeout)
	}

	webServiceConfig := imagebuilder.Config{
		BuildContextUrl:      cfg.ImageBuilderConfig.BuildContextUri,
		DockerfilePath:       cfg.ImageBuilderConfig.DockerfilePath,
		BaseImage:            cfg.ImageBuilderConfig.BaseImage,
		BuildNamespace:       cfg.ImageBuilderConfig.BuildNamespace,
		DockerRegistry:       cfg.ImageBuilderConfig.DockerRegistry,
		ContextSubPath:       cfg.ImageBuilderConfig.ContextSubPath,
		BuildTimeoutDuration: timeout,

		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,

		Environment: cfg.Environment,
	}
	webserviceBuilder = imagebuilder.NewModelServiceImageBuilder(kubeClient, webServiceConfig)
	predJobConfig := imagebuilder.Config{
		BuildContextUrl:      cfg.ImageBuilderConfig.PredictionJobBuildContextUri,
		DockerfilePath:       cfg.ImageBuilderConfig.PredictionJobDockerfilePath,
		BaseImage:            cfg.ImageBuilderConfig.PredictionJobBaseImage,
		BuildNamespace:       cfg.ImageBuilderConfig.BuildNamespace,
		DockerRegistry:       cfg.ImageBuilderConfig.DockerRegistry,
		ContextSubPath:       cfg.ImageBuilderConfig.PredictionJobContextSubPath,
		BuildTimeoutDuration: timeout,

		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,

		Environment: cfg.Environment,
	}
	predJobBuilder = imagebuilder.NewPredictionJobImageBuilder(kubeClient, predJobConfig)
	return
}

func initLogService(cfg *config.Config, vaultClient vault.VaultClient) service.LogService {
	// image builder cluster
	imgBuilderClusterSecret, err := vaultClient.GetClusterSecret(cfg.ImageBuilderConfig.ClusterName)
	if err != nil {
		log.Panicf("unable to retrieve secret for cluster %s from vault: %v", cfg.ImageBuilderConfig.ClusterName, err)
	}
	restConfig := &rest.Config{
		Host: imgBuilderClusterSecret.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   []byte(imgBuilderClusterSecret.CaCert),
			CertData: []byte(imgBuilderClusterSecret.ClientCert),
			KeyData:  []byte(imgBuilderClusterSecret.ClientKey),
		},
	}

	c, err := corev1.NewForConfig(restConfig)

	var clusterClients = make(map[string]corev1.CoreV1Interface)
	clusterClients[cfg.ImageBuilderConfig.ClusterName] = c

	for _, env := range cfg.EnvironmentConfigs {
		clusterName := env.Cluster
		clusterSecret, err := vaultClient.GetClusterSecret(clusterName)
		if err != nil {
			log.Panicf("unable to get cluster secret of cluster %s %v", clusterName, err)
		}

		restConfig := &rest.Config{
			Host: clusterSecret.Endpoint,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: false,
				CAData:   []byte(clusterSecret.CaCert),
				CertData: []byte(clusterSecret.ClientCert),
				KeyData:  []byte(clusterSecret.ClientKey),
			},
		}

		c, err = corev1.NewForConfig(restConfig)
		clusterClients[clusterName] = c
	}

	return service.NewLogService(clusterClients)
}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}
