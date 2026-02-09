// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceRepositoryCloneTask is a single row in the workspace_repository_clone_tasks table.
type WorkspaceRepositoryCloneTask struct {
	ID           ID
	RepositoryID ID
	RunnerID     *ID
	Status       string
	RunnerTaskID string
	Result       *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// WorkspaceRepositoryCloneTaskRepository is the interface for clone task persistence per repo.
type WorkspaceRepositoryCloneTaskRepository interface {
	Create(t *WorkspaceRepositoryCloneTask) error
	GetByID(id ID) (*WorkspaceRepositoryCloneTask, error)
	Update(t *WorkspaceRepositoryCloneTask) error
	UpdateStatus(id ID, status string, result *string) error
	Delete(id ID) error
	ListByRepositoryID(repositoryID ID, limit, offset int) ([]*WorkspaceRepositoryCloneTask, error)
	CountByRepositoryID(repositoryID ID) (int64, error)
	LatestStatusByRepositoryID(repositoryID ID) (string, error)
}

// NewWorkspaceRepositoryCloneTaskRepository returns the repository for the current driver.
func NewWorkspaceRepositoryCloneTaskRepository(db *sql.DB) WorkspaceRepositoryCloneTaskRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceRepositoryCloneTaskRepoPostgres{db: db}
	}
	return &workspaceRepositoryCloneTaskRepoSQLite{db: db}
}

func scanWorkspaceRepositoryCloneTaskRows(rows *sql.Rows) ([]*WorkspaceRepositoryCloneTask, error) {
	var list []*WorkspaceRepositoryCloneTask
	for rows.Next() {
		t := &WorkspaceRepositoryCloneTask{}
		var runnerID interface{}
		var runnerTaskID, result sql.NullString
		err := rows.Scan(
			&t.ID,
			&t.RepositoryID,
			&runnerID,
			&t.Status,
			&runnerTaskID,
			&result,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if runnerID != nil {
			idVal := ID("")
			_ = idVal.Scan(runnerID)
			t.RunnerID = &idVal
		}
		if runnerTaskID.Valid {
			t.RunnerTaskID = runnerTaskID.String
		}
		if result.Valid {
			t.Result = &result.String
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// --- SQLite ---

type workspaceRepositoryCloneTaskRepoSQLite struct {
	db *sql.DB
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) Create(t *WorkspaceRepositoryCloneTask) error {
	result, err := r.db.Exec(
		`INSERT INTO workspace_repository_clone_tasks
		(repository_id, runner_id, status, runner_task_id, result)
		VALUES (?, ?, ?, ?, ?)`,
		t.RepositoryID.String(),
		t.RunnerID.ValueOrNilPtr(),
		t.Status,
		nullStr(t.RunnerTaskID),
		nullStrPtr(t.Result),
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT
			created_at,
			updated_at
		FROM workspace_repository_clone_tasks
		WHERE id = ?`,
		t.ID.String(),
	)
	return row.Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) GetByID(id ID) (*WorkspaceRepositoryCloneTask, error) {
	row := r.db.QueryRow(
		`SELECT
			id, repository_id, runner_id, status, runner_task_id,
			result, created_at, updated_at
		FROM workspace_repository_clone_tasks
		WHERE id = ?`,
		id.String(),
	)
	t := &WorkspaceRepositoryCloneTask{}
	var runnerID interface{}
	var result sql.NullString
	var runnerTaskID sql.NullString
	err := row.Scan(
		&t.ID,
		&t.RepositoryID,
		&runnerID,
		&t.Status,
		&runnerTaskID,
		&result,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if runnerID != nil {
		idVal := ID("")
		_ = idVal.Scan(runnerID)
		t.RunnerID = &idVal
	}
	if runnerTaskID.Valid {
		t.RunnerTaskID = runnerTaskID.String
	}
	if result.Valid {
		t.Result = &result.String
	}
	return t, nil
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) Update(t *WorkspaceRepositoryCloneTask) error {
	_, err := r.db.Exec(
		`UPDATE workspace_repository_clone_tasks
		SET
			runner_id = ?,
			status = ?,
			runner_task_id = ?,
			result = ?,
			updated_at = ?
		WHERE id = ?`,
		t.RunnerID.ValueOrNilPtr(),
		t.Status,
		nullStr(t.RunnerTaskID),
		nullStrPtr(t.Result),
		time.Now().UTC(),
		t.ID.String(),
	)
	return err
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) UpdateStatus(id ID, status string, result *string) error {
	_, err := r.db.Exec(
		`UPDATE workspace_repository_clone_tasks
		SET
			status = ?,
			result = ?,
			updated_at = ?
		WHERE id = ?`,
		status,
		nullStrPtr(result),
		time.Now().UTC(),
		id.String(),
	)
	return err
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(`DELETE FROM workspace_repository_clone_tasks WHERE id = ?`, id.String())
	return err
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) ListByRepositoryID(repositoryID ID, limit, offset int) ([]*WorkspaceRepositoryCloneTask, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT
			id, repository_id, runner_id, status, runner_task_id,
			result, created_at, updated_at
		FROM workspace_repository_clone_tasks
		WHERE repository_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		repositoryID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceRepositoryCloneTaskRows(rows)
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) CountByRepositoryID(repositoryID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_repository_clone_tasks
		WHERE repository_id = ?`,
		repositoryID.String(),
	).Scan(&count)
	return count, err
}

func (r *workspaceRepositoryCloneTaskRepoSQLite) LatestStatusByRepositoryID(repositoryID ID) (string, error) {
	var status string
	err := r.db.QueryRow(
		`SELECT status
		FROM workspace_repository_clone_tasks
		WHERE repository_id = ?
		ORDER BY created_at DESC
		LIMIT 1`,
		repositoryID.String(),
	).Scan(&status)
	if err == sql.ErrNoRows {
		return "unknown", nil
	}
	if err != nil {
		return "", err
	}
	return status, nil
}

type workspaceRepositoryCloneTaskRepoPostgres struct {
	db *sql.DB
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) Create(t *WorkspaceRepositoryCloneTask) error {
	return r.db.QueryRow(
		`INSERT INTO workspace_repository_clone_tasks
		(repository_id, runner_id, status, runner_task_id, result)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		t.RepositoryID.String(),
		t.RunnerID.ValueOrNilPtr(),
		t.Status,
		nullStr(t.RunnerTaskID),
		nullStrPtr(t.Result),
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) GetByID(id ID) (*WorkspaceRepositoryCloneTask, error) {
	row := r.db.QueryRow(
		`SELECT
			id, repository_id, runner_id, status, runner_task_id,
			result, created_at, updated_at
		FROM workspace_repository_clone_tasks
		WHERE id = $1`,
		id.String(),
	)
	t := &WorkspaceRepositoryCloneTask{}
	var runnerID interface{}
	var result sql.NullString
	var runnerTaskID sql.NullString
	err := row.Scan(
		&t.ID,
		&t.RepositoryID,
		&runnerID,
		&t.Status,
		&runnerTaskID,
		&result,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if runnerID != nil {
		idVal := ID("")
		_ = idVal.Scan(runnerID)
		t.RunnerID = &idVal
	}
	if runnerTaskID.Valid {
		t.RunnerTaskID = runnerTaskID.String
	}
	if result.Valid {
		t.Result = &result.String
	}
	return t, nil
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) Update(t *WorkspaceRepositoryCloneTask) error {
	_, err := r.db.Exec(
		`UPDATE workspace_repository_clone_tasks
		SET
			runner_id = $1,
			status = $2,
			runner_task_id = $3,
			result = $4,
			updated_at = $5
		WHERE id = $6`,
		t.RunnerID.ValueOrNilPtr(),
		t.Status,
		nullStr(t.RunnerTaskID),
		nullStrPtr(t.Result),
		time.Now().UTC(),
		t.ID.String(),
	)
	return err
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) UpdateStatus(id ID, status string, result *string) error {
	_, err := r.db.Exec(
		`UPDATE workspace_repository_clone_tasks
		SET
			status = $1,
			result = $2,
			updated_at = $3
		WHERE id = $4`,
		status,
		nullStrPtr(result),
		time.Now().UTC(),
		id.String(),
	)
	return err
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(`DELETE FROM workspace_repository_clone_tasks WHERE id = $1`, id.String())
	return err
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) ListByRepositoryID(repositoryID ID, limit, offset int) ([]*WorkspaceRepositoryCloneTask, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT
			id, repository_id, runner_id, status, runner_task_id,
			result, created_at, updated_at
		FROM workspace_repository_clone_tasks
		WHERE repository_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		repositoryID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceRepositoryCloneTaskRows(rows)
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) CountByRepositoryID(repositoryID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_repository_clone_tasks
		WHERE repository_id = $1`,
		repositoryID.String(),
	).Scan(&count)
	return count, err
}

func (r *workspaceRepositoryCloneTaskRepoPostgres) LatestStatusByRepositoryID(repositoryID ID) (string, error) {
	var status string
	err := r.db.QueryRow(
		`SELECT status
		FROM workspace_repository_clone_tasks
		WHERE repository_id = $1
		ORDER BY created_at DESC
		LIMIT 1`,
		repositoryID.String(),
	).Scan(&status)
	if err == sql.ErrNoRows {
		return "unknown", nil
	}
	if err != nil {
		return "", err
	}
	return status, nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullStrPtr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
