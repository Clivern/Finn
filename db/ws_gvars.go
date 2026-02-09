// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceVariableGroup is the DB row for a workspace variable group.
type WorkspaceVariableGroup struct {
	ID                         ID
	WorkspaceID                ID
	Name                       string
	ExtraVariables             string
	EnvironmentVariables       string
	SecretExtraVariables       string
	SecretEnvironmentVariables string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

// WorkspaceVariableGroupRepository does CRUD for workspace variable groups.
type WorkspaceVariableGroupRepository interface {
	Create(g *WorkspaceVariableGroup) error
	GetByID(id ID) (*WorkspaceVariableGroup, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceVariableGroup, error)
	Update(g *WorkspaceVariableGroup) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceVariableGroup, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceVariableGroupRepository returns the repository for the current driver.
func NewWorkspaceVariableGroupRepository(db *sql.DB) WorkspaceVariableGroupRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceVariableGroupRepoPostgres{db: db}
	}
	return &workspaceVariableGroupRepoSQLite{db: db}
}

func scanWorkspaceVariableGroupRows(rows *sql.Rows) ([]*WorkspaceVariableGroup, error) {
	var list []*WorkspaceVariableGroup
	for rows.Next() {
		g := &WorkspaceVariableGroup{}
		if err := rows.Scan(
			&g.ID,
			&g.WorkspaceID,
			&g.Name,
			&g.ExtraVariables,
			&g.EnvironmentVariables,
			&g.SecretExtraVariables,
			&g.SecretEnvironmentVariables,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

type workspaceVariableGroupRepoSQLite struct {
	db *sql.DB
}

// Create creates a new workspace variable group.
func (x *workspaceVariableGroupRepoSQLite) Create(g *WorkspaceVariableGroup) error {
	result, err := x.db.Exec(
		`INSERT INTO workspace_variable_groups
		(workspace_id, name, extra_variables, environment_variables, secret_extra_variables, secret_environment_variables)
		VALUES (?, ?, ?, ?, ?, ?)`,
		g.WorkspaceID.String(),
		g.Name,
		g.ExtraVariables,
		g.EnvironmentVariables,
		g.SecretExtraVariables,
		g.SecretEnvironmentVariables,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	g.ID = ID(strconv.FormatInt(lid, 10))
	row := x.db.QueryRow(
		`SELECT
			created_at,
			updated_at
		FROM workspace_variable_groups
		WHERE id = ?`,
		g.ID.String(),
	)
	_ = row.Scan(&g.CreatedAt, &g.UpdatedAt)
	return nil
}

// GetByID returns a workspace variable group by ID.
func (x *workspaceVariableGroupRepoSQLite) GetByID(id ID) (*WorkspaceVariableGroup, error) {
	g := &WorkspaceVariableGroup{}
	err := x.db.QueryRow(
		`SELECT
			id, workspace_id, name, extra_variables, environment_variables,
			secret_extra_variables, secret_environment_variables, created_at, updated_at
		FROM workspace_variable_groups
		WHERE id = ?`,
		id.String(),
	).Scan(
		&g.ID,
		&g.WorkspaceID,
		&g.Name,
		&g.ExtraVariables,
		&g.EnvironmentVariables,
		&g.SecretExtraVariables,
		&g.SecretEnvironmentVariables,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return g, nil
}

// GetByIDAndWorkspace returns a workspace variable group by ID scoped to a workspace.
func (x *workspaceVariableGroupRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceVariableGroup, error) {
	g, err := x.GetByID(id)
	if err != nil || g == nil {
		return g, err
	}
	if g.WorkspaceID != workspaceID {
		return nil, nil
	}
	return g, nil
}

// Update updates a workspace variable group.
func (x *workspaceVariableGroupRepoSQLite) Update(g *WorkspaceVariableGroup) error {
	_, err := x.db.Exec(
		`UPDATE workspace_variable_groups
		SET
			name = ?,
			extra_variables = ?,
			environment_variables = ?,
			secret_extra_variables = ?,
			secret_environment_variables = ?,
			updated_at = ?
		WHERE id = ?`,
		g.Name,
		g.ExtraVariables,
		g.EnvironmentVariables,
		g.SecretExtraVariables,
		g.SecretEnvironmentVariables,
		time.Now().UTC(),
		g.ID.String(),
	)
	return err
}

// Delete deletes a workspace variable group.
func (x *workspaceVariableGroupRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_variable_groups WHERE id = ?`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID lists workspace variable groups for a workspace with pagination.
func (x *workspaceVariableGroupRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceVariableGroup, error) {
	rows, err := x.db.Query(
		`SELECT
			id, workspace_id, name, extra_variables, environment_variables,
			secret_extra_variables, secret_environment_variables, created_at, updated_at
		FROM workspace_variable_groups
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
	return scanWorkspaceVariableGroupRows(rows)
}

// CountByWorkspaceID returns how many variable groups exist in a workspace.
func (x *workspaceVariableGroupRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_variable_groups
		WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

type workspaceVariableGroupRepoPostgres struct {
	db *sql.DB
}

// Create creates a new workspace variable group.
func (x *workspaceVariableGroupRepoPostgres) Create(g *WorkspaceVariableGroup) error {
	err := x.db.QueryRow(
		`INSERT INTO workspace_variable_groups
		(workspace_id, name, extra_variables, environment_variables, secret_extra_variables, secret_environment_variables)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		g.WorkspaceID.String(),
		g.Name,
		g.ExtraVariables,
		g.EnvironmentVariables,
		g.SecretExtraVariables,
		g.SecretEnvironmentVariables,
	).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
	return err
}

// GetByID returns a workspace variable group by ID.
func (x *workspaceVariableGroupRepoPostgres) GetByID(id ID) (*WorkspaceVariableGroup, error) {
	g := &WorkspaceVariableGroup{}
	err := x.db.QueryRow(
		`SELECT
			id, workspace_id, name, extra_variables, environment_variables,
			secret_extra_variables, secret_environment_variables, created_at, updated_at
		FROM workspace_variable_groups
		WHERE id = $1`,
		id.String(),
	).Scan(
		&g.ID,
		&g.WorkspaceID,
		&g.Name,
		&g.ExtraVariables,
		&g.EnvironmentVariables,
		&g.SecretExtraVariables,
		&g.SecretEnvironmentVariables,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return g, nil
}

// GetByIDAndWorkspace returns a workspace variable group by ID scoped to a workspace.
func (x *workspaceVariableGroupRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceVariableGroup, error) {
	g, err := x.GetByID(id)
	if err != nil || g == nil {
		return g, err
	}
	if g.WorkspaceID != workspaceID {
		return nil, nil
	}
	return g, nil
}

// Update updates a workspace variable group.
func (x *workspaceVariableGroupRepoPostgres) Update(g *WorkspaceVariableGroup) error {
	_, err := x.db.Exec(
		`UPDATE workspace_variable_groups
		SET
			name = $1,
			extra_variables = $2,
			environment_variables = $3,
			secret_extra_variables = $4,
			secret_environment_variables = $5,
			updated_at = $6
		WHERE id = $7`,
		g.Name,
		g.ExtraVariables,
		g.EnvironmentVariables,
		g.SecretExtraVariables,
		g.SecretEnvironmentVariables,
		time.Now().UTC(),
		g.ID.String(),
	)
	return err
}

// Delete deletes a workspace variable group.
func (x *workspaceVariableGroupRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_variable_groups WHERE id = $1`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID lists workspace variable groups for a workspace with pagination.
func (x *workspaceVariableGroupRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceVariableGroup, error) {
	rows, err := x.db.Query(
		`SELECT
			id, workspace_id, name, extra_variables, environment_variables,
			secret_extra_variables, secret_environment_variables, created_at, updated_at
		FROM workspace_variable_groups
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
	return scanWorkspaceVariableGroupRows(rows)
}

// CountByWorkspaceID returns how many variable groups exist in a workspace.
func (x *workspaceVariableGroupRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_variable_groups
		WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
