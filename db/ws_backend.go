// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceBackend is a single row in the workspace_backend table.
type WorkspaceBackend struct {
	ID          ID
	WorkspaceID ID
	Name        string
	Slug        string
	Username    string
	Password    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceBackendRepository defines workspace backend CRUD.
type WorkspaceBackendRepository interface {
	Create(r *WorkspaceBackend) error
	GetByID(id ID) (*WorkspaceBackend, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceBackend, error)
	GetBySlug(slug string) (*WorkspaceBackend, error)
	Update(r *WorkspaceBackend) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceBackend, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceBackendRepository returns the repository for the current driver.
func NewWorkspaceBackendRepository(db *sql.DB) WorkspaceBackendRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceBackendRepoPostgres{db: db}
	}
	return &workspaceBackendRepoSQLite{db: db}
}

func scanWorkspaceBackendRows(rows *sql.Rows) ([]*WorkspaceBackend, error) {
	var list []*WorkspaceBackend
	for rows.Next() {
		b := &WorkspaceBackend{}
		if err := rows.Scan(
			&b.ID,
			&b.WorkspaceID,
			&b.Name,
			&b.Slug,
			&b.Username,
			&b.Password,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

type workspaceBackendRepoSQLite struct {
	db *sql.DB
}

func (r *workspaceBackendRepoSQLite) Create(b *WorkspaceBackend) error {
	result, err := r.db.Exec(
		`INSERT INTO workspace_backend (workspace_id, name, slug, username, password)
		VALUES (?, ?, ?, ?, ?)`,
		b.WorkspaceID.String(),
		b.Name,
		b.Slug,
		b.Username,
		b.Password,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT created_at, updated_at FROM workspace_backend WHERE id = ?`,
		b.ID.String(),
	)
	return row.Scan(
		&b.CreatedAt,
		&b.UpdatedAt,
	)
}

func (r *workspaceBackendRepoSQLite) GetByID(id ID) (*WorkspaceBackend, error) {
	b := &WorkspaceBackend{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, name, slug, username, password, created_at, updated_at
		FROM workspace_backend WHERE id = ?`,
		id.String(),
	).Scan(
		&b.ID,
		&b.WorkspaceID,
		&b.Name,
		&b.Slug,
		&b.Username,
		&b.Password,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *workspaceBackendRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceBackend, error) {
	b, err := r.GetByID(id)
	if err != nil || b == nil {
		return b, err
	}
	if b.WorkspaceID != workspaceID {
		return nil, nil
	}
	return b, nil
}

func (r *workspaceBackendRepoSQLite) GetBySlug(slug string) (*WorkspaceBackend, error) {
	b := &WorkspaceBackend{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, name, slug, username, password, created_at, updated_at
		FROM workspace_backend WHERE slug = ?`,
		slug,
	).Scan(
		&b.ID,
		&b.WorkspaceID,
		&b.Name,
		&b.Slug,
		&b.Username,
		&b.Password,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *workspaceBackendRepoSQLite) Update(b *WorkspaceBackend) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend SET name = ?, slug = ?, username = ?, password = ?, updated_at = ?
		WHERE id = ?`,
		b.Name,
		b.Slug,
		b.Username,
		b.Password,
		time.Now().UTC(),
		b.ID.String(),
	)
	return err
}

func (r *workspaceBackendRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(`DELETE FROM workspace_backend WHERE id = ?`, id.String())
	return err
}

func (r *workspaceBackendRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceBackend, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, workspace_id, name, slug, username, password, created_at, updated_at
		FROM workspace_backend WHERE workspace_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceBackendRows(rows)
}

func (r *workspaceBackendRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_backend WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

type workspaceBackendRepoPostgres struct {
	db *sql.DB
}

func (r *workspaceBackendRepoPostgres) Create(b *WorkspaceBackend) error {
	return r.db.QueryRow(
		`INSERT INTO workspace_backend (workspace_id, name, slug, username, password)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		b.WorkspaceID.String(),
		b.Name,
		b.Slug,
		b.Username,
		b.Password,
	).Scan(
		&b.ID,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
}

func (r *workspaceBackendRepoPostgres) GetByID(id ID) (*WorkspaceBackend, error) {
	b := &WorkspaceBackend{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, name, slug, username, password, created_at, updated_at
		FROM workspace_backend WHERE id = $1`,
		id.String(),
	).Scan(
		&b.ID,
		&b.WorkspaceID,
		&b.Name,
		&b.Slug,
		&b.Username,
		&b.Password,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *workspaceBackendRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceBackend, error) {
	b, err := r.GetByID(id)
	if err != nil || b == nil {
		return b, err
	}
	if b.WorkspaceID != workspaceID {
		return nil, nil
	}
	return b, nil
}

func (r *workspaceBackendRepoPostgres) GetBySlug(slug string) (*WorkspaceBackend, error) {
	b := &WorkspaceBackend{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, name, slug, username, password, created_at, updated_at
		FROM workspace_backend WHERE slug = $1`,
		slug,
	).Scan(
		&b.ID,
		&b.WorkspaceID,
		&b.Name,
		&b.Slug,
		&b.Username,
		&b.Password,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *workspaceBackendRepoPostgres) Update(b *WorkspaceBackend) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend SET name = $1, slug = $2, username = $3, password = $4, updated_at = $5 WHERE id = $6`,
		b.Name,
		b.Slug,
		b.Username,
		b.Password,
		time.Now().UTC(),
		b.ID.String(),
	)
	return err
}

func (r *workspaceBackendRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(`DELETE FROM workspace_backend WHERE id = $1`, id.String())
	return err
}

func (r *workspaceBackendRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceBackend, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, workspace_id, name, slug, username, password, created_at, updated_at
		FROM workspace_backend WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceBackendRows(rows)
}

func (r *workspaceBackendRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_backend WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
