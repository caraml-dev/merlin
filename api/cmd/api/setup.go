package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	"github.com/caraml-dev/mlp/api/pkg/artifact"
	"github.com/caraml-dev/mlp/api/pkg/auth"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/caraml-dev/merlin/api"
	"github.com/caraml-dev/merlin/batch"
	"github.com/caraml-dev/merlin/cluster"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/istio"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/imagebuilder"
	"github.com/caraml-dev/merlin/pkg/observability/deployment"
	"github.com/caraml-dev/merlin/pkg/observability/event"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/queue/work"
	"github.com/caraml-dev/merlin/service"
	"github.com/caraml-dev/merlin/storage"
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
)

type deps struct {
	apiContext              api.AppContext
	modelDeployment         *work.ModelServiceDeployment
	batchDeployment         *work.BatchDeployment
	observabilityDeployment *work.ObservabilityPublisherDeployment
	imageBuilderJanitor     *imagebuilder.Janitor
}

func initMLPAPIClient(ctx context.Context, cfg config.MlpAPIConfig) mlp.APIClient {
	mlpHTTPClient := http.DefaultClient
	googleClient, err := auth.InitGoogleClient(context.Background())
	if err == nil {
		mlpHTTPClient = googleClient
	} else {
		log.Infof("Google default credential not found. Fallback to default HTTP client.")
	}

	return mlp.NewAPIClient(mlpHTTPClient, cfg.APIHost)
}

func initFeastCoreClient(feastCoreURL, feastAuthAudience string, enableAuth bool) core.CoreServiceClient {
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if enableAuth {
		cred, err := feast.NewGoogleCredential(feastAuthAudience)
		if err != nil {
			log.Panicf(err.Error())
		}
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(cred))
	}

	cc, err := grpc.Dial(feastCoreURL, dialOpts...)
	if err != nil {
		log.Panicf(err.Error())
	}
	return core.NewCoreServiceClient(cc)
}

func initImageBuilder(cfg *config.Config) (webserviceBuilder imagebuilder.ImageBuilder, predJobBuilder imagebuilder.ImageBuilder, imageBuilderJanitor *imagebuilder.Janitor) {
	clusterCfg := cluster.Config{
		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,
	}

	var restConfig *rest.Config
	var err error
	if cfg.ClusterConfig.InClusterConfig {
		clusterCfg.InClusterConfig = true
		restConfig, err = rest.InClusterConfig()
	} else {
		clusterCfg.Credentials = mlpcluster.NewK8sClusterCreds(cfg.ImageBuilderConfig.K8sConfig)
		restConfig, err = clusterCfg.Credentials.ToRestConfig()
	}
	if err != nil {
		log.Panicf("%s, unable to get image builder k8s config", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Panicf("%s, unable to initialize image builder", err.Error())
	}

	timeout, err := time.ParseDuration(cfg.ImageBuilderConfig.BuildTimeout)
	if err != nil {
		log.Panicf("unable to parse image builder timeout to time.Duration %s", cfg.ImageBuilderConfig.BuildTimeout)
	}

	var artifactService artifact.Service
	if cfg.ImageBuilderConfig.ArtifactServiceType == "gcs" {
		gcsClient, err := gcs.NewClient(context.Background())
		if err != nil {
			log.Panicf("%s,failed initializing gcs for mlflow delete package", err.Error())
		}
		artifactService = artifact.NewGcsArtifactClient(gcsClient)
	} else if cfg.ImageBuilderConfig.ArtifactServiceType == "nop" {
		artifactService = artifact.NewNopArtifactClient()
	} else {
		log.Panicf("invalid artifact service type %s", cfg.ImageBuilderConfig.ArtifactServiceType)
	}

	webServiceConfig := imagebuilder.Config{
		BaseImage:               cfg.ImageBuilderConfig.BaseImage,
		BuildNamespace:          cfg.ImageBuilderConfig.BuildNamespace,
		DockerRegistry:          cfg.ImageBuilderConfig.DockerRegistry,
		BuildTimeoutDuration:    timeout,
		KanikoImage:             cfg.ImageBuilderConfig.KanikoImage,
		KanikoServiceAccount:    cfg.ImageBuilderConfig.KanikoServiceAccount,
		KanikoAdditionalArgs:    cfg.ImageBuilderConfig.KanikoAdditionalArgs,
		DefaultResources:        cfg.ImageBuilderConfig.DefaultResources,
		Tolerations:             cfg.ImageBuilderConfig.Tolerations,
		NodeSelectors:           cfg.ImageBuilderConfig.NodeSelectors,
		MaximumRetry:            cfg.ImageBuilderConfig.MaximumRetry,
		SafeToEvict:             cfg.ImageBuilderConfig.SafeToEvict,
		ClusterName:             cfg.ImageBuilderConfig.ClusterName,
		GcpProject:              cfg.ImageBuilderConfig.GcpProject,
		Environment:             cfg.Environment,
		SupportedPythonVersions: cfg.ImageBuilderConfig.SupportedPythonVersions,
	}
	webserviceBuilder = imagebuilder.NewModelServiceImageBuilder(kubeClient, webServiceConfig, artifactService)

	predJobConfig := imagebuilder.Config{
		BaseImage:               cfg.ImageBuilderConfig.PredictionJobBaseImage,
		BuildNamespace:          cfg.ImageBuilderConfig.BuildNamespace,
		DockerRegistry:          cfg.ImageBuilderConfig.DockerRegistry,
		BuildTimeoutDuration:    timeout,
		KanikoImage:             cfg.ImageBuilderConfig.KanikoImage,
		KanikoServiceAccount:    cfg.ImageBuilderConfig.KanikoServiceAccount,
		KanikoAdditionalArgs:    cfg.ImageBuilderConfig.KanikoAdditionalArgs,
		DefaultResources:        cfg.ImageBuilderConfig.DefaultResources,
		Tolerations:             cfg.ImageBuilderConfig.Tolerations,
		NodeSelectors:           cfg.ImageBuilderConfig.NodeSelectors,
		MaximumRetry:            cfg.ImageBuilderConfig.MaximumRetry,
		SafeToEvict:             cfg.ImageBuilderConfig.SafeToEvict,
		ClusterName:             cfg.ImageBuilderConfig.ClusterName,
		GcpProject:              cfg.ImageBuilderConfig.GcpProject,
		Environment:             cfg.Environment,
		SupportedPythonVersions: cfg.ImageBuilderConfig.SupportedPythonVersions,
	}
	predJobBuilder = imagebuilder.NewPredictionJobImageBuilder(kubeClient, predJobConfig, artifactService)

	ctl, err := cluster.NewController(
		clusterCfg,
		config.DeploymentConfig{}, // We don't need deployment config here because we're going to retrieve the log not deploy model.
	)
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

	for _, envCfg := range cfg.ClusterConfig.EnvironmentConfigs {
		var isDefault *bool = nil
		if envCfg.IsDefault {
			isDefault = &envCfg.IsDefault
		}

		var isDefaultPredictionJob *bool = nil
		if envCfg.IsDefaultPredictionJob {
			isDefaultPredictionJob = &envCfg.IsDefaultPredictionJob
		}

		deploymentCfg := config.ParseDeploymentConfig(envCfg, cfg)

		env, err := envSvc.GetEnvironment(envCfg.Name)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Panicf("unable to get environment %s: %v", envCfg.Name, err)
			}

			// Create new environment

			log.Infof("adding environment %s: cluster: %s, is_default: %v", envCfg.Name, envCfg.Cluster, envCfg.IsDefault)
			env = &models.Environment{
				Name:              envCfg.Name,
				Cluster:           envCfg.Cluster,
				IsDefault:         isDefault,
				Region:            envCfg.Region,
				GcpProject:        envCfg.GcpProject,
				MaxCPU:            envCfg.MaxCPU,
				MaxMemory:         envCfg.MaxMemory,
				MaxAllowedReplica: envCfg.MaxAllowedReplica,
				GPUs:              models.ParseGPUsConfig(envCfg.GPUs),
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    deploymentCfg.DefaultModelResourceRequests.MinReplica,
					MaxReplica:    deploymentCfg.DefaultModelResourceRequests.MaxReplica,
					CPURequest:    deploymentCfg.DefaultModelResourceRequests.CPURequest,
					MemoryRequest: deploymentCfg.DefaultModelResourceRequests.MemoryRequest,
				},
				DefaultTransformerResourceRequest: &models.ResourceRequest{
					MinReplica:    deploymentCfg.DefaultTransformerResourceRequests.MinReplica,
					MaxReplica:    deploymentCfg.DefaultTransformerResourceRequests.MaxReplica,
					CPURequest:    deploymentCfg.DefaultTransformerResourceRequests.CPURequest,
					MemoryRequest: deploymentCfg.DefaultTransformerResourceRequests.MemoryRequest,
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
			env.MaxAllowedReplica = envCfg.MaxAllowedReplica
			env.GPUs = models.ParseGPUsConfig(envCfg.GPUs)
			env.DefaultResourceRequest = &models.ResourceRequest{
				MinReplica:    deploymentCfg.DefaultModelResourceRequests.MinReplica,
				MaxReplica:    deploymentCfg.DefaultModelResourceRequests.MaxReplica,
				CPURequest:    deploymentCfg.DefaultModelResourceRequests.CPURequest,
				MemoryRequest: deploymentCfg.DefaultModelResourceRequests.MemoryRequest,
			}
			env.DefaultTransformerResourceRequest = &models.ResourceRequest{
				MinReplica:    deploymentCfg.DefaultTransformerResourceRequests.MinReplica,
				MaxReplica:    deploymentCfg.DefaultTransformerResourceRequests.MaxReplica,
				CPURequest:    deploymentCfg.DefaultTransformerResourceRequests.CPURequest,
				MemoryRequest: deploymentCfg.DefaultTransformerResourceRequests.MemoryRequest,
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

func initModelEndpointService(cfg *config.Config, db *gorm.DB, observabilityEvent event.EventProducer) service.ModelEndpointsService {
	istioClients := make(map[string]istio.Client)
	for _, env := range cfg.ClusterConfig.EnvironmentConfigs {
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

	return service.NewModelEndpointsService(istioClients, storage.NewModelEndpointStorage(db), storage.NewVersionEndpointStorage(db), cfg.Environment, observabilityEvent)
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
	for _, env := range cfg.ClusterConfig.EnvironmentConfigs {
		if !env.IsPredictionJobEnabled {
			continue
		}

		creds := mlpcluster.NewK8sClusterCreds(env.K8sConfig)
		clusterName := env.Cluster

		var restConfig *rest.Config
		var err error
		if cfg.ClusterConfig.InClusterConfig {
			restConfig, err = rest.InClusterConfig()
			if err != nil {
				log.Panicf("unable to get in cluster configs: %v", err)
			}
		} else {
			restConfig, err = creds.ToRestConfig()
			if err != nil {
				log.Panicf("unable to get cluster config of cluster: %s %v", clusterName, err)
			}
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

		batchJobTemplator := batch.NewBatchJobTemplater(cfg.BatchConfig)

		ctl := batch.NewController(predictionJobStorage, mlpAPIClient, sparkClient, kubeClient, manifestManager,
			envMetadata, batchJobTemplator)
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

func initModelServiceDeployment(cfg *config.Config, builder imagebuilder.ImageBuilder, controllers map[string]cluster.Controller, db *gorm.DB, observabilityEvent event.EventProducer) *work.ModelServiceDeployment {
	return &work.ModelServiceDeployment{
		ClusterControllers:         controllers,
		ImageBuilder:               builder,
		Storage:                    storage.NewVersionEndpointStorage(db),
		DeploymentStorage:          storage.NewDeploymentStorage(db),
		LoggerDestinationURL:       cfg.LoggerDestinationURL,
		MLObsLoggerDestinationURL:  cfg.MLObsLoggerDestinationURL,
		ObservabilityEventProducer: observabilityEvent,
	}
}

func initObservabilityPublisherDeployment(cfg *config.Config, observabilityPublisherStorage storage.ObservabilityPublisherStorage) *work.ObservabilityPublisherDeployment {
	var envCfg *config.EnvironmentConfig
	for _, env := range cfg.ClusterConfig.EnvironmentConfigs {
		if env.Name == cfg.ObservabilityPublisher.EnvironmentName {
			envCfg = env
			break
		}
	}
	if envCfg == nil {
		log.Panicf("could not find destination environment for observability publisher")
	}

	clusterCfg := cluster.Config{
		ClusterName: envCfg.Cluster,
		GcpProject:  envCfg.GcpProject,
	}

	var restConfig *rest.Config
	var err error
	if cfg.ClusterConfig.InClusterConfig {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Panicf("unable to get in cluster configs: %v", err)
		}
	} else {
		creds := mlpcluster.NewK8sClusterCreds(envCfg.K8sConfig)
		restConfig, err = creds.ToRestConfig()
		if err != nil {
			log.Panicf("unable to get cluster config of cluster: %s %v", clusterCfg.ClusterName, err)
		}
	}
	deployer, err := deployment.New(restConfig, cfg.ObservabilityPublisher)
	if err != nil {
		log.Panicf("unable to initialize observability deployer with err: %w", err)
	}

	return &work.ObservabilityPublisherDeployment{
		Deployer:                      deployer,
		ObservabilityPublisherStorage: observabilityPublisherStorage,
	}
}

func initClusterControllers(cfg *config.Config) map[string]cluster.Controller {
	controllers := make(map[string]cluster.Controller)
	for _, env := range cfg.ClusterConfig.EnvironmentConfigs {
		clusterCfg := cluster.Config{
			ClusterName: env.Cluster,
			GcpProject:  env.GcpProject,
		}

		if cfg.ClusterConfig.InClusterConfig {
			clusterCfg.InClusterConfig = true
		} else {
			clusterCfg.Credentials = mlpcluster.NewK8sClusterCreds(env.K8sConfig)
		}

		ctl, err := cluster.NewController(
			clusterCfg,
			config.ParseDeploymentConfig(env, cfg),
		)
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
		MLObsLoggerDestinationURL: cfg.MLObsLoggerDestinationURL,
		JobProducer:               producer,
		FeastCoreClient:           feastCoreClient,
		StandardTransformerConfig: cfg.StandardTransformerConfig,
	})
}

func initLogService(cfg *config.Config) service.LogService {
	clusterCfg := cluster.Config{
		ClusterName: cfg.ImageBuilderConfig.ClusterName,
		GcpProject:  cfg.ImageBuilderConfig.GcpProject,
	}

	if cfg.ClusterConfig.InClusterConfig {
		clusterCfg.InClusterConfig = true
	} else {
		clusterCfg.Credentials = mlpcluster.NewK8sClusterCreds(cfg.ImageBuilderConfig.K8sConfig)
	}

	ctl, err := cluster.NewController(
		clusterCfg,
		config.DeploymentConfig{}, // We don't need deployment config here because we're going to retrieve the log not deploy model.
	)
	if err != nil {
		log.Panicf("unable to initialize cluster controller %v", err)
	}

	clusterControllers := make(map[string]cluster.Controller)
	clusterControllers[cfg.ImageBuilderConfig.ClusterName] = ctl

	for _, env := range cfg.ClusterConfig.EnvironmentConfigs {
		clusterCfg := cluster.Config{
			ClusterName: env.Cluster,
			GcpProject:  env.GcpProject,
		}

		if cfg.ClusterConfig.InClusterConfig {
			clusterCfg.InClusterConfig = true
		} else {
			clusterCfg.Credentials = mlpcluster.NewK8sClusterCreds(env.K8sConfig)
		}

		ctl, err := cluster.NewController(
			clusterCfg,
			config.ParseDeploymentConfig(env, cfg),
		)
		if err != nil {
			log.Panicf("unable to initialize cluster controller %v", err)
		}

		clusterControllers[env.Cluster] = ctl
	}

	return service.NewLogService(clusterControllers)
}
