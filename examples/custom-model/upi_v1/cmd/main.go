package main

import (
	"custom-model-upi/pkg/server"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

func main() {
	cfg := server.Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	instrumentRouter := server.NewInstrumentationRouter()
	server.RunUPIServer(&cfg, instrumentRouter, logger)
}
