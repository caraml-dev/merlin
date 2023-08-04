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

package database

import (
	"errors"
	"fmt"

	// required for gomigrate
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caraml-dev/merlin/config"
)

// InitDB initialises a database connection as well as runs the migration scripts.
// It is important to close the database after using it by calling defer db.Close()
func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Migrate
	err := migrateDB(cfg)
	if err != nil {
		return nil, err
	}

	// Init db
	db, err := gorm.Open(
		pg.Open(connectionString(cfg)),
		&gorm.Config{
			FullSaveAssociations: true,
			Logger:               logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to start Gorm DB: %w", err)
	}

	// Get the underlying SQL DB and apply connection properties
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

func connectionString(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Database,
		cfg.Password,
	)
}

// Migrate migrates the database, returns the Migrate object.
func migrateDB(cfg *config.DatabaseConfig) error {
	// run db migrations
	m, err := migrate.New(
		fmt.Sprintf(cfg.MigrationPath),
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		),
	)
	if err != nil {
		return fmt.Errorf("Failed to open migration path: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("Failed to run migrations: %w", err)
	}
	if sourceErr, dbErr := m.Close(); sourceErr != nil {
		return fmt.Errorf("Failed to close source after migration")
	} else if dbErr != nil {
		return fmt.Errorf("Failed to close database after migration")
	}

	return nil
}
