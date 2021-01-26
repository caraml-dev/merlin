package main

import (
	"log"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

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

		StandardTransformerConfigJson string `envconfig:"STANDARD_TRANSFORMER_CONFIG" required:"true"`
	}{}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalln(errors.Wrap(err, "Error processing environment variables"))
	}
	log.Printf("Configuration: %+v", cfg)

	log.Println("Starting FeastTransformer Transformer...")

	config := &transformer.StandardTransformerConfig{}
	err := jsonpb.UnmarshalString(cfg.StandardTransformerConfigJson, config)
	if err != nil {
		log.Panicf("Unable to parse standard transformer config: %s", err.Error())
	}

	feastClient, err := feastSdk.NewGrpcClient(cfg.Feast.ServingAddress, cfg.Feast.ServingPort)
	if err != nil {
		log.Panicf("Unable to initialie feastSdk client: %s", err.Error())
	}

	f := feast.NewTransformer(feastClient, config)

	s := server.New(&cfg.Server)
	s.PreprocessHandler = f.TransformHandler
	s.Run()
}
