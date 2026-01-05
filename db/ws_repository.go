// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceRepo is a single row in the workspace_repositories table.
type WorkspaceRepo struct {
	ID               ID
	WorkspaceID      ID
	Name             string
	GitURL           string
	GitBranch        string
	GitTag           string
	SSHKeyID         *ID
	GitCredentialsID *ID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// WorkspaceRepoRepository is the interface for workspace repository CRUD.
type WorkspaceRepoRepository interface {
	Create(r *WorkspaceRepo) error
	GetByID(id ID) (*WorkspaceRepo, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceRepo, error)
	Update(r *WorkspaceRepo) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceRepo, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceRepoRepository returns the repository for the current driver.
func NewWorkspaceRepoRepository(db *sql.DB) WorkspaceRepoRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceRepoRepoPostgres{db: db}
	}
	return &workspaceRepoRepoSQLite{db: db}
}

func scanWorkspaceRepoRows(rows *sql.Rows) ([]*WorkspaceRepo, error) {
	var list []*WorkspaceRepo
	for rows.Next() {
		r := &WorkspaceRepo{}
		var sshKeyID, gitCredID interface{}
		if err := rows.Scan(
			&r.ID,
			&r.WorkspaceID,
			&r.Name,
			&r.GitURL,
			&r.GitBranch,
			&r.GitTag,
			&sshKeyID,
			&gitCredID,
			&r.CreatedAt,
			&r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if sshKeyID != nil {
			id := ID("")
			_ = id.Scan(sshKeyID)
			r.SSHKeyID = &id
		}
		if gitCredID != nil {
			id := ID("")
			_ = id.Scan(gitCredID)
			r.GitCredentialsID = &id
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

type workspaceRepoRepoSQLite struct {
	db *sql.DB
}

func (x *workspaceRepoRepoSQLite) Create(r *WorkspaceRepo) error {
	result, err := x.db.Exec(
		`INSERT INTO workspace_repositories (workspace_id, name, git_url, git_branch, git_tag, ssh_key_id, credentials_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.WorkspaceID.String(),
		r.Name,
		r.GitURL,
		r.GitBranch,
		r.GitTag,
		r.SSHKeyID.ValueOrNilPtr(),
		r.GitCredentialsID.ValueOrNilPtr(),
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = ID(strconv.FormatInt(lid, 10))
	row := x.db.QueryRow(
		`SELECT created_at, updated_at FROM workspace_repositories WHERE id = ?`,
		r.ID.String(),
	)
	_ = row.Scan(&r.CreatedAt, &r.UpdatedAt)
	return nil
}

func (x *workspaceRepoRepoSQLite) GetByID(id ID) (*WorkspaceRepo, error) {
	r := &WorkspaceRepo{}
	var sshKeyID, gitCredID interface{}
	err := x.db.QueryRow(
		`SELECT id, workspace_id, name, git_url, git_branch, git_tag, ssh_key_id, credentials_id, created_at, updated_at
		FROM workspace_repositories WHERE id = ?`,
		id.String(),
	).Scan(
		&r.ID,
		&r.WorkspaceID,
		&r.Name,
		&r.GitURL,
		&r.GitBranch,
		&r.GitTag,
		&sshKeyID,
		&gitCredID,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if sshKeyID != nil {
		idVal := ID("")
		_ = idVal.Scan(sshKeyID)
		r.SSHKeyID = &idVal
	}
	if gitCredID != nil {
		idVal := ID("")
		_ = idVal.Scan(gitCredID)
		r.GitCredentialsID = &idVal
	}
	return r, nil
}

func (x *workspaceRepoRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceRepo, error) {
	r, err := x.GetByID(id)
	if err != nil || r == nil {
		return r, err
	}
	if r.WorkspaceID != workspaceID {
		return nil, nil
	}
	return r, nil
}

func (x *workspaceRepoRepoSQLite) Update(r *WorkspaceRepo) error {
	_, err := x.db.Exec(
		`UPDATE workspace_repositories SET name = ?, git_url = ?, git_branch = ?, git_tag = ?, ssh_key_id = ?, credentials_id = ?, updated_at = ?
		WHERE id = ?`,
		r.Name,
		r.GitURL,
		r.GitBranch,
		r.GitTag,
		r.SSHKeyID.ValueOrNilPtr(),
		r.GitCredentialsID.ValueOrNilPtr(),
		time.Now().UTC(),
		r.ID.String(),
	)
	return err
}

func (x *workspaceRepoRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_repositories WHERE id = ?`,
		id.String(),
	)
	return err
}

func (x *workspaceRepoRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceRepo, error) {
	rows, err := x.db.Query(
		`SELECT id, workspace_id, name, git_url, git_branch, git_tag, ssh_key_id, credentials_id, created_at, updated_at
		FROM workspace_repositories WHERE workspace_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceRepoRows(rows)
}

func (x *workspaceRepoRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_repositories WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

type workspaceRepoRepoPostgres struct {
	db *sql.DB
}

func (x *workspaceRepoRepoPostgres) Create(r *WorkspaceRepo) error {
	err := x.db.QueryRow(
		`INSERT INTO workspace_repositories (workspace_id, name, git_url, git_branch, git_tag, ssh_key_id, credentials_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`,
		r.WorkspaceID.String(),
		r.Name,
		r.GitURL,
		r.GitBranch,
		r.GitTag,
		r.SSHKeyID.ValueOrNilPtr(),
		r.GitCredentialsID.ValueOrNilPtr(),
	).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
	return err
}

func (x *workspaceRepoRepoPostgres) GetByID(id ID) (*WorkspaceRepo, error) {
	r := &WorkspaceRepo{}
	var sshKeyID, gitCredID interface{}
	err := x.db.QueryRow(
		`SELECT id, workspace_id, name, git_url, git_branch, git_tag, ssh_key_id, credentials_id, created_at, updated_at
		FROM workspace_repositories WHERE id = $1`,
		id.String(),
	).Scan(
		&r.ID,
		&r.WorkspaceID,
		&r.Name,
		&r.GitURL,
		&r.GitBranch,
		&r.GitTag,
		&sshKeyID,
		&gitCredID,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if sshKeyID != nil {
		idVal := ID("")
		_ = idVal.Scan(sshKeyID)
		r.SSHKeyID = &idVal
	}
	if gitCredID != nil {
		idVal := ID("")
		_ = idVal.Scan(gitCredID)
		r.GitCredentialsID = &idVal
	}
	return r, nil
}

func (x *workspaceRepoRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceRepo, error) {
	r, err := x.GetByID(id)
	if err != nil || r == nil {
		return r, err
	}
	if r.WorkspaceID != workspaceID {
		return nil, nil
	}
	return r, nil
}

func (x *workspaceRepoRepoPostgres) Update(r *WorkspaceRepo) error {
	_, err := x.db.Exec(
		`UPDATE workspace_repositories SET name = $1, git_url = $2, git_branch = $3, git_tag = $4, ssh_key_id = $5, credentials_id = $6, updated_at = $7 WHERE id = $8`,
		r.Name,
		r.GitURL,
		r.GitBranch,
		r.GitTag,
		r.SSHKeyID.ValueOrNilPtr(),
		r.GitCredentialsID.ValueOrNilPtr(),
		time.Now().UTC(),
		r.ID.String(),
	)
	return err
}

func (x *workspaceRepoRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_repositories WHERE id = $1`,
		id.String(),
	)
	return err
}

func (x *workspaceRepoRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceRepo, error) {
	rows, err := x.db.Query(
		`SELECT id, workspace_id, name, git_url, git_branch, git_tag, ssh_key_id, credentials_id, created_at, updated_at
		FROM workspace_repositories WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceRepoRows(rows)
}

func (x *workspaceRepoRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_repositories WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
