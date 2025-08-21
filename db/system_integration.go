// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// SystemIntegration is the DB row for a system-wide integration (webhook, Slack, etc.).
type SystemIntegration struct {
	ID        ID
	Type      string
	Name      string
	Config    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SystemIntegrationRepository is the interface for system integration CRUD.
type SystemIntegrationRepository interface {
	Create(integration *SystemIntegration) error
	GetByID(id ID) (*SystemIntegration, error)
	Update(integration *SystemIntegration) error
	Delete(id ID) error
	List(limit, offset int) ([]*SystemIntegration, error)
	Count() (int64, error)
}

// NewSystemIntegrationRepository returns the repository for the current driver
func NewSystemIntegrationRepository(db *sql.DB) SystemIntegrationRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &systemIntegrationRepoPostgres{db: db}
	}
	return &systemIntegrationRepoSQLite{db: db}
}

// scanSystemIntegrationRows builds a slice of SystemIntegration from query rows
func scanSystemIntegrationRows(rows *sql.Rows) ([]*SystemIntegration, error) {
	var list []*SystemIntegration
	for rows.Next() {
		inv := &SystemIntegration{}
		if err := rows.Scan(
			&inv.ID,
			&inv.Type,
			&inv.Name,
			&inv.Config,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

// --- SQLite ---

// systemIntegrationRepoSQLite implements SystemIntegrationRepository for SQLite
type systemIntegrationRepoSQLite struct {
	db *sql.DB
}

// Create inserts a row
func (r *systemIntegrationRepoSQLite) Create(integration *SystemIntegration) error {
	result, err := r.db.Exec(
		`INSERT INTO system_integrations (type, name, config)
		VALUES (?, ?, ?)`,
		integration.Type,
		integration.Name,
		integration.Config,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	integration.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT created_at, updated_at
		FROM system_integrations WHERE id = ?`,
		integration.ID.String(),
	)
	_ = row.Scan(&integration.CreatedAt, &integration.UpdatedAt)
	return nil
}

// GetByID returns a system integration by ID
func (r *systemIntegrationRepoSQLite) GetByID(id ID) (*SystemIntegration, error) {
	inv := &SystemIntegration{}
	err := r.db.QueryRow(
		`SELECT id, type, name, config, created_at, updated_at
		FROM system_integrations WHERE id = ?`,
		id.String(),
	).Scan(
		&inv.ID,
		&inv.Type,
		&inv.Name,
		&inv.Config,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// Update updates a system integration
func (r *systemIntegrationRepoSQLite) Update(integration *SystemIntegration) error {
	_, err := r.db.Exec(
		`UPDATE system_integrations SET
			type = ?, name = ?, config = ?, updated_at = ?
		WHERE id = ?`,
		integration.Type,
		integration.Name,
		integration.Config,
		time.Now().UTC(),
		integration.ID.String(),
	)
	return err
}

// Delete removes a system integration
func (r *systemIntegrationRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM system_integrations WHERE id = ?`,
		id.String(),
	)
	return err
}

// List returns a list of system integrations
func (r *systemIntegrationRepoSQLite) List(limit, offset int) ([]*SystemIntegration, error) {
	rows, err := r.db.Query(
		`SELECT id, type, name, config, created_at, updated_at
		FROM system_integrations
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemIntegrationRows(rows)
}

// Count returns the total number of system integrations
func (r *systemIntegrationRepoSQLite) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM system_integrations`,
	).Scan(&count)
	return count, err
}

// --- Postgres ---

// systemIntegrationRepoPostgres implements SystemIntegrationRepository for PostgreSQL
type systemIntegrationRepoPostgres struct {
	db *sql.DB
}

// Create inserts a row
func (r *systemIntegrationRepoPostgres) Create(integration *SystemIntegration) error {
	err := r.db.QueryRow(
		`INSERT INTO system_integrations (type, name, config)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`,
		integration.Type,
		integration.Name,
		integration.Config,
	).Scan(&integration.ID, &integration.CreatedAt, &integration.UpdatedAt)
	return err
}

// GetByID returns a system integration by ID
func (r *systemIntegrationRepoPostgres) GetByID(id ID) (*SystemIntegration, error) {
	inv := &SystemIntegration{}
	err := r.db.QueryRow(
		`SELECT id, type, name, config, created_at, updated_at
		FROM system_integrations WHERE id = $1`,
		id.String(),
	).Scan(
		&inv.ID,
		&inv.Type,
		&inv.Name,
		&inv.Config,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// Update updates a system integration
func (r *systemIntegrationRepoPostgres) Update(integration *SystemIntegration) error {
	_, err := r.db.Exec(
		`UPDATE system_integrations SET
			type = $1, name = $2, config = $3, updated_at = $4
		WHERE id = $5`,
		integration.Type,
		integration.Name,
		integration.Config,
		time.Now().UTC(),
		integration.ID.String(),
	)
	return err
}

// Delete removes a system integration
func (r *systemIntegrationRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM system_integrations WHERE id = $1`,
		id.String(),
	)
	return err
}

// List returns a list of system integrations
func (r *systemIntegrationRepoPostgres) List(limit, offset int) ([]*SystemIntegration, error) {
	rows, err := r.db.Query(
		`SELECT id, type, name, config, created_at, updated_at
		FROM system_integrations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemIntegrationRows(rows)
}

// Count returns the total number of system integrations
func (r *systemIntegrationRepoPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM system_integrations`,
	).Scan(&count)
	return count, err
}
