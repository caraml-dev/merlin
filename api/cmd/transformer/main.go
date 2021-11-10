package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"

	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kelseyhightower/envconfig"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	jcfg "github.com/uber/jaeger-client-go/config"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"

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

	// By default the value is 0, users should configure this value below the memory requested
	InitHeapSizeInMB int `envconfig:"INIT_HEAP_SIZE_IN_MB" default:"0"`
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

	featureSpecs := make([]*core.FeatureTableSpec, 0)
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
			featureSpec := &core.FeatureTableSpec{}
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
	feastServingClients, err := initFeastServingClients(appConfig.Feast, featureSpecs, logger)
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
		compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastServingClients, &appConfig.Feast, logger)
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

func initFeastServingClients(feastOptions feast.Options, specs []*core.FeatureTableSpec, logger *zap.Logger) (feast.Clients, error) {
	clients := feast.Clients{}

	// Default Feast gRPC client. Used if user does not specify serving url
	// on standard transformer config.
	client, err := newFeastGrpcClient(feastOptions.DefaultServingURL)
	if err != nil {
		return nil, err
	}
	clients[feast.URL(feast.DefaultClientURLKey)] = client
	logger.Info("Default Feast gRPC client initialized", zap.String("url", feastOptions.DefaultServingURL))

	for _, url := range feastOptions.ServingURLs {
		client, err := newFeastGrpcClient(url)
		if err != nil {
			return nil, err
		}
		clients[feast.URL(url)] = client
		logger.Info("Feast gRPC client initialized", zap.String("url", url))
	}

	return clients, nil
}

func newFeastGrpcClient(url string) (*feastSdk.GrpcClient, error) {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		return nil, errors.Errorf("Unable to parse Feast Serving host (%s): %s", url, err)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.Errorf("Unable to parse Feast Serving port (%s): %s", url, err)
	}

	client, err := feastSdk.NewGrpcClient(host, portInt)
	if err != nil {
		return nil, errors.Errorf("Unable to initialize a Feast gRPC client: %s", err)
	}

	return client, nil
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
