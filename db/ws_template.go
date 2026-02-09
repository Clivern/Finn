// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceTemplate is a single row in the workspace_templates table.
type WorkspaceTemplate struct {
	ID                  ID
	WorkspaceID         ID
	RepositoryID        ID
	Name                string
	Description         string
	Engine              string
	WorkingDirectory    string
	GitBranchOverride   string
	GitTagOverride      string
	RunnerTags          string
	AllowConcurrentRuns bool
	MaxConcurrentRuns   int
	ApprovalRequired    bool
	ScriptConfig        string
	IaCConfig           string
	AnsibleConfig       string
	ExtraArgs           string
	GroupVarsID         *ID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// WorkspaceTemplateRepository is the interface for workspace template CRUD.
type WorkspaceTemplateRepository interface {
	Create(t *WorkspaceTemplate) error
	GetByID(id ID) (*WorkspaceTemplate, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceTemplate, error)
	Update(t *WorkspaceTemplate) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceTemplate, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceTemplateRepository returns the repository for the current driver.
func NewWorkspaceTemplateRepository(db *sql.DB) WorkspaceTemplateRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceTemplateRepoPostgres{db: db}
	}
	return &workspaceTemplateRepoSQLite{db: db}
}

func scanWorkspaceTemplateRows(rows *sql.Rows) ([]*WorkspaceTemplate, error) {
	var list []*WorkspaceTemplate
	for rows.Next() {
		t := &WorkspaceTemplate{}
		var description, gitBranchOverride, gitTagOverride, runnerTags sql.NullString
		var scriptConfig, iacConfig, ansibleConfig, extraArgs sql.NullString
		var groupVarsID interface{}

		if err := rows.Scan(
			&t.ID,
			&t.WorkspaceID,
			&t.RepositoryID,
			&t.Name,
			&description,
			&t.Engine,
			&t.WorkingDirectory,
			&gitBranchOverride,
			&gitTagOverride,
			&runnerTags,
			&t.AllowConcurrentRuns,
			&t.MaxConcurrentRuns,
			&t.ApprovalRequired,
			&scriptConfig,
			&iacConfig,
			&ansibleConfig,
			&extraArgs,
			&groupVarsID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			t.Description = description.String
		}
		if gitBranchOverride.Valid {
			t.GitBranchOverride = gitBranchOverride.String
		}
		if gitTagOverride.Valid {
			t.GitTagOverride = gitTagOverride.String
		}
		if runnerTags.Valid {
			t.RunnerTags = runnerTags.String
		}
		if scriptConfig.Valid {
			t.ScriptConfig = scriptConfig.String
		}
		if iacConfig.Valid {
			t.IaCConfig = iacConfig.String
		}
		if ansibleConfig.Valid {
			t.AnsibleConfig = ansibleConfig.String
		}
		if extraArgs.Valid {
			t.ExtraArgs = extraArgs.String
		}
		if groupVarsID != nil {
			id := ID("")
			_ = id.Scan(groupVarsID)
			t.GroupVarsID = &id
		}

		list = append(list, t)
	}
	return list, rows.Err()
}

type workspaceTemplateRepoSQLite struct {
	db *sql.DB
}

// Create creates a new workspace template.
func (x *workspaceTemplateRepoSQLite) Create(t *WorkspaceTemplate) error {
	result, err := x.db.Exec(
		`INSERT INTO workspace_templates
		(
			workspace_id,
			repository_id,
			name,
			description,
			engine,
			working_directory,
			git_branch_override,
			git_tag_override,
			runner_tags,
			allow_concurrent_runs,
			max_concurrent_runs,
			approval_required,
			script_config,
			iac_config,
			ansible_config,
			extra_args,
			group_vars_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.WorkspaceID.String(),
		t.RepositoryID.String(),
		t.Name,
		t.Description,
		t.Engine,
		t.WorkingDirectory,
		t.GitBranchOverride,
		t.GitTagOverride,
		t.RunnerTags,
		t.AllowConcurrentRuns,
		t.MaxConcurrentRuns,
		t.ApprovalRequired,
		t.ScriptConfig,
		t.IaCConfig,
		t.AnsibleConfig,
		t.ExtraArgs,
		t.GroupVarsID.ValueOrNilPtr(),
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = ID(strconv.FormatInt(lid, 10))
	row := x.db.QueryRow(
		`SELECT created_at, updated_at
		FROM workspace_templates
		WHERE id = ?`,
		t.ID.String(),
	)
	_ = row.Scan(&t.CreatedAt, &t.UpdatedAt)
	return nil
}

// GetByID returns a workspace template by ID.
func (x *workspaceTemplateRepoSQLite) GetByID(id ID) (*WorkspaceTemplate, error) {
	t := &WorkspaceTemplate{}
	var description, gitBranchOverride, gitTagOverride, runnerTags sql.NullString
	var scriptConfig, iacConfig, ansibleConfig, extraArgs sql.NullString
	var groupVarsID interface{}
	err := x.db.QueryRow(
		`SELECT
			id,
			workspace_id,
			repository_id,
			name,
			description,
			engine,
			working_directory,
			git_branch_override,
			git_tag_override,
			runner_tags,
			allow_concurrent_runs,
			max_concurrent_runs,
			approval_required,
			script_config,
			iac_config,
			ansible_config,
			extra_args,
			group_vars_id,
			created_at,
			updated_at
		FROM workspace_templates
		WHERE id = ?`,
		id.String(),
	).Scan(
		&t.ID,
		&t.WorkspaceID,
		&t.RepositoryID,
		&t.Name,
		&description,
		&t.Engine,
		&t.WorkingDirectory,
		&gitBranchOverride,
		&gitTagOverride,
		&runnerTags,
		&t.AllowConcurrentRuns,
		&t.MaxConcurrentRuns,
		&t.ApprovalRequired,
		&scriptConfig,
		&iacConfig,
		&ansibleConfig,
		&extraArgs,
		&groupVarsID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if description.Valid {
		t.Description = description.String
	}
	if gitBranchOverride.Valid {
		t.GitBranchOverride = gitBranchOverride.String
	}
	if gitTagOverride.Valid {
		t.GitTagOverride = gitTagOverride.String
	}
	if runnerTags.Valid {
		t.RunnerTags = runnerTags.String
	}
	if scriptConfig.Valid {
		t.ScriptConfig = scriptConfig.String
	}
	if iacConfig.Valid {
		t.IaCConfig = iacConfig.String
	}
	if ansibleConfig.Valid {
		t.AnsibleConfig = ansibleConfig.String
	}
	if extraArgs.Valid {
		t.ExtraArgs = extraArgs.String
	}
	if groupVarsID != nil {
		idVal := ID("")
		_ = idVal.Scan(groupVarsID)
		t.GroupVarsID = &idVal
	}
	return t, nil
}

// GetByIDAndWorkspace returns a workspace template by ID scoped to a workspace.
func (x *workspaceTemplateRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceTemplate, error) {
	t, err := x.GetByID(id)
	if err != nil || t == nil {
		return t, err
	}
	if t.WorkspaceID != workspaceID {
		return nil, nil
	}
	return t, nil
}

// Update updates a workspace template.
func (x *workspaceTemplateRepoSQLite) Update(t *WorkspaceTemplate) error {
	_, err := x.db.Exec(
		`UPDATE workspace_templates
		SET
			repository_id = ?,
			name = ?,
			description = ?,
			engine = ?,
			working_directory = ?,
			git_branch_override = ?,
			git_tag_override = ?,
			runner_tags = ?,
			allow_concurrent_runs = ?,
			max_concurrent_runs = ?,
			approval_required = ?,
			script_config = ?,
			iac_config = ?,
			ansible_config = ?,
			extra_args = ?,
			group_vars_id = ?,
			updated_at = ?
		WHERE id = ?`,
		t.RepositoryID.String(),
		t.Name,
		t.Description,
		t.Engine,
		t.WorkingDirectory,
		t.GitBranchOverride,
		t.GitTagOverride,
		t.RunnerTags,
		t.AllowConcurrentRuns,
		t.MaxConcurrentRuns,
		t.ApprovalRequired,
		t.ScriptConfig,
		t.IaCConfig,
		t.AnsibleConfig,
		t.ExtraArgs,
		t.GroupVarsID.ValueOrNilPtr(),
		time.Now().UTC(),
		t.ID.String(),
	)
	return err
}

// Delete deletes a workspace template.
func (x *workspaceTemplateRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_templates WHERE id = ?`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID lists workspace templates for a workspace with pagination.
func (x *workspaceTemplateRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceTemplate, error) {
	rows, err := x.db.Query(
		`SELECT
			id,
			workspace_id,
			repository_id,
			name,
			description,
			engine,
			working_directory,
			git_branch_override,
			git_tag_override,
			runner_tags,
			allow_concurrent_runs,
			max_concurrent_runs,
			approval_required,
			script_config,
			iac_config,
			ansible_config,
			extra_args,
			group_vars_id,
			created_at,
			updated_at
		FROM workspace_templates
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
	return scanWorkspaceTemplateRows(rows)
}

// CountByWorkspaceID returns how many templates exist in a workspace.
func (x *workspaceTemplateRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_templates
		WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

type workspaceTemplateRepoPostgres struct {
	db *sql.DB
}

// Create creates a new workspace template.
func (x *workspaceTemplateRepoPostgres) Create(t *WorkspaceTemplate) error {
	err := x.db.QueryRow(
		`INSERT INTO workspace_templates
		(
			workspace_id,
			repository_id,
			name,
			description,
			engine,
			working_directory,
			git_branch_override,
			git_tag_override,
			runner_tags,
			allow_concurrent_runs,
			max_concurrent_runs,
			approval_required,
			script_config,
			iac_config,
			ansible_config,
			extra_args,
			group_vars_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at`,
		t.WorkspaceID.String(),
		t.RepositoryID.String(),
		t.Name,
		t.Description,
		t.Engine,
		t.WorkingDirectory,
		t.GitBranchOverride,
		t.GitTagOverride,
		t.RunnerTags,
		t.AllowConcurrentRuns,
		t.MaxConcurrentRuns,
		t.ApprovalRequired,
		t.ScriptConfig,
		t.IaCConfig,
		t.AnsibleConfig,
		t.ExtraArgs,
		t.GroupVarsID.ValueOrNilPtr(),
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

// GetByID returns a workspace template by ID.
func (x *workspaceTemplateRepoPostgres) GetByID(id ID) (*WorkspaceTemplate, error) {
	t := &WorkspaceTemplate{}
	var description, gitBranchOverride, gitTagOverride, runnerTags sql.NullString
	var scriptConfig, iacConfig, ansibleConfig, extraArgs sql.NullString
	var groupVarsID interface{}
	err := x.db.QueryRow(
		`SELECT
			id,
			workspace_id,
			repository_id,
			name,
			description,
			engine,
			working_directory,
			git_branch_override,
			git_tag_override,
			runner_tags,
			allow_concurrent_runs,
			max_concurrent_runs,
			approval_required,
			script_config,
			iac_config,
			ansible_config,
			extra_args,
			group_vars_id,
			created_at,
			updated_at
		FROM workspace_templates
		WHERE id = $1`,
		id.String(),
	).Scan(
		&t.ID,
		&t.WorkspaceID,
		&t.RepositoryID,
		&t.Name,
		&description,
		&t.Engine,
		&t.WorkingDirectory,
		&gitBranchOverride,
		&gitTagOverride,
		&runnerTags,
		&t.AllowConcurrentRuns,
		&t.MaxConcurrentRuns,
		&t.ApprovalRequired,
		&scriptConfig,
		&iacConfig,
		&ansibleConfig,
		&extraArgs,
		&groupVarsID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if description.Valid {
		t.Description = description.String
	}
	if gitBranchOverride.Valid {
		t.GitBranchOverride = gitBranchOverride.String
	}
	if gitTagOverride.Valid {
		t.GitTagOverride = gitTagOverride.String
	}
	if runnerTags.Valid {
		t.RunnerTags = runnerTags.String
	}
	if scriptConfig.Valid {
		t.ScriptConfig = scriptConfig.String
	}
	if iacConfig.Valid {
		t.IaCConfig = iacConfig.String
	}
	if ansibleConfig.Valid {
		t.AnsibleConfig = ansibleConfig.String
	}
	if extraArgs.Valid {
		t.ExtraArgs = extraArgs.String
	}
	if groupVarsID != nil {
		idVal := ID("")
		_ = idVal.Scan(groupVarsID)
		t.GroupVarsID = &idVal
	}
	return t, nil
}

// GetByIDAndWorkspace returns a workspace template by ID scoped to a workspace.
func (x *workspaceTemplateRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceTemplate, error) {
	t, err := x.GetByID(id)
	if err != nil || t == nil {
		return t, err
	}
	if t.WorkspaceID != workspaceID {
		return nil, nil
	}
	return t, nil
}

// Update updates a workspace template.
func (x *workspaceTemplateRepoPostgres) Update(t *WorkspaceTemplate) error {
	_, err := x.db.Exec(
		`UPDATE workspace_templates
		SET
			repository_id = $1,
			name = $2,
			description = $3,
			engine = $4,
			working_directory = $5,
			git_branch_override = $6,
			git_tag_override = $7,
			runner_tags = $8,
			allow_concurrent_runs = $9,
			max_concurrent_runs = $10,
			approval_required = $11,
			script_config = $12,
			iac_config = $13,
			ansible_config = $14,
			extra_args = $15,
			group_vars_id = $16,
			updated_at = $17
		WHERE id = $18`,
		t.RepositoryID.String(),
		t.Name,
		t.Description,
		t.Engine,
		t.WorkingDirectory,
		t.GitBranchOverride,
		t.GitTagOverride,
		t.RunnerTags,
		t.AllowConcurrentRuns,
		t.MaxConcurrentRuns,
		t.ApprovalRequired,
		t.ScriptConfig,
		t.IaCConfig,
		t.AnsibleConfig,
		t.ExtraArgs,
		t.GroupVarsID.ValueOrNilPtr(),
		time.Now().UTC(),
		t.ID.String(),
	)
	return err
}

// Delete deletes a workspace template.
func (x *workspaceTemplateRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_templates WHERE id = $1`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID lists workspace templates for a workspace with pagination.
func (x *workspaceTemplateRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceTemplate, error) {
	rows, err := x.db.Query(
		`SELECT
			id,
			workspace_id,
			repository_id,
			name,
			description,
			engine,
			working_directory,
			git_branch_override,
			git_tag_override,
			runner_tags,
			allow_concurrent_runs,
			max_concurrent_runs,
			approval_required,
			script_config,
			iac_config,
			ansible_config,
			extra_args,
			group_vars_id,
			created_at,
			updated_at
		FROM workspace_templates
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
	return scanWorkspaceTemplateRows(rows)
}

// CountByWorkspaceID returns how many templates exist in a workspace.
func (x *workspaceTemplateRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_templates
		WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
