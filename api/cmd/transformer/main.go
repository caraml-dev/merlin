package main

import (
	"encoding/json"
	"io"
	"log"

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

	"github.com/gojek/merlin/pkg/hystrix"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
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
	RedisOverwriteConfig          feast.RedisOverwriteConfig
	BigtableOverwriteConfig       feast.BigtableOverwriteConfig

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

	closer, err := initTracing(appConfig.Server.ModelFullName + "-transformer")
	if err != nil {
		logger.Error("Unable to initialize tracing", zap.Error(err))
	}
	if closer != nil {
		defer closer.Close()
	}

	handler, err := createPipelineHandler(appConfig, logger)
	if err != nil {
		logger.Fatal("Got error when creating handler", zap.Error(err))
	}

	runHTTPServer(&appConfig.Server, handler, logger)
	runGrpcServer(&appConfig.Server, handler, logger)
}

func createPipelineHandler(appConfig AppConfig, logger *zap.Logger) (*pipeline.Handler, error) {
	transformerConfig := &spec.StandardTransformerConfig{}
	if err := jsonpb.UnmarshalString(appConfig.StandardTransformerConfigJSON, transformerConfig); err != nil {
		return nil, errors.Wrap(err, "unable to parse standard transformer transformerConfig")
	}

	// TODO Refactor this
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
				return nil, errors.Wrap(err, "unable to parse standard transformer featureTableSpecs config")
			}
			featureSpecs = append(featureSpecs, featureSpec)
		}
	}

	feastOpts := feast.OverwriteFeastOptionsConfig(appConfig.Feast, appConfig.RedisOverwriteConfig, appConfig.BigtableOverwriteConfig)
	logger.Info("feast options", zap.Any("val", feastOpts))

	feastServingClients, err := feast.InitFeastServingClients(feastOpts, featureSpecs, transformerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize feast clients")
	}

	compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastServingClients, &feastOpts, logger, false)
	compiledPipeline, err := compiler.Compile(transformerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compile standard transformer")
	}

	handler := pipeline.NewHandler(compiledPipeline, logger)
	return handler, nil
}

func runHTTPServer(opts *server.Options, handler *pipeline.Handler, logger *zap.Logger) {
	s := server.NewWithHandler(opts, handler, logger)
	s.Run()
}

func runGrpcServer(opts *server.Options, handler *pipeline.Handler, logger *zap.Logger) error {
	s, err := server.NewUPIServer(opts, handler, logger)
	if err != nil {
		return err
	}
	s.RunServer()
	return nil
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
