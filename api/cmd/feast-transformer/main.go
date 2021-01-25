package main

import (
	"encoding/json"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	"github.com/gojek/merlin/pkg/transformers/feast"
	feastSpec "github.com/gojek/merlin/pkg/transformers/feast/spec"
	"github.com/gojek/merlin/pkg/transformers/server"
)

type TransformerConfig struct {
	Config struct {
		Feast []feastSpec.RequestParsing `json:"feast"`
	} `json:"transformerConfig"`
}

func (tc *TransformerConfig) Decode(value string) error {
	return json.Unmarshal([]byte(value), tc)
}

func init() {
	prometheus.MustRegister(version.NewCollector("feast_transformer"))
}

func main() {
	cfg := struct {
		Server server.Options
		Feast  feast.Options

		RequestParsingConfig TransformerConfig `envconfig:"REQUEST_PARSING_CONFIGURATION" required:"true"`
	}{}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalln(errors.Wrap(err, "Error processing environment variables"))
	}
	log.Printf("Configuration: %+v", cfg)

	log.Println("Starting Feast Transformer...")

	f := feast.New(&cfg.Feast, cfg.RequestParsingConfig.Config.Feast)

	s := server.New(&cfg.Server)
	s.PreprocessHandler = f.TransformHandler
	s.Run()
}
