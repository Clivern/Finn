// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceGitCred is a single row in the workspace_git_credentials table.
type WorkspaceGitCred struct {
	ID              ID
	WorkspaceID     ID
	Name            string
	CommitterName   *string
	CommitterEmail  *string
	Username        *string
	PasswordOrToken *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WorkspaceGitCredRepository defines workspace git credentials CRUD.
type WorkspaceGitCredRepository interface {
	Create(c *WorkspaceGitCred) error
	GetByID(id ID) (*WorkspaceGitCred, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceGitCred, error)
	Update(c *WorkspaceGitCred) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceGitCred, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceGitCredRepository returns the repository for the current driver.
func NewWorkspaceGitCredRepository(db *sql.DB) WorkspaceGitCredRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceGitCredRepoPostgres{db: db}
	}
	return &workspaceGitCredRepoSQLite{db: db}
}

func scanWorkspaceGitCredRows(rows *sql.Rows) ([]*WorkspaceGitCred, error) {
	var list []*WorkspaceGitCred
	for rows.Next() {
		c := &WorkspaceGitCred{}
		if err := rows.Scan(
			&c.ID,
			&c.WorkspaceID,
			&c.Name,
			&c.CommitterName,
			&c.CommitterEmail,
			&c.Username,
			&c.PasswordOrToken,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

type workspaceGitCredRepoSQLite struct {
	db *sql.DB
}

func (x *workspaceGitCredRepoSQLite) Create(c *WorkspaceGitCred) error {
	result, err := x.db.Exec(
		`INSERT INTO workspace_git_credentials (workspace_id, name, committer_name, committer_email, username, password_or_token)
		VALUES (?, ?, ?, ?, ?, ?)`,
		c.WorkspaceID.String(), c.Name, c.CommitterName, c.CommitterEmail, c.Username, c.PasswordOrToken,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = ID(strconv.FormatInt(lid, 10))
	row := x.db.QueryRow(`SELECT created_at, updated_at FROM workspace_git_credentials WHERE id = ?`, c.ID.String())
	_ = row.Scan(&c.CreatedAt, &c.UpdatedAt)
	return nil
}

func (x *workspaceGitCredRepoSQLite) GetByID(id ID) (*WorkspaceGitCred, error) {
	c := &WorkspaceGitCred{}
	err := x.db.QueryRow(
		`SELECT id, workspace_id, name, committer_name, committer_email, username, password_or_token, created_at, updated_at
		FROM workspace_git_credentials WHERE id = ?`,
		id.String(),
	).Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.CommitterName, &c.CommitterEmail, &c.Username, &c.PasswordOrToken, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (x *workspaceGitCredRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceGitCred, error) {
	c, err := x.GetByID(id)
	if err != nil || c == nil {
		return c, err
	}
	if c.WorkspaceID != workspaceID {
		return nil, nil
	}
	return c, nil
}

func (x *workspaceGitCredRepoSQLite) Update(c *WorkspaceGitCred) error {
	_, err := x.db.Exec(
		`UPDATE workspace_git_credentials SET name = ?, committer_name = ?, committer_email = ?, username = ?, password_or_token = ?, updated_at = ?
		WHERE id = ?`,
		c.Name, c.CommitterName, c.CommitterEmail, c.Username, c.PasswordOrToken, time.Now().UTC(), c.ID.String(),
	)
	return err
}

func (x *workspaceGitCredRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(`DELETE FROM workspace_git_credentials WHERE id = ?`, id.String())
	return err
}

func (x *workspaceGitCredRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceGitCred, error) {
	rows, err := x.db.Query(
		`SELECT id, workspace_id, name, committer_name, committer_email, username, password_or_token, created_at, updated_at
		FROM workspace_git_credentials WHERE workspace_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		workspaceID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceGitCredRows(rows)
}

func (x *workspaceGitCredRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(`SELECT COUNT(*) FROM workspace_git_credentials WHERE workspace_id = ?`, workspaceID.String()).Scan(&count)
	return count, err
}

type workspaceGitCredRepoPostgres struct {
	db *sql.DB
}

func (x *workspaceGitCredRepoPostgres) Create(c *WorkspaceGitCred) error {
	err := x.db.QueryRow(
		`INSERT INTO workspace_git_credentials (workspace_id, name, committer_name, committer_email, username, password_or_token)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`,
		c.WorkspaceID.String(), c.Name, c.CommitterName, c.CommitterEmail, c.Username, c.PasswordOrToken,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return err
}

func (x *workspaceGitCredRepoPostgres) GetByID(id ID) (*WorkspaceGitCred, error) {
	c := &WorkspaceGitCred{}
	err := x.db.QueryRow(
		`SELECT id, workspace_id, name, committer_name, committer_email, username, password_or_token, created_at, updated_at
		FROM workspace_git_credentials WHERE id = $1`,
		id.String(),
	).Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.CommitterName, &c.CommitterEmail, &c.Username, &c.PasswordOrToken, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (x *workspaceGitCredRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceGitCred, error) {
	c, err := x.GetByID(id)
	if err != nil || c == nil {
		return c, err
	}
	if c.WorkspaceID != workspaceID {
		return nil, nil
	}
	return c, nil
}

func (x *workspaceGitCredRepoPostgres) Update(c *WorkspaceGitCred) error {
	_, err := x.db.Exec(
		`UPDATE workspace_git_credentials SET name = $1, committer_name = $2, committer_email = $3, username = $4, password_or_token = $5, updated_at = $6 WHERE id = $7`,
		c.Name, c.CommitterName, c.CommitterEmail, c.Username, c.PasswordOrToken, time.Now().UTC(), c.ID.String(),
	)
	return err
}

func (x *workspaceGitCredRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(`DELETE FROM workspace_git_credentials WHERE id = $1`, id.String())
	return err
}

func (x *workspaceGitCredRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceGitCred, error) {
	rows, err := x.db.Query(
		`SELECT id, workspace_id, name, committer_name, committer_email, username, password_or_token, created_at, updated_at
		FROM workspace_git_credentials WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		workspaceID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceGitCredRows(rows)
}

func (x *workspaceGitCredRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(`SELECT COUNT(*) FROM workspace_git_credentials WHERE workspace_id = $1`, workspaceID.String()).Scan(&count)
	return count, err
}
