// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceUser is a single row in the workspace_users table
type WorkspaceUser struct {
	ID          ID
	WorkspaceID ID
	UserID      ID
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceUserRepository defines workspace-user membership CRUD
type WorkspaceUserRepository interface {
	Create(m *WorkspaceUser) error
	GetByID(id ID) (*WorkspaceUser, error)
	GetByWorkspaceAndUser(workspaceID, userID ID) (*WorkspaceUser, error)
	Update(m *WorkspaceUser) error
	Delete(id ID) error
	List(limit, offset int) ([]*WorkspaceUser, error)
	Count() (int64, error)
	ListByWorkspace(workspaceID ID, limit, offset int) ([]*WorkspaceUser, error)
	CountByWorkspace(workspaceID ID) (int64, error)
}

// NewWorkspaceUserRepository returns the repository for the current driver
func NewWorkspaceUserRepository(db *sql.DB) WorkspaceUserRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceUserRepoPostgres{db: db}
	}
	return &workspaceUserRepoSQLite{db: db}
}

// scanWorkspaceUserRows builds a slice of WorkspaceUser from query rows
func scanWorkspaceUserRows(rows *sql.Rows) ([]*WorkspaceUser, error) {
	var list []*WorkspaceUser
	for rows.Next() {
		m := &WorkspaceUser{}
		if err := rows.Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// --- SQLite ---

// workspaceUserRepoSQLite implements WorkspaceUserRepository for SQLite
type workspaceUserRepoSQLite struct {
	db *sql.DB
}

// Create inserts a row
func (r *workspaceUserRepoSQLite) Create(m *WorkspaceUser) error {
	result, err := r.db.Exec(
		`INSERT INTO workspace_users (workspace_id, user_id, role)
		VALUES (?, ?, ?)`,
		m.WorkspaceID.String(), m.UserID.String(), m.Role,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	m.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT created_at, updated_at
		FROM workspace_users WHERE id = ?`,
		m.ID.String(),
	)
	_ = row.Scan(&m.CreatedAt, &m.UpdatedAt)
	return nil
}

// GetByID returns a workspace user by ID
func (r *workspaceUserRepoSQLite) GetByID(id ID) (*WorkspaceUser, error) {
	m := &WorkspaceUser{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users WHERE id = ?`,
		id.String(),
	).Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetByWorkspaceAndUser returns a workspace user by workspace and user ID
func (r *workspaceUserRepoSQLite) GetByWorkspaceAndUser(workspaceID, userID ID) (*WorkspaceUser, error) {
	m := &WorkspaceUser{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users WHERE workspace_id = ? AND user_id = ?`,
		workspaceID.String(), userID.String(),
	).Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Update updates a workspace user
func (r *workspaceUserRepoSQLite) Update(m *WorkspaceUser) error {
	_, err := r.db.Exec(
		`UPDATE workspace_users SET workspace_id = ?, user_id = ?, role = ?, updated_at = ?
		WHERE id = ?`,
		m.WorkspaceID.String(), m.UserID.String(), m.Role, time.Now().UTC(),
		m.ID.String(),
	)
	return err
}

// Delete removes a workspace user
func (r *workspaceUserRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_users WHERE id = ?`,
		id.String(),
	)
	return err
}

// List returns a list of workspace users
func (r *workspaceUserRepoSQLite) List(limit, offset int) ([]*WorkspaceUser, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceUserRows(rows)
}

// Count returns the total number of workspace users
func (r *workspaceUserRepoSQLite) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_users`,
	).Scan(&count)
	return count, err
}

// ListByWorkspace returns workspace users for the given workspace with pagination
func (r *workspaceUserRepoSQLite) ListByWorkspace(workspaceID ID, limit, offset int) ([]*WorkspaceUser, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users
		WHERE workspace_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		workspaceID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceUserRows(rows)
}

// CountByWorkspace returns the number of users in the workspace
func (r *workspaceUserRepoSQLite) CountByWorkspace(workspaceID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_users WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

// --- Postgres ---

// workspaceUserRepoPostgres implements WorkspaceUserRepository for PostgreSQL
type workspaceUserRepoPostgres struct {
	db *sql.DB
}

// Create inserts a row
func (r *workspaceUserRepoPostgres) Create(m *WorkspaceUser) error {
	err := r.db.QueryRow(
		`INSERT INTO workspace_users (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`,
		m.WorkspaceID.String(), m.UserID.String(), m.Role,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	return err
}

// GetByID returns a workspace user by ID
func (r *workspaceUserRepoPostgres) GetByID(id ID) (*WorkspaceUser, error) {
	m := &WorkspaceUser{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users WHERE id = $1`,
		id.String(),
	).Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetByWorkspaceAndUser returns a workspace user by workspace and user ID
func (r *workspaceUserRepoPostgres) GetByWorkspaceAndUser(workspaceID, userID ID) (*WorkspaceUser, error) {
	m := &WorkspaceUser{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID.String(), userID.String(),
	).Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Update updates a workspace user
func (r *workspaceUserRepoPostgres) Update(m *WorkspaceUser) error {
	_, err := r.db.Exec(
		`UPDATE workspace_users SET workspace_id = $1, user_id = $2, role = $3, updated_at = $4
		WHERE id = $5`,
		m.WorkspaceID.String(), m.UserID.String(), m.Role, time.Now().UTC(),
		m.ID.String(),
	)
	return err
}

// Delete removes a workspace user
func (r *workspaceUserRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_users WHERE id = $1`,
		id.String(),
	)
	return err
}

// List returns a list of workspace users
func (r *workspaceUserRepoPostgres) List(limit, offset int) ([]*WorkspaceUser, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceUserRows(rows)
}

// Count returns the total number of workspace users
func (r *workspaceUserRepoPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_users`,
	).Scan(&count)
	return count, err
}

// ListByWorkspace returns workspace users for the given workspace with pagination
func (r *workspaceUserRepoPostgres) ListByWorkspace(workspaceID ID, limit, offset int) ([]*WorkspaceUser, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		workspaceID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceUserRows(rows)
}

// CountByWorkspace returns the number of users in the workspace
func (r *workspaceUserRepoPostgres) CountByWorkspace(workspaceID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_users WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
