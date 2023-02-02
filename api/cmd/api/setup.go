package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"

	"github.com/gojek/merlin/api"
	"github.com/gojek/merlin/batch"
	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/istio"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/imagebuilder"
	"github.com/gojek/merlin/queue"
	"github.com/gojek/merlin/queue/work"
	"github.com/gojek/merlin/service"
	"github.com/gojek/merlin/storage"
	mlpcluster "github.com/gojek/mlp/api/pkg/cluster"
)

type deps struct {
	apiContext          api.AppContext
	modelDeployment     *work.ModelServiceDeployment
	batchDeployment     *work.BatchDeployment
	imageBuilderJanitor *imagebuilder.Janitor
}

func initDB(cfg config.DatabaseConfig) (*gorm.DB, func()) {
	databaseURL := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Database,
		cfg.Password)

	db, err := gorm.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}
	db.LogMode(false)

	sqlDB := db.DB()
	if sqlDB == nil {
		panic("fail to get the underlying database connection")
	}

	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, func() { db.Close() } //nolint:errcheck
}

func runDBMigration(db *gorm.DB, migrationPath string) {
	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
}

func initMLPAPIClient(ctx context.Context, cfg config.MlpAPIConfig) mlp.APIClient {
	mlpHTTPClient := http.DefaultClient
	googleClient, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/userinfo.email")
	if err == nil {
		mlpHTTPClient = googleClient
	} else {
		log.Infof("Google default credential not found. Fallback to default HTTP client.")
	}

	return mlp.NewAPIClient(mlpHTTPClient, cfg.APIHost, cfg.EncryptionKey)
}

func initFeastCoreClient(feastCoreURL, feastAuthAudience string, enableAuth bool) core.CoreServiceClient {
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if enableAuth {
		cred, err := feast.NewGoogleCredential(feastAuthAudience)
		if err != nil {
			panic(err)
		}
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(cred))
	}

	cc, err := grpc.Dial(feastCoreURL, dialOpts...)
	if err != nil {
		panic(err)
	}
	return core.NewCoreServiceClient(cc)
}

func initImageBuilder(cfg *config.Config) (webserviceBuilder imagebuilder.ImageBuilder, predJobBuilder imagebuilder.ImageBuilder, imageBuilderJanitor *imagebuilder.Janitor) {
	imgBuilderK8sConfig := cfg.ImageBuilderConfig.K8sConfig
	creds := mlpcluster.NewK8sClusterCreds(&imgBuilderK8sConfig)

	restConfig, err := creds.ToRestConfig()
	if err != nil {
		log.Panicf("%s, unable to get image builder k8s config", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Panicf("%s unable to initialize image builder", err.Error())
	}

	timeout, err := time.ParseDuration(cfg.ImageBuilderConfig.BuildTimeout)
	if err != nil {
		log.Panicf("unable to parse image builder timeout to time.Duration %s", cfg.ImageBuilderConfig.BuildTimeout)
	}

	webServiceConfig := imagebuilder.Config{
		BaseImages:           cfg.ImageBuilderConfig.BaseImages,
		BuildNamespace:       cfg.ImageBuilderConfig.BuildNamespace,
		DockerRegistry:       cfg.ImageBuilderConfig.DockerRegistry,
		ContextSubPath:       cfg.ImageBuilderConfig.ContextSubPath,
		BuildTimeoutDuration: timeout,
		KanikoImage:          cfg.ImageBuilderConfig.KanikoImage,
		Tolerations:          cfg.ImageBuilderConfig.Tolerations,
		NodeSelectors:        cfg.ImageBuilderConfig.NodeSelectors,
		MaximumRetry:         cfg.ImageBuilderConfig.MaximumRetry,

		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,

		Environment: cfg.Environment,
	}
	webserviceBuilder = imagebuilder.NewModelServiceImageBuilder(kubeClient, webServiceConfig)

	predJobConfig := imagebuilder.Config{
		BaseImages:           cfg.ImageBuilderConfig.PredictionJobBaseImages,
		BuildNamespace:       cfg.ImageBuilderConfig.BuildNamespace,
		DockerRegistry:       cfg.ImageBuilderConfig.DockerRegistry,
		ContextSubPath:       cfg.ImageBuilderConfig.PredictionJobContextSubPath,
		BuildTimeoutDuration: timeout,
		KanikoImage:          cfg.ImageBuilderConfig.KanikoImage,
		Tolerations:          cfg.ImageBuilderConfig.Tolerations,
		NodeSelectors:        cfg.ImageBuilderConfig.NodeSelectors,
		MaximumRetry:         cfg.ImageBuilderConfig.MaximumRetry,

		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,

		Environment: cfg.Environment,
	}
	predJobBuilder = imagebuilder.NewPredictionJobImageBuilder(kubeClient, predJobConfig)

	ctl, err := cluster.NewController(cluster.Config{
		Credentials: creds,
		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,
	},
		config.DeploymentConfig{}, // We don't need deployment config here because we're going to retrieve the log not deploy model.
		config.StandardTransformerConfig{})
	imageBuilderJanitor = imagebuilder.NewJanitor(ctl, imagebuilder.JanitorConfig{
		BuildNamespace: cfg.ImageBuilderConfig.BuildNamespace,
		Retention:      cfg.ImageBuilderConfig.Retention,
	})

	if err != nil {
		log.Panicf("unable to initialize cluster controller")
	}

	return
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
				MaxCPU:     envCfg.MaxCPU,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    cfg.DefaultModelResourceRequests.MinReplica,
					MaxReplica:    cfg.DefaultModelResourceRequests.MaxReplica,
					CPURequest:    cfg.DefaultModelResourceRequests.CPURequest,
					MemoryRequest: cfg.DefaultModelResourceRequests.MemoryRequest,
				},
				DefaultTransformerResourceRequest: &models.ResourceRequest{
					MinReplica:    cfg.DefaultTransformerResourceRequests.MinReplica,
					MaxReplica:    cfg.DefaultTransformerResourceRequests.MaxReplica,
					CPURequest:    cfg.DefaultTransformerResourceRequests.CPURequest,
					MemoryRequest: cfg.DefaultTransformerResourceRequests.MemoryRequest,
				},
				IsDefaultPredictionJob: isDefaultPredictionJob,
				IsPredictionJobEnabled: envCfg.IsPredictionJobEnabled,
			}

			if envCfg.IsPredictionJobEnabled {
				env.DefaultPredictionJobResourceRequest = &models.PredictionJobResourceRequest{
					DriverCPURequest:      envCfg.DefaultPredictionJobConfig.DriverCPURequest,
					DriverMemoryRequest:   envCfg.DefaultPredictionJobConfig.DriverMemoryRequest,
					ExecutorReplica:       envCfg.DefaultPredictionJobConfig.ExecutorReplica,
					ExecutorCPURequest:    envCfg.DefaultPredictionJobConfig.ExecutorCPURequest,
					ExecutorMemoryRequest: envCfg.DefaultPredictionJobConfig.ExecutorMemoryRequest,
				}
			}

		} else {
			// Update
			log.Infof("updating environment %s: cluster: %s, is_default: %v", envCfg.Name, envCfg.Cluster, envCfg.IsDefault)

			env.Cluster = envCfg.Cluster
			env.IsDefault = isDefault
			env.Region = envCfg.Region
			env.GcpProject = envCfg.GcpProject
			env.MaxCPU = envCfg.MaxCPU
			env.MaxMemory = envCfg.MaxMemory
			env.DefaultResourceRequest = &models.ResourceRequest{
				MinReplica:    cfg.DefaultModelResourceRequests.MinReplica,
				MaxReplica:    cfg.DefaultModelResourceRequests.MaxReplica,
				CPURequest:    cfg.DefaultModelResourceRequests.CPURequest,
				MemoryRequest: cfg.DefaultModelResourceRequests.MemoryRequest,
			}
			env.DefaultTransformerResourceRequest = &models.ResourceRequest{
				MinReplica:    cfg.DefaultTransformerResourceRequests.MinReplica,
				MaxReplica:    cfg.DefaultTransformerResourceRequests.MaxReplica,
				CPURequest:    cfg.DefaultTransformerResourceRequests.CPURequest,
				MemoryRequest: cfg.DefaultTransformerResourceRequests.MemoryRequest,
			}
			env.IsDefaultPredictionJob = isDefaultPredictionJob
			env.IsPredictionJobEnabled = envCfg.IsPredictionJobEnabled

			if envCfg.IsPredictionJobEnabled {
				env.DefaultPredictionJobResourceRequest = &models.PredictionJobResourceRequest{
					DriverCPURequest:      envCfg.DefaultPredictionJobConfig.DriverCPURequest,
					DriverMemoryRequest:   envCfg.DefaultPredictionJobConfig.DriverMemoryRequest,
					ExecutorReplica:       envCfg.DefaultPredictionJobConfig.ExecutorReplica,
					ExecutorCPURequest:    envCfg.DefaultPredictionJobConfig.ExecutorCPURequest,
					ExecutorMemoryRequest: envCfg.DefaultPredictionJobConfig.ExecutorMemoryRequest,
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

func initModelEndpointService(cfg *config.Config, db *gorm.DB) service.ModelEndpointsService {
	istioClients := make(map[string]istio.Client)
	for _, env := range cfg.EnvironmentConfigs {
		creds := mlpcluster.NewK8sClusterCreds(env.K8sConfig)

		istioClient, err := istio.NewClient(istio.Config{
			ClusterHost: env.K8sConfig.Cluster.Server,
			Credentials: creds,
		})
		if err != nil {
			log.Panicf("unable to initialize cluster controller %v", err)
		}

		istioClients[env.Name] = istioClient
	}

	return service.NewModelEndpointsService(istioClients, storage.NewModelEndpointStorage(db), storage.NewVersionEndpointStorage(db), cfg.Environment)
}

func initBatchDeployment(cfg *config.Config, db *gorm.DB, controllers map[string]batch.Controller, builder imagebuilder.ImageBuilder) *work.BatchDeployment {
	return &work.BatchDeployment{
		Store:            storage.NewPredictionJobStorage(db),
		ImageBuilder:     builder,
		BatchControllers: controllers,
		Clock:            clock.RealClock{},
		EnvironmentLabel: cfg.Environment,
	}
}

func initBatchControllers(cfg *config.Config, db *gorm.DB, mlpAPIClient mlp.APIClient) map[string]batch.Controller {
	controllers := make(map[string]batch.Controller)
	predictionJobStorage := storage.NewPredictionJobStorage(db)
	for _, env := range cfg.EnvironmentConfigs {
		if !env.IsPredictionJobEnabled {
			continue
		}

		creds := mlpcluster.NewK8sClusterCreds(env.K8sConfig)
		clusterName := env.Cluster
		restConfig, err := creds.ToRestConfig()
		if err != nil {
			log.Panicf("unable to get cluster config of cluster: %s %v", clusterName, err)
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

		ctl := batch.NewController(predictionJobStorage, mlpAPIClient, sparkClient, kubeClient, manifestManager, envMetadata)
		stopCh := make(chan struct{})
		go ctl.Run(stopCh)

		controllers[env.Name] = ctl
	}
	return controllers
}

func initPredictionJobService(cfg *config.Config, controllers map[string]batch.Controller, builder imagebuilder.ImageBuilder, db *gorm.DB, producer queue.Producer) service.PredictionJobService {
	predictionJobStorage := storage.NewPredictionJobStorage(db)
	return service.NewPredictionJobService(controllers, builder, predictionJobStorage, clock.RealClock{}, cfg.Environment, producer)
}

func initModelServiceDeployment(cfg *config.Config, builder imagebuilder.ImageBuilder, controllers map[string]cluster.Controller, db *gorm.DB) *work.ModelServiceDeployment {
	return &work.ModelServiceDeployment{
		ClusterControllers:   controllers,
		ImageBuilder:         builder,
		Storage:              storage.NewVersionEndpointStorage(db),
		DeploymentStorage:    storage.NewDeploymentStorage(db),
		LoggerDestinationURL: cfg.LoggerDestinationURL,
	}
}

func initClusterControllers(cfg *config.Config) map[string]cluster.Controller {
	controllers := make(map[string]cluster.Controller)
	for _, env := range cfg.EnvironmentConfigs {
		clusterName := env.Cluster
		creds := mlpcluster.NewK8sClusterCreds(env.K8sConfig)

		ctl, err := cluster.NewController(cluster.Config{
			Credentials: creds,

			ClusterName: clusterName,
			GcpProject:  env.GcpProject,
		},
			config.ParseDeploymentConfig(env),
			cfg.StandardTransformerConfig)
		if err != nil {
			log.Panicf("unable to initialize cluster controller %v", err)
		}

		controllers[env.Name] = ctl
	}
	return controllers
}

func initVersionEndpointService(cfg *config.Config, builder imagebuilder.ImageBuilder, controllers map[string]cluster.Controller, db *gorm.DB, feastCoreClient core.CoreServiceClient, producer queue.Producer) service.EndpointsService {
	return service.NewEndpointService(service.EndpointServiceParams{
		ClusterControllers:        controllers,
		ImageBuilder:              builder,
		Storage:                   storage.NewVersionEndpointStorage(db),
		DeploymentStorage:         storage.NewDeploymentStorage(db),
		MonitoringConfig:          cfg.FeatureToggleConfig.MonitoringConfig,
		LoggerDestinationURL:      cfg.LoggerDestinationURL,
		JobProducer:               producer,
		FeastCoreClient:           feastCoreClient,
		StandardTransformerConfig: cfg.StandardTransformerConfig,
	})
}

func initLogService(cfg *config.Config) service.LogService {
	creds := mlpcluster.NewK8sClusterCreds(&cfg.ImageBuilderConfig.K8sConfig)
	ctl, err := cluster.NewController(cluster.Config{
		Credentials: creds,

		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,
	},
		config.DeploymentConfig{}, // We don't need deployment config here because we're going to retrieve the log not deploy model.
		cfg.StandardTransformerConfig)
	if err != nil {
		log.Panicf("unable to initialize cluster controller %v", err)
	}

	clusterControllers := make(map[string]cluster.Controller)
	clusterControllers[cfg.ImageBuilderConfig.ClusterName] = ctl

	for _, env := range cfg.EnvironmentConfigs {
		clusterName := env.Cluster
		creds := mlpcluster.NewK8sClusterCreds(env.K8sConfig)

		ctl, err := cluster.NewController(cluster.Config{
			Credentials: creds,
			ClusterName: clusterName,
			GcpProject:  env.GcpProject,
		},
			config.ParseDeploymentConfig(env),
			cfg.StandardTransformerConfig)
		if err != nil {
			log.Panicf("unable to initialize cluster controller %v", err)
		}

		clusterControllers[clusterName] = ctl
	}

	return service.NewLogService(clusterControllers)
}
