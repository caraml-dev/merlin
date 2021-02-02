package main

import (
	"log"
	"net"
	"strconv"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/server"
)

func init() {
	prometheus.MustRegister(version.NewCollector("feast_transformer"))
}

func main() {
	cfg := struct {
		Server           server.Options
		Feast            feast.Options
		MonitoringOption feast.FeatureMonitoringOptions

		StandardTransformerConfigJSON string `envconfig:"STANDARD_TRANSFORMER_CONFIG" required:"true"`

		LogLevel string `envconfig:"LOG_LEVEL"`
	}{}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalln(errors.Wrap(err, "Error processing environment variables"))
	}

	logger, _ := zap.NewProduction()
	if cfg.LogLevel == "DEBUG" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("configuration loaded", zap.Any("cfg", cfg))

	config := &transformer.StandardTransformerConfig{}
	if err := jsonpb.UnmarshalString(cfg.StandardTransformerConfigJSON, config); err != nil {
		log.Fatalln(errors.Wrap(err, "Unable to parse standard transformer config"))
	}

	feastHost, feastPort, err := net.SplitHostPort(cfg.Feast.ServingURL)
	if err != nil {
		log.Panicf("Unable to parse feast serving URL %s: %v", cfg.Feast.ServingURL, err)
	}

	feastPortInt, err := strconv.Atoi(feastPort)
	if err != nil {
		log.Panicf("Unable to parse feast serving port %s: %v", feastPort, err)
	}
	feastClient, err := feastSdk.NewGrpcClient(feastHost, feastPortInt)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "Unable to initialie feastSdk client"))
	}

	f := feast.NewTransformer(feastClient, config, &cfg.MonitoringOption, logger)

	s := server.New(&cfg.Server, logger)
	s.PreprocessHandler = f.Transform

	s.Run()
}
