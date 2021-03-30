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

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/server"
)

func init() {
	prometheus.MustRegister(version.NewCollector("feast_transformer"))
}

func main() {
	cfg := struct {
		Server server.Options
		Feast  feast.Options
		Cache  cache.Options

		StandardTransformerConfigJSON string `envconfig:"STANDARD_TRANSFORMER_CONFIG" required:"true"`
		LogLevel                      string `envconfig:"LOG_LEVEL"`
	}{}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "Error processing environment variables"))
	}

	logger, _ := zap.NewProduction()
	if cfg.LogLevel == "DEBUG" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("configuration loaded", zap.Any("cfg", cfg))

	closer, err := initTracing(cfg.Server.ModelName + "-transformer")
	if err != nil {
		logger.Error("Unable to initialize tracing", zap.Error(err))
	}
	if closer != nil {
		defer closer.Close()
	}

	config := &transformer.StandardTransformerConfig{}
	if err := jsonpb.UnmarshalString(cfg.StandardTransformerConfigJSON, config); err != nil {
		logger.Fatal("Unable to parse standard transformer config", zap.Error(err))
	}

	feastHost, feastPort, err := net.SplitHostPort(cfg.Feast.ServingURL)
	if err != nil {
		logger.Fatal("Unable to parse Feast Serving URL", zap.String("feast-serving-url", cfg.Feast.ServingURL), zap.Error(err))
	}
	feastPortInt, err := strconv.Atoi(feastPort)
	if err != nil {
		logger.Fatal("Unable to parse Feast Serving Port", zap.String("feast-serving-port", feastPort), zap.Error(err))
	}

	cred, err := feastSdk.NewGoogleCredential(cfg.Feast.AuthAudience)
	if err != nil {
		logger.Fatal("Unable to create credentials", zap.Error(err))
	}
	feastClient, err := feastSdk.NewSecureGrpcClient(feastHost, feastPortInt, feastSdk.SecurityConfig{
		Credential: cred,
	})
	if err != nil {
		logger.Fatal("Unable to initialize Feast client", zap.Error(err))
	}

	var memoryCache *cache.Cache
	if cfg.Feast.CacheEnabled {
		memoryCache = cache.NewCache(cfg.Cache)
	}

	f, err := feast.NewTransformer(feastClient, config, &cfg.Feast, logger, memoryCache)
	if err != nil {
		logger.Fatal("Unable to initialize transformer", zap.Error(err))
	}

	s := server.New(&cfg.Server, logger)
	s.PreprocessHandler = f.Transform

	s.Run()
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
