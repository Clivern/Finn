// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// Integration is the DB row for a workspace integration (webhook, Slack, etc.).
type Integration struct {
	ID          ID
	WorkspaceID ID
	Type        string
	Name        string
	Config      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IntegrationRepository is the interface for integration CRUD.
type IntegrationRepository interface {
	Create(integration *Integration) error
	GetByID(id ID) (*Integration, error)
	Update(integration *Integration) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*Integration, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewIntegrationRepository returns the repository for the current driver
func NewIntegrationRepository(db *sql.DB) IntegrationRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &integrationRepoPostgres{db: db}
	}
	return &integrationRepoSQLite{db: db}
}

// scanIntegrationRows builds a slice of Integration from query rows
func scanIntegrationRows(rows *sql.Rows) ([]*Integration, error) {
	var list []*Integration
	for rows.Next() {
		inv := &Integration{}
		if err := rows.Scan(
			&inv.ID,
			&inv.WorkspaceID,
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

// integrationRepoSQLite implements IntegrationRepository for SQLite
type integrationRepoSQLite struct {
	db *sql.DB
}

// Create inserts a row
func (r *integrationRepoSQLite) Create(integration *Integration) error {
	result, err := r.db.Exec(
		`INSERT INTO integrations (workspace_id, type, name, config)
		VALUES (?, ?, ?, ?)`,
		integration.WorkspaceID.String(),
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
		FROM integrations WHERE id = ?`,
		integration.ID.String(),
	)
	_ = row.Scan(&integration.CreatedAt, &integration.UpdatedAt)
	return nil
}

// GetByID returns a integration by ID
func (r *integrationRepoSQLite) GetByID(id ID) (*Integration, error) {
	inv := &Integration{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, type, name, config, created_at, updated_at
		FROM integrations WHERE id = ?`,
		id.String(),
	).Scan(
		&inv.ID,
		&inv.WorkspaceID,
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

// Update updates a integration
func (r *integrationRepoSQLite) Update(integration *Integration) error {
	_, err := r.db.Exec(
		`UPDATE integrations SET
			workspace_id = ?, type = ?, name = ?, config = ?, updated_at = ?
		WHERE id = ?`,
		integration.WorkspaceID.String(),
		integration.Type,
		integration.Name,
		integration.Config,
		time.Now().UTC(),
		integration.ID.String(),
	)
	return err
}

// Delete removes a integration
func (r *integrationRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM integrations WHERE id = ?`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID returns a list of integrations by workspace ID
func (r *integrationRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*Integration, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, type, name, config, created_at, updated_at
		FROM integrations
		WHERE workspace_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntegrationRows(rows)
}

// CountByWorkspaceID returns the total number of integrations by workspace ID
func (r *integrationRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM integrations
		WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

// --- Postgres ---

// integrationRepoPostgres implements IntegrationRepository for PostgreSQL
type integrationRepoPostgres struct {
	db *sql.DB
}

// Create inserts a row
func (r *integrationRepoPostgres) Create(integration *Integration) error {
	err := r.db.QueryRow(
		`INSERT INTO integrations (workspace_id, type, name, config)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		integration.WorkspaceID.String(),
		integration.Type,
		integration.Name,
		integration.Config,
	).Scan(&integration.ID, &integration.CreatedAt, &integration.UpdatedAt)
	return err
}

// GetByID returns a integration by ID
func (r *integrationRepoPostgres) GetByID(id ID) (*Integration, error) {
	inv := &Integration{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, type, name, config, created_at, updated_at
		FROM integrations WHERE id = $1`,
		id.String(),
	).Scan(
		&inv.ID,
		&inv.WorkspaceID,
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

// Update updates a integration
func (r *integrationRepoPostgres) Update(integration *Integration) error {
	_, err := r.db.Exec(
		`UPDATE integrations SET
			workspace_id = $1, type = $2, name = $3, config = $4, updated_at = $5
		WHERE id = $6`,
		integration.WorkspaceID.String(),
		integration.Type,
		integration.Name,
		integration.Config,
		time.Now().UTC(),
		integration.ID.String(),
	)
	return err
}

// Delete removes a integration
func (r *integrationRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM integrations WHERE id = $1`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID returns a list of integrations by workspace ID
func (r *integrationRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*Integration, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, type, name, config, created_at, updated_at
		FROM integrations
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntegrationRows(rows)
}

// CountByWorkspaceID returns the total number of integrations by workspace ID
func (r *integrationRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM integrations
		WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
