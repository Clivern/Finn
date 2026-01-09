// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	// dbConn holds the singleton database connection
	dbConn *Connection
	// dbDriver is the driver name from config
	dbDriver string
	// mu protects dbConn during initialization
	mu sync.RWMutex
)

// InitDB initializes the global database connection
func InitDB(config Config) error {
	mu.Lock()
	defer mu.Unlock()

	if dbConn != nil {
		log.Warn().Msg("Database connection already initialized")
		return nil
	}

	conn, err := NewConnection(config)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	dbConn = conn
	dbDriver = config.Driver
	log.Info().Msg("Global database connection initialized")
	return nil
}

// GetDB returns the global database connection
func GetDB() *sql.DB {
	mu.RLock()
	defer mu.RUnlock()

	if dbConn == nil {
		log.Error().Msg("Database not initialized")
		return nil
	}

	return dbConn.DB
}

// GetDriver returns the database driver name from config
func GetDriver() string {
	return dbDriver
}

// CloseDB closes the global database connection
func CloseDB() error {
	mu.Lock()
	defer mu.Unlock()

	if dbConn == nil {
		return nil
	}

	err := dbConn.Close()
	dbConn = nil
	dbDriver = ""
	return err
}
