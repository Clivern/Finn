// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceBackendLock is a single row in the workspace_backend_lock table.
type WorkspaceBackendLock struct {
	ID        ID
	BackendID ID
	LockID    string // Client lock ID (e.g. Terraform's UUID) for unlock
	Operation string
	Info      string
	Who       string
	Version   string
	Path      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WorkspaceBackendLockRepository defines workspace backend lock CRUD.
type WorkspaceBackendLockRepository interface {
	Create(l *WorkspaceBackendLock) error
	GetByID(id ID) (*WorkspaceBackendLock, error)
	GetAnyActiveByBackendID(backendID ID) (*WorkspaceBackendLock, error)
	GetActiveByBackendID(backendID ID, lockID string) (*WorkspaceBackendLock, error)
	HasActiveByBackendID(backendID ID) (bool, error)
	IsLocked(backendID ID) bool
	Update(l *WorkspaceBackendLock) error
	Delete(id ID) error
	ListByBackendID(backendID ID, limit, offset int) ([]*WorkspaceBackendLock, error)
	CountByBackendID(backendID ID) (int64, error)
	SetInactive(id ID) error
}

// NewWorkspaceBackendLockRepository returns the repository for the current driver.
func NewWorkspaceBackendLockRepository(db *sql.DB) WorkspaceBackendLockRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceBackendLockRepoPostgres{db: db}
	}
	return &workspaceBackendLockRepoSQLite{db: db}
}

// scanIsActive reads is_active from DB (SQLite: int 0/1, Postgres: bool).
func scanIsActive(raw interface{}) bool {
	if raw == nil {
		return false
	}
	switch v := raw.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case int:
		return v != 0
	default:
		return false
	}
}

func scanWorkspaceBackendLockRows(rows *sql.Rows) ([]*WorkspaceBackendLock, error) {
	var list []*WorkspaceBackendLock
	for rows.Next() {
		l := &WorkspaceBackendLock{}
		var isActiveRaw interface{}
		if err := rows.Scan(
			&l.ID,
			&l.BackendID,
			&l.LockID,
			&l.Operation,
			&l.Info,
			&l.Who,
			&l.Version,
			&l.Path,
			&isActiveRaw,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.IsActive = scanIsActive(isActiveRaw)
		list = append(list, l)
	}
	return list, rows.Err()
}

type workspaceBackendLockRepoSQLite struct {
	db *sql.DB
}

func isActiveVal(isActive bool) int64 {
	if isActive {
		return 1
	}
	return 0
}

func (r *workspaceBackendLockRepoSQLite) Create(l *WorkspaceBackendLock) error {
	result, err := r.db.Exec(
		`INSERT INTO workspace_backend_lock (backend_id, lock_id, operation, info, who, version, path, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		l.BackendID.String(),
		l.LockID,
		l.Operation,
		l.Info,
		l.Who,
		l.Version,
		l.Path,
		isActiveVal(l.IsActive),
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	l.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT created_at, updated_at FROM workspace_backend_lock WHERE id = ?`,
		l.ID.String(),
	)
	return row.Scan(
		&l.CreatedAt,
		&l.UpdatedAt,
	)
}

func (r *workspaceBackendLockRepoSQLite) GetByID(id ID) (*WorkspaceBackendLock, error) {
	l := &WorkspaceBackendLock{}
	var isActiveRaw interface{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE id = ?`,
		id.String(),
	).Scan(
		&l.ID,
		&l.BackendID,
		&l.LockID,
		&l.Operation,
		&l.Info,
		&l.Who,
		&l.Version,
		&l.Path,
		&isActiveRaw,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.IsActive = scanIsActive(isActiveRaw)
	return l, nil
}

func (r *workspaceBackendLockRepoSQLite) GetAnyActiveByBackendID(backendID ID) (*WorkspaceBackendLock, error) {
	l := &WorkspaceBackendLock{}
	var isActiveRaw interface{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE backend_id = ? AND is_active = 1 LIMIT 1`,
		backendID.String(),
	).Scan(
		&l.ID,
		&l.BackendID,
		&l.LockID,
		&l.Operation,
		&l.Info,
		&l.Who,
		&l.Version,
		&l.Path,
		&isActiveRaw,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.IsActive = true
	return l, nil
}

func (r *workspaceBackendLockRepoSQLite) HasActiveByBackendID(backendID ID) (bool, error) {
	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM workspace_backend_lock WHERE backend_id = ? AND is_active = 1 LIMIT 1`,
		backendID.String(),
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *workspaceBackendLockRepoSQLite) IsLocked(backendID ID) bool {
	ok, _ := r.HasActiveByBackendID(backendID)
	return ok
}

func (r *workspaceBackendLockRepoSQLite) GetActiveByBackendID(backendID ID, lockID string) (*WorkspaceBackendLock, error) {
	l := &WorkspaceBackendLock{}
	var isActiveRaw interface{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE backend_id = ? AND lock_id = ? AND is_active = 1 LIMIT 1`,
		backendID.String(),
		lockID,
	).Scan(
		&l.ID,
		&l.BackendID,
		&l.LockID,
		&l.Operation,
		&l.Info,
		&l.Who,
		&l.Version,
		&l.Path,
		&isActiveRaw,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.IsActive = true
	return l, nil
}

func (r *workspaceBackendLockRepoSQLite) Update(l *WorkspaceBackendLock) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend_lock SET operation = ?, info = ?, who = ?, version = ?, path = ?, is_active = ?, updated_at = ?
		WHERE id = ?`,
		l.Operation,
		l.Info,
		l.Who,
		l.Version,
		l.Path,
		isActiveVal(l.IsActive),
		time.Now().UTC(),
		l.ID.String(),
	)
	return err
}

func (r *workspaceBackendLockRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_backend_lock WHERE id = ?`,
		id.String(),
	)
	return err
}

func (r *workspaceBackendLockRepoSQLite) SetInactive(id ID) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend_lock SET is_active = 0, updated_at = ? WHERE id = ?`,
		time.Now().UTC(),
		id.String(),
	)
	return err
}

func (r *workspaceBackendLockRepoSQLite) ListByBackendID(backendID ID, limit, offset int) ([]*WorkspaceBackendLock, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE backend_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		backendID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceBackendLockRows(rows)
}

func (r *workspaceBackendLockRepoSQLite) CountByBackendID(backendID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_backend_lock WHERE backend_id = ?`,
		backendID.String(),
	).Scan(&count)
	return count, err
}

type workspaceBackendLockRepoPostgres struct {
	db *sql.DB
}

func (r *workspaceBackendLockRepoPostgres) Create(l *WorkspaceBackendLock) error {
	return r.db.QueryRow(
		`INSERT INTO workspace_backend_lock (backend_id, lock_id, operation, info, who, version, path, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`,
		l.BackendID.String(),
		l.LockID,
		l.Operation,
		l.Info,
		l.Who,
		l.Version,
		l.Path,
		l.IsActive,
	).Scan(
		&l.ID,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
}

func (r *workspaceBackendLockRepoPostgres) GetByID(id ID) (*WorkspaceBackendLock, error) {
	l := &WorkspaceBackendLock{}
	var isActiveRaw interface{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE id = $1`,
		id.String(),
	).Scan(
		&l.ID,
		&l.BackendID,
		&l.LockID,
		&l.Operation,
		&l.Info,
		&l.Who,
		&l.Version,
		&l.Path,
		&isActiveRaw,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.IsActive = scanIsActive(isActiveRaw)
	return l, nil
}

func (r *workspaceBackendLockRepoPostgres) GetAnyActiveByBackendID(backendID ID) (*WorkspaceBackendLock, error) {
	l := &WorkspaceBackendLock{}
	var isActiveRaw interface{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE backend_id = $1 AND is_active = true LIMIT 1`,
		backendID.String(),
	).Scan(
		&l.ID,
		&l.BackendID,
		&l.LockID,
		&l.Operation,
		&l.Info,
		&l.Who,
		&l.Version,
		&l.Path,
		&isActiveRaw,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.IsActive = true
	return l, nil
}

func (r *workspaceBackendLockRepoPostgres) HasActiveByBackendID(backendID ID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM workspace_backend_lock WHERE backend_id = $1 AND is_active = true)`,
		backendID.String(),
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *workspaceBackendLockRepoPostgres) IsLocked(backendID ID) bool {
	ok, _ := r.HasActiveByBackendID(backendID)
	return ok
}

func (r *workspaceBackendLockRepoPostgres) GetActiveByBackendID(backendID ID, lockID string) (*WorkspaceBackendLock, error) {
	l := &WorkspaceBackendLock{}
	var isActiveRaw interface{}
	err := r.db.QueryRow(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE backend_id = $1 AND lock_id = $2 AND is_active = true LIMIT 1`,
		backendID.String(),
		lockID,
	).Scan(
		&l.ID,
		&l.BackendID,
		&l.LockID,
		&l.Operation,
		&l.Info,
		&l.Who,
		&l.Version,
		&l.Path,
		&isActiveRaw,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.IsActive = true
	return l, nil
}

func (r *workspaceBackendLockRepoPostgres) Update(l *WorkspaceBackendLock) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend_lock SET operation = $1, info = $2, who = $3, version = $4, path = $5, is_active = $6, updated_at = $7 WHERE id = $8`,
		l.Operation,
		l.Info,
		l.Who,
		l.Version,
		l.Path,
		l.IsActive,
		time.Now().UTC(),
		l.ID.String(),
	)
	return err
}

func (r *workspaceBackendLockRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_backend_lock WHERE id = $1`,
		id.String(),
	)
	return err
}

func (r *workspaceBackendLockRepoPostgres) SetInactive(id ID) error {
	_, err := r.db.Exec(
		`UPDATE workspace_backend_lock SET is_active = false, updated_at = $1 WHERE id = $2`,
		time.Now().UTC(),
		id.String(),
	)
	return err
}

func (r *workspaceBackendLockRepoPostgres) ListByBackendID(backendID ID, limit, offset int) ([]*WorkspaceBackendLock, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, backend_id, lock_id, operation, info, who, version, path, is_active, created_at, updated_at
		FROM workspace_backend_lock WHERE backend_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		backendID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceBackendLockRows(rows)
}

func (r *workspaceBackendLockRepoPostgres) CountByBackendID(backendID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_backend_lock WHERE backend_id = $1`,
		backendID.String(),
	).Scan(&count)
	return count, err
}
