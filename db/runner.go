// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// Runner status constants
const (
	RunnerStatusActive      = "active"
	RunnerStatusInactive    = "inactive"
	RunnerStatusMaintenance = "maintenance"
)

// Runner is the DB row for a runner.
type Runner struct {
	ID               ID
	Token            string
	Name             string
	Tags             string
	MaxParallelTasks int
	Status           string
	URL              string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// RunnerRepository is the interface for runner CRUD.
type RunnerRepository interface {
	Create(runner *Runner) error
	GetByID(id ID) (*Runner, error)
	GetByToken(token string) (*Runner, error)
	Update(runner *Runner) error
	Delete(id ID) error
	List(limit, offset int) ([]*Runner, error)
	ListActive() ([]*Runner, error)
	Count() (int64, error)
	CountActive() (int64, error)
}

// NewRunnerRepository returns the repository for the current driver
func NewRunnerRepository(db *sql.DB) RunnerRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &runnerRepoPostgres{db: db}
	}
	return &runnerRepoSQLite{db: db}
}

// scanRunnerRows builds a slice of Runner from query rows
func scanRunnerRows(rows *sql.Rows) ([]*Runner, error) {
	var list []*Runner
	for rows.Next() {
		run := &Runner{}
		if err := rows.Scan(
			&run.ID,
			&run.Token,
			&run.Name,
			&run.Tags,
			&run.MaxParallelTasks,
			&run.Status,
			&run.URL,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, run)
	}
	return list, rows.Err()
}

// --- SQLite ---

// runnerRepoSQLite implements RunnerRepository for SQLite
type runnerRepoSQLite struct {
	db *sql.DB
}

// Create inserts a row
func (r *runnerRepoSQLite) Create(runner *Runner) error {
	result, err := r.db.Exec(
		`INSERT INTO runners
		(token, name, tags, max_parallel_tasks, status, url)
		VALUES (?, ?, ?, ?, ?, ?)`,
		runner.Token,
		runner.Name,
		runner.Tags,
		runner.MaxParallelTasks,
		runner.Status,
		runner.URL,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	runner.ID = ID(strconv.FormatInt(lid, 10))
	return nil
}

// GetByID returns a runner by ID
func (r *runnerRepoSQLite) GetByID(id ID) (*Runner, error) {
	run := &Runner{}
	err := r.db.QueryRow(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		WHERE id = ?`,
		id.String(),
	).Scan(
		&run.ID,
		&run.Token,
		&run.Name,
		&run.Tags,
		&run.MaxParallelTasks,
		&run.Status,
		&run.URL,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return run, nil
}

// GetByToken returns a runner by token
func (r *runnerRepoSQLite) GetByToken(token string) (*Runner, error) {
	run := &Runner{}
	err := r.db.QueryRow(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		WHERE token = ?`,
		token,
	).Scan(
		&run.ID,
		&run.Token,
		&run.Name,
		&run.Tags,
		&run.MaxParallelTasks,
		&run.Status,
		&run.URL,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return run, nil
}

// Update updates a runner
func (r *runnerRepoSQLite) Update(runner *Runner) error {
	_, err := r.db.Exec(
		`UPDATE runners
		SET
			token = ?,
			name = ?,
			tags = ?,
			max_parallel_tasks = ?,
			status = ?,
			url = ?,
			updated_at = ?
		WHERE id = ?`,
		runner.Token,
		runner.Name,
		runner.Tags,
		runner.MaxParallelTasks,
		runner.Status,
		runner.URL,
		time.Now().UTC(),
		runner.ID.String(),
	)
	return err
}

// Delete removes a runner
func (r *runnerRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM runners WHERE id = ?`,
		id.String(),
	)
	return err
}

// List returns a list of runners
func (r *runnerRepoSQLite) List(limit, offset int) ([]*Runner, error) {
	rows, err := r.db.Query(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRunnerRows(rows)
}

// ListActive returns a list of active runners
func (r *runnerRepoSQLite) ListActive() ([]*Runner, error) {
	rows, err := r.db.Query(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		WHERE status = 'active'
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRunnerRows(rows)
}

// Count returns the total number of runners
func (r *runnerRepoSQLite) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM runners`,
	).Scan(&count)
	return count, err
}

// CountActive returns the total number of active runners
func (r *runnerRepoSQLite) CountActive() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM runners
		WHERE status = 'active'`,
	).Scan(&count)
	return count, err
}

// --- Postgres ---

// runnerRepoPostgres implements RunnerRepository for PostgreSQL
type runnerRepoPostgres struct {
	db *sql.DB
}

// Create inserts a row
func (r *runnerRepoPostgres) Create(runner *Runner) error {
	err := r.db.QueryRow(
		`INSERT INTO runners
		(token, name, tags, max_parallel_tasks, status, url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		runner.Token,
		runner.Name,
		runner.Tags,
		runner.MaxParallelTasks,
		runner.Status,
		runner.URL,
	).Scan(&runner.ID)
	return err
}

// GetByID returns a runner by ID
func (r *runnerRepoPostgres) GetByID(id ID) (*Runner, error) {
	run := &Runner{}
	err := r.db.QueryRow(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		WHERE id = $1`,
		id.String(),
	).Scan(
		&run.ID,
		&run.Token,
		&run.Name,
		&run.Tags,
		&run.MaxParallelTasks,
		&run.Status,
		&run.URL,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return run, nil
}

// GetByToken returns a runner by token
func (r *runnerRepoPostgres) GetByToken(token string) (*Runner, error) {
	run := &Runner{}
	err := r.db.QueryRow(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		WHERE token = $1`,
		token,
	).Scan(
		&run.ID,
		&run.Token,
		&run.Name,
		&run.Tags,
		&run.MaxParallelTasks,
		&run.Status,
		&run.URL,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return run, nil
}

// Update updates a runner
func (r *runnerRepoPostgres) Update(runner *Runner) error {
	_, err := r.db.Exec(
		`UPDATE runners
		SET
			token = $1,
			name = $2,
			tags = $3,
			max_parallel_tasks = $4,
			status = $5,
			url = $6,
			updated_at = $7
		WHERE id = $8`,
		runner.Token,
		runner.Name,
		runner.Tags,
		runner.MaxParallelTasks,
		runner.Status,
		runner.URL,
		time.Now().UTC(),
		runner.ID.String(),
	)
	return err
}

// Delete removes a runner
func (r *runnerRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM runners WHERE id = $1`,
		id.String(),
	)
	return err
}

// List returns a list of runners
func (r *runnerRepoPostgres) List(limit, offset int) ([]*Runner, error) {
	rows, err := r.db.Query(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRunnerRows(rows)
}

// ListActive returns a list of active runners
func (r *runnerRepoPostgres) ListActive() ([]*Runner, error) {
	rows, err := r.db.Query(
		`SELECT
			id, token, name, tags, max_parallel_tasks, status,
			url, created_at, updated_at
		FROM runners
		WHERE status = 'active'
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRunnerRows(rows)
}

// Count returns the total number of runners
func (r *runnerRepoPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM runners`,
	).Scan(&count)
	return count, err
}

// CountActive returns the total number of active runners
func (r *runnerRepoPostgres) CountActive() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM runners
		WHERE status = 'active'`,
	).Scan(&count)
	return count, err
}
