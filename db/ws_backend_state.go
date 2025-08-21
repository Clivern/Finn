// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceBackendState is a single row in the workspace_backend_state table.
type WorkspaceBackendState struct {
	ID        ID
	BackendID ID
	Name      string
	Value     string
	Checksum  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WorkspaceBackendStateRepository is the interface for workspace backend state CRUD.
type WorkspaceBackendStateRepository interface {
	Create(s *WorkspaceBackendState) error
	GetByID(id ID) (*WorkspaceBackendState, error)
	GetByBackendIDAndName(backendID ID, name string) (*WorkspaceBackendState, error)
	GetLatestByBackendID(backendID ID) (*WorkspaceBackendState, error)
	Update(s *WorkspaceBackendState) error
	Delete(id ID) error
	ListByBackendID(backendID ID, limit, offset int) ([]*WorkspaceBackendState, error)
	CountByBackendID(backendID ID) (int64, error)
}

// NewWorkspaceBackendStateRepository returns the repository for the current driver.
func NewWorkspaceBackendStateRepository(db *sql.DB) WorkspaceBackendStateRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceBackendStateRepoPostgres{db: db}
	}
	return &workspaceBackendStateRepoSQLite{db: db}
}

func scanWorkspaceBackendStateRows(rows *sql.Rows) ([]*WorkspaceBackendState, error) {
	var list []*WorkspaceBackendState
	for rows.Next() {
		s := &WorkspaceBackendState{}
		if err := rows.Scan(
			&s.ID,
			&s.BackendID,
			&s.Name,
			&s.Value,
			&s.Checksum,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

type workspaceBackendStateRepoSQLite struct {
	db *sql.DB
}

func (r *workspaceBackendStateRepoSQLite) Create(s *WorkspaceBackendState) error {
	result, err := r.db.Exec(
		`INSERT INTO workspace_backend_state (backend_id, name, value, checksum)
		VALUES (?, ?, ?, ?)`,
		s.BackendID.String(),
		s.Name,
		s.Value,
		s.Checksum,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT created_at, updated_at FROM workspace_backend_state WHERE id = ?`,
		s.ID.String(),
	)
	return row.Scan(
		&s.CreatedAt,
		&s.UpdatedAt,
	)
}

func (r *workspaceBackendStateRepoSQLite) GetByID(id ID) (*WorkspaceBackendState, error) {
	s := &WorkspaceBackendState{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE id = ?`,
		id.String(),
	).Scan(
		&s.ID,
		&s.BackendID,
		&s.Name,
		&s.Value,
		&s.Checksum,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *workspaceBackendStateRepoSQLite) GetByBackendIDAndName(backendID ID, name string) (*WorkspaceBackendState, error) {
	s := &WorkspaceBackendState{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE backend_id = ? AND name = ?`,
		backendID.String(),
		name,
	).Scan(
		&s.ID,
		&s.BackendID,
		&s.Name,
		&s.Value,
		&s.Checksum,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *workspaceBackendStateRepoSQLite) GetLatestByBackendID(backendID ID) (*WorkspaceBackendState, error) {
	s := &WorkspaceBackendState{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE backend_id = ? ORDER BY created_at DESC LIMIT 1`,
		backendID.String(),
	).Scan(
		&s.ID,
		&s.BackendID,
		&s.Name,
		&s.Value,
		&s.Checksum,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *workspaceBackendStateRepoSQLite) Update(s *WorkspaceBackendState) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend_state SET value = ?, checksum = ?, updated_at = ?
		WHERE id = ?`,
		s.Value,
		s.Checksum,
		time.Now().UTC(),
		s.ID.String(),
	)
	return err
}

func (r *workspaceBackendStateRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_backend_state WHERE id = ?`,
		id.String(),
	)
	return err
}

func (r *workspaceBackendStateRepoSQLite) ListByBackendID(backendID ID, limit, offset int) ([]*WorkspaceBackendState, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE backend_id = ? ORDER BY name LIMIT ? OFFSET ?`,
		backendID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceBackendStateRows(rows)
}

func (r *workspaceBackendStateRepoSQLite) CountByBackendID(backendID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_backend_state WHERE backend_id = ?`,
		backendID.String(),
	).Scan(&count)
	return count, err
}

type workspaceBackendStateRepoPostgres struct {
	db *sql.DB
}

func (r *workspaceBackendStateRepoPostgres) Create(s *WorkspaceBackendState) error {
	return r.db.QueryRow(
		`INSERT INTO workspace_backend_state (backend_id, name, value, checksum)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`,
		s.BackendID.String(),
		s.Name,
		s.Value,
		s.Checksum,
	).Scan(
		&s.ID,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
}

func (r *workspaceBackendStateRepoPostgres) GetByID(id ID) (*WorkspaceBackendState, error) {
	s := &WorkspaceBackendState{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE id = $1`,
		id.String(),
	).Scan(
		&s.ID,
		&s.BackendID,
		&s.Name,
		&s.Value,
		&s.Checksum,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *workspaceBackendStateRepoPostgres) GetByBackendIDAndName(backendID ID, name string) (*WorkspaceBackendState, error) {
	s := &WorkspaceBackendState{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE backend_id = $1 AND name = $2`,
		backendID.String(),
		name,
	).Scan(
		&s.ID,
		&s.BackendID,
		&s.Name,
		&s.Value,
		&s.Checksum,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *workspaceBackendStateRepoPostgres) GetLatestByBackendID(backendID ID) (*WorkspaceBackendState, error) {
	s := &WorkspaceBackendState{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE backend_id = $1 ORDER BY created_at DESC LIMIT 1`,
		backendID.String(),
	).Scan(
		&s.ID,
		&s.BackendID,
		&s.Name,
		&s.Value,
		&s.Checksum,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *workspaceBackendStateRepoPostgres) Update(s *WorkspaceBackendState) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend_state SET value = $1, checksum = $2, updated_at = $3 WHERE id = $4`,
		s.Value,
		s.Checksum,
		time.Now().UTC(),
		s.ID.String(),
	)
	return err
}

func (r *workspaceBackendStateRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_backend_state WHERE id = $1`,
		id.String(),
	)
	return err
}

func (r *workspaceBackendStateRepoPostgres) ListByBackendID(backendID ID, limit, offset int) ([]*WorkspaceBackendState, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, backend_id, name, value, checksum, created_at, updated_at
		FROM workspace_backend_state WHERE backend_id = $1 ORDER BY name LIMIT $2 OFFSET $3`,
		backendID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceBackendStateRows(rows)
}

func (r *workspaceBackendStateRepoPostgres) CountByBackendID(backendID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_backend_state WHERE backend_id = $1`,
		backendID.String(),
	).Scan(&count)
	return count, err
}
