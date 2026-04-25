// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"encoding/json"
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

func TemplateToPayload(t *WorkspaceTemplate) map[string]interface{} {
	return TemplateToPayloadFull(t, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

// TemplateToPayloadFull builds a nested payload suitable for JSON encoding and comparison
// with `template.json`. Related entities are optional; pass nil if not available.
//
// Notes:
//   - Some DB fields are JSON-encoded strings (RunnerTags / ScriptConfig / IaCConfig / AnsibleConfig).
//     These are decoded into nested maps when valid JSON.
//   - SSH keys / git credentials are passed as maps so callers can choose whether values are
//     encrypted-at-rest strings or decrypted display values.
func TemplateToPayloadFull(
	t *WorkspaceTemplate,
	repo *WorkspaceRepo,
	repoSSHKey map[string]interface{},
	repoCred map[string]interface{},
	groupVars *WorkspaceVariableGroup,
	ansibleInventory *WorkspaceInventory,
	invSSHKey map[string]interface{},
	invCred map[string]interface{},
	becomeSSHKey map[string]interface{},
	becomeCred map[string]interface{},
	invRepo *WorkspaceRepo,
	invRepoSSHKey map[string]interface{},
	invRepoCred map[string]interface{},
) map[string]interface{} {
	if t == nil {
		return map[string]interface{}{}
	}

	payload := map[string]interface{}{
		"id":                    t.ID.String(),
		"name":                  t.Name,
		"description":           t.Description,
		"engine":                t.Engine,
		"working_dir":           t.WorkingDirectory,
		"git_branch_override":   t.GitBranchOverride,
		"git_tag_override":      t.GitTagOverride,
		"allow_concurrent_runs": t.AllowConcurrentRuns,
		"max_concurrent_runs":   t.MaxConcurrentRuns,
		"approval_required":     t.ApprovalRequired,
		"extra_args":            t.ExtraArgs,
	}

	if m := decodeJSONToMap(t.RunnerTags); m != nil {
		payload["runner_tags"] = m
	}
	if m := decodeJSONToMap(t.ScriptConfig); m != nil {
		payload["script_config"] = m
	}
	if m := decodeJSONToMap(t.IaCConfig); m != nil {
		payload["iac_config"] = m
	}

	var ansibleCfg map[string]interface{}
	if m := decodeJSONToMap(t.AnsibleConfig); m != nil {
		ansibleCfg = m
		payload["ansible_config"] = ansibleCfg
	}

	if repo != nil {
		r := map[string]interface{}{
			"id":         repo.ID.String(),
			"name":       repo.Name,
			"git_url":    repo.GitURL,
			"git_branch": repo.GitBranch,
			"git_tag":    repo.GitTag,
		}
		if len(repoSSHKey) > 0 {
			r["ssh_key"] = repoSSHKey
		}
		if len(repoCred) > 0 {
			r["git_credential"] = repoCred
		}
		payload["repository"] = r
	}

	if groupVars != nil {
		payload["group_vars"] = map[string]interface{}{
			"id":                           groupVars.ID.String(),
			"name":                         groupVars.Name,
			"extra_variables":              groupVars.ExtraVariables,
			"environment_variables":        groupVars.EnvironmentVariables,
			"secret_extra_variables":       groupVars.SecretExtraVariables,
			"secret_environment_variables": groupVars.SecretEnvironmentVariables,
		}
	}

	if ansibleCfg != nil && ansibleInventory != nil {
		inv := map[string]interface{}{
			"id":             ansibleInventory.ID.String(),
			"name":           ansibleInventory.Name,
			"type":           ansibleInventory.Type,
			"inventory":      ansibleInventory.Inventory,
			"inventory_path": ansibleInventory.InventoryPath,
		}
		if len(invSSHKey) > 0 {
			inv["ssh_key"] = invSSHKey
		}
		if len(invCred) > 0 {
			inv["git_credential"] = invCred
		}
		if len(becomeSSHKey) > 0 {
			inv["become_ssh_key"] = becomeSSHKey
		}
		if len(becomeCred) > 0 {
			inv["become_git_credential"] = becomeCred
		}
		if invRepo != nil {
			invR := map[string]interface{}{
				"id":         invRepo.ID.String(),
				"name":       invRepo.Name,
				"git_url":    invRepo.GitURL,
				"git_branch": invRepo.GitBranch,
				"git_tag":    invRepo.GitTag,
			}
			if len(invRepoSSHKey) > 0 {
				invR["ssh_key"] = invRepoSSHKey
			}
			if len(invRepoCred) > 0 {
				invR["git_credential"] = invRepoCred
			}
			inv["repository"] = invR
		}
		ansibleCfg["inventory"] = inv
	}

	return payload
}

func decodeJSONToMap(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	if len(m) == 0 {
		return nil
	}
	return m
}
