package main

import (
	"context"
	"encoding/json"
	"log"

	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/caraml-dev/merlin/pkg/hystrix"
	"github.com/caraml-dev/merlin/pkg/kafka"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/pipeline"
	serverConf "github.com/caraml-dev/merlin/pkg/transformer/server/config"
	grpc "github.com/caraml-dev/merlin/pkg/transformer/server/grpc"
	rest "github.com/caraml-dev/merlin/pkg/transformer/server/rest"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	"github.com/caraml-dev/merlin/pkg/transformer/types/expression"
)

func init() {
	prometheus.MustRegister(version.NewCollector("transformer"))

	hystrixCollector := hystrix.NewPrometheusCollector("hystrix", map[string]string{"app": "transformer"})
	metricCollector.Registry.Register(hystrixCollector)
}

// AppConfig holds configuration for standard transformer
type AppConfig struct {
	// Server configuration either HTTP or GRPC
	Server serverConf.Options
	// Feast configuration
	Feast feast.Options
	// StandardTransformerConfigJSON is standard transformer configuration in JSON string format
	StandardTransformerConfigJSON string `envconfig:"STANDARD_TRANSFORMER_CONFIG" required:"true"`
	// FeatureTableSpecJsons is feature table metadata specs in JSON string format
	FeatureTableSpecJsons string `envconfig:"FEAST_FEATURE_TABLE_SPECS_JSONS"`
	// LogLevel
	LogLevel string `envconfig:"LOG_LEVEL"`
	// RedisOverwriteConfig is user configuration for redis that will overwrite default configuration supplied by merlin
	RedisOverwriteConfig feast.RedisOverwriteConfig
	// BigtableOverwriteConfig is user configuration for bigtable that will overwrite default configuration supplied by merlin
	BigtableOverwriteConfig feast.BigtableOverwriteConfig
	// Kafka config
	KafkaConfig kafka.Config

	// By default the value is 0, users should configure this value below the memory requested
	InitHeapSizeInMB int `envconfig:"INIT_HEAP_SIZE_IN_MB" default:"0"`

	Tracing JaegerTracing
}

// JaegerConfig holds configuration about jaeger tracing
type JaegerTracing struct {
	// Disabled the tracing, will create NoOpTracerProvider
	Disabled bool `envconfig:"JAEGER_DISABLED" default:"true"`
	// Collector endpoint URL
	CollectorURL string `envconfig:"JAEGER_COLLECTOR_URL"`
	// Probability the trace will be sampled
	SamplerProbability float64 `envconfig:"JAEGER_SAMPLER_PARAM" default:"0.1"`
}

// Trick GC frequency based on this https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap-26c2462549a2/
var initialHeapAllocation []byte //nolint:unused

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
	defer logger.Sync() //nolint:errcheck

	logger.Info("configuration loaded", zap.Any("appConfig", appConfig))

	tracingProvider, err := initTracing(appConfig.Server.ModelFullName+"-transformer", appConfig.Tracing)
	if err != nil {
		logger.Error("Unable to initialize tracing", zap.Error(err))
	}
	otel.SetTracerProvider(tracingProvider)
	propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader))
	otel.SetTextMapPropagator(propagator)

	defer func() {
		if err := tracingProvider.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	transformerConfig := &spec.StandardTransformerConfig{}
	if err := protojson.Unmarshal([]byte(appConfig.StandardTransformerConfigJSON), transformerConfig); err != nil {
		logger.Fatal("unable to parse standard transformer transformerConfig", zap.Error(err))
	}

	featureTableMetadata, err := parseFeatureTableMetadata(appConfig.FeatureTableSpecJsons)
	if err != nil {
		logger.Fatal("unable to parse feature table metadata", zap.Error(err))
	}

	if transformerConfig.TransformerConfig.Feast != nil {
		// Feast Enricher
		runFeastEnricherServer(appConfig, transformerConfig, featureTableMetadata, logger)
		return
	}

	opts := []pipeline.CompilerOptions{
		pipeline.WithProtocol(appConfig.Server.Protocol),
		pipeline.WithLogger(logger),
	}

	predictionLogConfig := transformerConfig.PredictionLogConfig
	if predictionLogConfig != nil && predictionLogConfig.Enable && appConfig.Server.Protocol == protocol.UpiV1 {
		producer, err := kafka.NewProducer(appConfig.KafkaConfig, logger)
		if err != nil {
			logger.Fatal("failed to initialize kafka producer", zap.Error(err))
		}

		opts = append(opts, pipeline.WithPredictionLogProducer(producer))

		defer producer.Close()
	}

	handler, err := createPipelineHandler(
		appConfig,
		transformerConfig,
		featureTableMetadata,
		logger,
		opts...,
	)
	if err != nil {
		logger.Fatal("got error when creating handler", zap.Error(err))
	}

	if appConfig.Server.Protocol == protocol.UpiV1 {
		instRouter := rest.NewInstrumentationRouter()
		runGrpcServer(&appConfig.Server, handler, instRouter, logger)
	} else {
		runHTTPServer(&appConfig.Server, handler, logger)
	}
}

// TODO: Feast enricher will be deprecated soon all associated functions will be deleted
func runFeastEnricherServer(appConfig AppConfig, transformerConfig *spec.StandardTransformerConfig, featureTableMetadata []*spec.FeatureTableMetadata, logger *zap.Logger) {
	feastOpts := feast.OverwriteFeastOptionsConfig(appConfig.Feast, appConfig.RedisOverwriteConfig, appConfig.BigtableOverwriteConfig)
	logger.Info("feast options", zap.Any("val", feastOpts))

	feastServingClients, err := feast.InitFeastServingClients(feastOpts, featureTableMetadata, transformerConfig)
	if err != nil {
		logger.Fatal("unable to initialize feast clients", zap.Error(err))
	}
	feastTransformer, err := initFeastTransformer(appConfig, feastServingClients, transformerConfig, logger)
	if err != nil {
		logger.Fatal("Unable to initialize transformer", zap.Error(err))
	}
	feastEnricherServer := rest.New(&appConfig.Server, logger)
	feastEnricherServer.PreprocessHandler = feastTransformer.Enrich
	feastEnricherServer.Run()
}

// TODO: Feast enricher will be deprecated soon all associated functions will be deleted
func initFeastTransformer(appCfg AppConfig,
	feastClient feast.Clients,
	transformerConfig *spec.StandardTransformerConfig,
	logger *zap.Logger,
) (*feast.Enricher, error) {
	// for feast enricher the only supported protocol is http_json hence json source type is Map
	compiledJSONPaths, err := feast.CompileJSONPaths(transformerConfig.TransformerConfig.Feast, jsonpath.Map)
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

func parseFeatureTableMetadata(featureTableSpecsJson string) ([]*spec.FeatureTableMetadata, error) {
	featureSpecs := make([]*spec.FeatureTableMetadata, 0)
	if featureTableSpecsJson != "" {
		featureTableSpec := make([]map[string]interface{}, 0)
		if err := json.Unmarshal([]byte(featureTableSpecsJson), &featureTableSpec); err != nil {
			return nil, err
		}
		for _, specJson := range featureTableSpec {
			s, err := json.Marshal(specJson)
			if err != nil {
				return nil, err
			}
			featureSpec := &spec.FeatureTableMetadata{}
			err = protojson.Unmarshal(s, featureSpec)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse standard transformer featureTableSpecs config")
			}
			featureSpecs = append(featureSpecs, featureSpec)
		}
	}
	return featureSpecs, nil
}

func createPipelineHandler(appConfig AppConfig, transformerConfig *spec.StandardTransformerConfig, featureTableMetadata []*spec.FeatureTableMetadata, logger *zap.Logger, options ...pipeline.CompilerOptions) (*pipeline.Handler, error) {
	feastOpts := feast.OverwriteFeastOptionsConfig(appConfig.Feast, appConfig.RedisOverwriteConfig, appConfig.BigtableOverwriteConfig)
	logger.Info("feast options", zap.Any("val", feastOpts))

	feastServingClients, err := feast.InitFeastServingClients(feastOpts, featureTableMetadata, transformerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize feast clients")
	}

	compiler := pipeline.NewCompiler(
		symbol.NewRegistry(),
		feastServingClients,
		&feastOpts,
		options...,
	)
	compiledPipeline, err := compiler.Compile(transformerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compile standard transformer")
	}

	handler := pipeline.NewHandler(compiledPipeline, logger)
	return handler, nil
}

func runHTTPServer(opts *serverConf.Options, handler *pipeline.Handler, logger *zap.Logger) {
	s := rest.NewWithHandler(opts, handler, logger)
	s.Run()
}

func runGrpcServer(opts *serverConf.Options, handler *pipeline.Handler, instrumentationRouter *mux.Router, logger *zap.Logger) {
	s, err := grpc.NewUPIServer(opts, handler, instrumentationRouter, logger)
	if err != nil {
		panic(err)
	}
	s.Run()
}

func initTracing(serviceName string, tracingCfg JaegerTracing) (*tracesdk.TracerProvider, error) {
	if tracingCfg.Disabled {
		return tracesdk.NewTracerProvider(tracesdk.WithSampler(tracesdk.NeverSample())), nil
	}
	// Create the Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(tracingCfg.CollectorURL)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(tracingCfg.SamplerProbability)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	return tp, nil
}
