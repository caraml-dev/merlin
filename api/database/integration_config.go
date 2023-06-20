// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration_local || integration
// +build integration_local integration

package database

import (
	"os"

	"github.com/caraml-dev/merlin/config"
)

var (
	mainDBConfig = &config.DatabaseConfig{
		Host:          getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:          5432,
		User:          getEnvOrDefault("POSTGRES_USER", "merlin"),
		Password:      getEnvOrDefault("POSTGRES_PASSWORD", "merlin"),
		Database:      getEnvOrDefault("POSTGRES_DB", "merlin"),
		MigrationPath: "file://../../db-migrations",
	}
)

func getTemporaryDBConfig(database string) *config.DatabaseConfig {
	cfg := *mainDBConfig
	cfg.Database = database
	return &cfg
}

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
