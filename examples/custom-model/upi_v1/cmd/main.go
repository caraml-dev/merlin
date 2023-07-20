package main

import (
	"custom-model-upi/pkg/server"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/log"
)

func main() {
	cfg := server.Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Panicf(err.Error())
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	instrumentRouter := server.NewInstrumentationRouter()
	server.RunUPIServer(&cfg, instrumentRouter, logger)
}
