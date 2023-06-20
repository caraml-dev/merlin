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
	"database/sql"
	"fmt"
	"testing"
	"time"

	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
)

func create(testDBCfg *config.DatabaseConfig) (*sql.DB, *sql.DB, error) {
	// Initialise main merlin DB
	connStr := connectionString(mainDBConfig)
	log.Infof("connecting to test db: %s", connStr)
	mainDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	// Use the main merlin database connection to create the temporary database.
	if _, err := mainDB.Exec("CREATE DATABASE " + testDBCfg.Database); err != nil {
		return mainDB, nil, err
	} else if testDb, err := sql.Open("postgres", connectionString(testDBCfg)); err != nil {
		if _, err := mainDB.Exec("DROP DATABASE " + testDBCfg.Database); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		}
		return mainDB, nil, err
	} else {
		return mainDB, testDb, nil
	}
}

// createTestDatabase connects to test postgreSQL instance (either local or the one
// at CI environment) and creates a new database with an up-to-date schema
func createTestDatabase() (*gorm.DB, func(), error) {
	testDBCfg := getTemporaryDBConfig(fmt.Sprintf("mlp_id_%d", time.Now().UnixNano()))
	mainDB, testDb, err := create(testDBCfg)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := testDb.Close(); err != nil {
			log.Fatalf("Failed to close connection to integration test database: \n%s", err)
		} else if _, err := mainDB.Exec("DROP DATABASE " + testDBCfg.Database); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		} else if err = mainDB.Close(); err != nil {
			log.Fatalf("Failed to close database: \n%s", err)
		}
	}

	if err = migrateDB(testDBCfg); err != nil {
		cleanup()
		return nil, nil, err
	} else if gormDb, err := gorm.Open(
		pg.New(pg.Config{Conn: testDb}),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	); err != nil {
		cleanup()
		return nil, nil, err
	} else {
		return gormDb, cleanup, nil
	}
}

// WithTestDatabase handles the lifecycle of the database creation/migration/destruction
// for a test case/suite.
func WithTestDatabase(t *testing.T, test func(t *testing.T, db *gorm.DB)) {
	if testDb, cleanupFn, err := createTestDatabase(); err != nil {
		t.Fatalf("Fail to create an integration test database: \n%s", err)
	} else {
		test(t, testDb)
		cleanupFn()
	}
}
