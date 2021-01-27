package main

import (
	"log"

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
		Server server.Options
		Feast  feast.Options

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

	feastClient, err := feastSdk.NewGrpcClient(cfg.Feast.ServingAddress, cfg.Feast.ServingPort)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "Unable to initialie feastSdk client"))
	}

	f := feast.NewTransformer(feastClient, config, logger)

	s := server.New(&cfg.Server, logger)
	s.PreprocessHandler = f.Transform

	logger.Info("starting transformer")
	s.Run()
}
