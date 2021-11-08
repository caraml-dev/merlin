package main

import (
	"encoding/json"
	"io"
	"log"
	"time"

	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kelseyhightower/envconfig"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	jcfg "github.com/uber/jaeger-client-go/config"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/gojek/merlin/pkg/hystrix"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
)

func init() {
	prometheus.MustRegister(version.NewCollector("transformer"))

	hystrixCollector := hystrix.NewPrometheusCollector("hystrix", map[string]string{"app": "transformer"})
	metricCollector.Registry.Register(hystrixCollector)
}

type AppConfig struct {
	Server server.Options
	Feast  feast.Options

	StandardTransformerConfigJSON string `envconfig:"STANDARD_TRANSFORMER_CONFIG" required:"true"`
	FeatureTableSpecJsons         string `envconfig:"FEAST_FEATURE_TABLE_SPECS_JSONS" default:""`
	LogLevel                      string `envconfig:"LOG_LEVEL"`
	RedisOverwriteConfig          RedisOverwriteConfig

	// By default the value is 0, users should configure this value below the memory requested
	InitHeapSizeInMB int `envconfig:"INIT_HEAP_SIZE_IN_MB" default:"0"`
}

type RedisOverwriteConfig struct {
	RedisDirectStorageEnabled *bool          `envconfig:"FEAST_REDIS_DIRECT_STORAGE_ENABLED"`
	PoolSize                  *int32         `envconfig:"FEAST_REDIS_POOL_SIZE"`
	ReadTimeout               *time.Duration `envconfig:"FEAST_REDIS_READ_TIMEOUT"`
	WriteTimeout              *time.Duration `envconfig:"FEAST_REDIS_WRITE_TIMEOUT"`
}

// Trick GC frequency based on this https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap-26c2462549a2/
var initialHeapAllocation []byte

func main() {
	appConfig := AppConfig{}
	if err := envconfig.Process("", &appConfig); err != nil {
		log.Fatal(errors.Wrap(err, "Error processing environment variables"))
	}

	if appConfig.InitHeapSizeInMB > 0 {
		initialHeapAllocation = make([]byte, appConfig.InitHeapSizeInMB<<20)
	}

	logger, _ := zap.NewProduction()
	if appConfig.LogLevel == "DEBUG" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("configuration loaded", zap.Any("appConfig", appConfig))

	closer, err := initTracing(appConfig.Server.ModelName + "-transformer")
	if err != nil {
		logger.Error("Unable to initialize tracing", zap.Error(err))
	}
	if closer != nil {
		defer closer.Close()
	}

	transformerConfig := &spec.StandardTransformerConfig{}
	if err := jsonpb.UnmarshalString(appConfig.StandardTransformerConfigJSON, transformerConfig); err != nil {
		logger.Fatal("Unable to parse standard transformer transformerConfig", zap.Error(err))
	}

	featureSpecs := make([]*spec.FeatureTableMetadata, 0)
	if appConfig.FeatureTableSpecJsons != "" {
		featureTableSpecJsons := make([]map[string]interface{}, 0)
		if err := json.Unmarshal([]byte(appConfig.FeatureTableSpecJsons), &featureTableSpecJsons); err != nil {
			panic(err)
		}
		for _, specJson := range featureTableSpecJsons {
			s, err := json.Marshal(specJson)
			if err != nil {
				panic(err)
			}
			featureSpec := &spec.FeatureTableMetadata{}
			err = jsonpb.UnmarshalString(string(s), featureSpec)
			if err != nil {
				return
			}

			if err != nil {
				logger.Fatal("Unable to parse standard transformer featureTableSpecs config", zap.Error(err))
			}
			featureSpecs = append(featureSpecs, featureSpec)
		}
	}
	feastSources := feast.GetFeastServingSources(transformerConfig)
	feastOpts := overwriteFeastOptionsConfig(appConfig)
	logger.Info("feast options", zap.Any("val", feastOpts))

	feastServingClients, err := feast.InitFeastServingClients(feastOpts, featureSpecs, feastSources, logger)
	if err != nil {
		logger.Fatal("Unable to initialize Feast Clients", zap.Error(err))
	}

	s := server.New(&appConfig.Server, logger)

	if transformerConfig.TransformerConfig.Feast != nil {
		// Feast Enricher
		feastTransformer, err := initFeastTransformer(appConfig, feastServingClients, transformerConfig, logger)
		if err != nil {
			logger.Fatal("Unable to initialize transformer", zap.Error(err))
		}
		s.PreprocessHandler = feastTransformer.Enrich
	} else {
		// Standard Enricher
		compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastServingClients, &feastOpts, logger)
		compiledPipeline, err := compiler.Compile(transformerConfig)
		if err != nil {
			logger.Fatal("Unable to compile standard transformer", zap.Error(err))
		}

		handler := pipeline.NewHandler(compiledPipeline, logger)
		s.PreprocessHandler = handler.Preprocess
		s.PostprocessHandler = handler.Postprocess
		s.ContextModifier = handler.EmbedEnvironment
	}

	s.Run()
}

func getServingType(userEnableDirectStorage *bool, currentServingType spec.ServingType) spec.ServingType {
	servingType := currentServingType
	servingTypeMap := map[bool]spec.ServingType{true: spec.ServingType_DIRECT_STORAGE, false: spec.ServingType_FEAST_GRPC}
	if userEnableDirectStorage != nil {
		servingType = servingTypeMap[*userEnableDirectStorage]
	}
	return servingType
}

func overwriteFeastOptionsConfig(appConfig AppConfig) feast.Options {
	feastOptions := appConfig.Feast
	redisOverwriteCfg := appConfig.RedisOverwriteConfig

	for _, storage := range feastOptions.StorageConfigs {
		switch storage.Storage.(type) {
		case *spec.OnlineStorage_Redis:
			storage.ServingType = getServingType(redisOverwriteCfg.RedisDirectStorageEnabled, storage.ServingType)
			redisStorage := storage.GetRedis()
			overwriteRedisOption(redisStorage.Option, redisOverwriteCfg)
		case *spec.OnlineStorage_RedisCluster:
			storage.ServingType = getServingType(redisOverwriteCfg.RedisDirectStorageEnabled, storage.ServingType)
			redisClusterStorage := storage.GetRedisCluster()
			overwriteRedisOption(redisClusterStorage.Option, redisOverwriteCfg)
		}
	}
	return feastOptions
}

func overwriteRedisOption(opts *spec.RedisOption, redisOverwriteConfig RedisOverwriteConfig) {
	if redisOverwriteConfig.PoolSize != nil {
		opts.PoolSize = *redisOverwriteConfig.PoolSize
	}
	if redisOverwriteConfig.ReadTimeout != nil {
		opts.ReadTimeout = durationpb.New(*redisOverwriteConfig.ReadTimeout)
	}
	if redisOverwriteConfig.WriteTimeout != nil {
		opts.WriteTimeout = durationpb.New(*redisOverwriteConfig.WriteTimeout)
	}
}

func initFeastTransformer(appCfg AppConfig,
	feastClient feast.Clients,
	transformerConfig *spec.StandardTransformerConfig,
	logger *zap.Logger) (*feast.Enricher, error) {

	compiledJSONPaths, err := feast.CompileJSONPaths(transformerConfig.TransformerConfig.Feast)
	if err != nil {
		return nil, err
	}

	compiledExpressions, err := feast.CompileExpressions(transformerConfig.TransformerConfig.Feast, symbol.NewRegistry())
	if err != nil {
		return nil, err
	}

	jsonPathStorage := jsonpath.NewStorage()
	jsonPathStorage.AddAll(compiledJSONPaths)
	expressionStorage := expression.NewStorage()
	expressionStorage.AddAll(compiledExpressions)
	entityExtractor := feast.NewEntityExtractor(jsonPathStorage, expressionStorage)
	featureRetriever := feast.NewFeastRetriever(feastClient,
		entityExtractor,
		transformerConfig.TransformerConfig.Feast,
		&appCfg.Feast,
		logger,
	)

	return feast.NewEnricher(featureRetriever, logger)
}

func initTracing(serviceName string) (io.Closer, error) {
	// Set tracing configuration defaults.
	cfg := &jcfg.Configuration{
		ServiceName: serviceName,
	}

	// Available options can be seen here:
	// https://github.com/jaegertracing/jaeger-client-go#environment-variables
	cfg, err := cfg.FromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get tracing config from environment")
	}

	tracer, closer, err := cfg.NewTracer(
		jcfg.Metrics(jprom.New()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize tracer")
	}

	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}
