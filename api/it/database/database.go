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

//go:build integration || integration_local
// +build integration integration_local

package database

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caraml-dev/merlin/log"
)

func connectionString(db string) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password='%s' sslmode=disable", host, 5432, user, db, password)
}

func create(conn *sql.DB, dbName string) (*sql.DB, error) {
	if _, err := conn.Exec("CREATE DATABASE " + dbName); err != nil {
		return nil, err
	} else if testDb, err := sql.Open("postgres", connectionString(dbName)); err != nil {
		if _, err := conn.Exec("DROP DATABASE " + dbName); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		}
		return nil, err
	} else {
		return testDb, nil
	}
}

func migrate(db *sql.DB, dbName string) (*sql.DB, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	defer driver.Close()

	if migrations, err := gomigrate.NewWithDatabaseInstance("file://../../db-migrations", dbName, driver); err != nil {
		return nil, err
	} else if err = migrations.Up(); err != nil {
		return nil, err
	}
	return sql.Open("postgres", connectionString(dbName))
}

// Connects to test postgreSQL instance (either local or the one at CI environment)
// and creates a new database with an up-to-date schema
func CreateTestDatabase() (*gorm.DB, func(), error) {
	testDbName := fmt.Sprintf("mlp_id_%d", time.Now().UnixNano())

	connStr := connectionString(database)
	log.Infof("connecting to test db: %s", connStr)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	testDb, err := create(conn, testDbName)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := testDb.Close(); err != nil {
			log.Fatalf("Failed to close connection to integration test database: \n%s", err)
		} else if _, err := conn.Exec("DROP DATABASE " + testDbName); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		} else if err = conn.Close(); err != nil {
			log.Fatalf("Failed to close database: \n%s", err)
		}
	}

	if testDb, err = migrate(testDb, testDbName); err != nil {
		cleanup()
		return nil, nil, err
	} else if gormDb, err := gorm.Open(pg.Open(connStr),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}); err != nil {
		cleanup()
		return nil, nil, err
	} else {
		return gormDb, cleanup, nil
	}
}

func WithTestDatabase(t *testing.T, test func(t *testing.T, db *gorm.DB)) {
	if testDb, cleanupFn, err := CreateTestDatabase(); err != nil {
		t.Fatalf("Fail to create an integration test database: \n%s", err)
	} else {
		test(t, testDb)
		cleanupFn()
	}
}
