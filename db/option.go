// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Option is a key-value setting in the DB.
type Option struct {
	ID        ID
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OptionRepository is the interface for option CRUD.
type OptionRepository interface {
	Create(key, value string) error
	Get(key string) (*Option, error)
	Update(key, value string) error
	Delete(key string) error
	List() ([]*Option, error)
}

// NewOptionRepository returns the repository for the current driver
func NewOptionRepository(db *sql.DB) OptionRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &optionRepoPostgres{db: db}
	}
	return &optionRepoSQLite{db: db}
}

// scanOptionRows builds a slice of Option from query rows
func scanOptionRows(rows *sql.Rows) ([]*Option, error) {
	var list []*Option
	for rows.Next() {
		o := &Option{}
		if err := rows.Scan(&o.ID, &o.Key, &o.Value, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

// --- SQLite ---

// optionRepoSQLite implements OptionRepository for SQLite
type optionRepoSQLite struct {
	db *sql.DB
}

// Create inserts a row
func (r *optionRepoSQLite) Create(key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO options (key, value) VALUES (?, ?)`,
		key, value,
	)
	return err
}

// Get returns an option by key
func (r *optionRepoSQLite) Get(key string) (*Option, error) {
	o := &Option{}
	err := r.db.QueryRow(
		`SELECT id, key, value, created_at, updated_at
		FROM options WHERE key = ?`,
		key,
	).Scan(&o.ID, &o.Key, &o.Value, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Update updates an option
func (r *optionRepoSQLite) Update(key, value string) error {
	_, err := r.db.Exec(
		`UPDATE options SET value = ?, updated_at = ? WHERE key = ?`,
		value, time.Now().UTC(), key,
	)
	return err
}

// Delete removes an option
func (r *optionRepoSQLite) Delete(key string) error {
	_, err := r.db.Exec(
		`DELETE FROM options WHERE key = ?`,
		key,
	)
	return err
}

// List returns a list of options
func (r *optionRepoSQLite) List() ([]*Option, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value, created_at, updated_at
		FROM options ORDER BY key`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOptionRows(rows)
}

// --- Postgres ---

// optionRepoPostgres implements OptionRepository for PostgreSQL
type optionRepoPostgres struct {
	db *sql.DB
}

// Create inserts a row
func (r *optionRepoPostgres) Create(key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO options (key, value) VALUES ($1, $2)`,
		key, value,
	)
	return err
}

// Get returns an option by key
func (r *optionRepoPostgres) Get(key string) (*Option, error) {
	o := &Option{}
	err := r.db.QueryRow(
		`SELECT id, key, value, created_at, updated_at
		FROM options WHERE key = $1`,
		key,
	).Scan(&o.ID, &o.Key, &o.Value, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Update updates an option
func (r *optionRepoPostgres) Update(key, value string) error {
	_, err := r.db.Exec(
		`UPDATE options SET value = $1, updated_at = $2 WHERE key = $3`,
		value, time.Now().UTC(), key,
	)
	return err
}

// Delete removes an option
func (r *optionRepoPostgres) Delete(key string) error {
	_, err := r.db.Exec(
		`DELETE FROM options WHERE key = $1`,
		key,
	)
	return err
}

// List returns a list of options
func (r *optionRepoPostgres) List() ([]*Option, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value, created_at, updated_at
		FROM options ORDER BY key`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOptionRows(rows)
}
