package main

import (
	"io"
	"log"
	"net"
	"strconv"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kelseyhightower/envconfig"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	jcfg "github.com/uber/jaeger-client-go/config"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
)

func init() {
	prometheus.MustRegister(version.NewCollector("feast_transformer"))
}

type AppConfig struct {
	Server server.Options
	Feast  feast.Options
	Cache  cache.Options

	StandardTransformerConfigJSON string `envconfig:"STANDARD_TRANSFORMER_CONFIG" required:"true"`
	LogLevel                      string `envconfig:"LOG_LEVEL"`
}

func main() {
	appConfig := AppConfig{}
	if err := envconfig.Process("", &appConfig); err != nil {
		log.Fatal(errors.Wrap(err, "Error processing environment variables"))
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

	feastHost, feastPort, err := net.SplitHostPort(appConfig.Feast.ServingURL)
	if err != nil {
		logger.Fatal("Unable to parse Feast Serving URL", zap.String("feast-serving-url", appConfig.Feast.ServingURL), zap.Error(err))
	}
	feastPortInt, err := strconv.Atoi(feastPort)
	if err != nil {
		logger.Fatal("Unable to parse Feast Serving Port", zap.String("feast-serving-port", feastPort), zap.Error(err))
	}

	feastClient, err := feastSdk.NewGrpcClient(feastHost, feastPortInt)
	if err != nil {
		logger.Fatal("Unable to initialize Feast client", zap.Error(err))
	}

	s := server.New(&appConfig.Server, logger)

	if transformerConfig.TransformerConfig.Feast != nil {
		// Feast Enricher
		feastTransformer, err := initFeastTransformer(appConfig, feastClient, transformerConfig, logger)
		if err != nil {
			logger.Fatal("Unable to initialize transformer", zap.Error(err))
		}
		s.PreprocessHandler = feastTransformer.Enrich
	} else {
		// Standard Enricher
		compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastClient, &appConfig.Feast, &appConfig.Cache, logger)
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

func initFeastTransformer(appCfg AppConfig,
	feastClient *feastSdk.GrpcClient,
	transformerConfig *spec.StandardTransformerConfig,
	logger *zap.Logger) (*feast.Enricher, error) {

	var memoryCache cache.Cache
	if appCfg.Feast.CacheEnabled {
		memoryCache = cache.NewInMemoryCache(&appCfg.Cache)
	}

	compiledJSONPaths, err := feast.CompileJSONPaths(transformerConfig.TransformerConfig.Feast)
	if err != nil {
		return nil, err
	}

	compiledExpressions, err := feast.CompileExpressions(transformerConfig.TransformerConfig.Feast)
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
		memoryCache,
		logger,
	)

	return feast.NewEnricher(featureRetriever, logger)
}

func initTracing(serviceName string) (io.Closer, error) {
	// Set tracing configuration defaults.
	cfg := &jcfg.Configuration{
		ServiceName: serviceName,
		Disabled:    true,
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
