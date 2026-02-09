// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

const CredentialTypeGit = "git"
const CredentialTypeNone = "none"
const CredentialTypeCredential = "credential"

// WorkspaceCred is a single row in the workspace_credentials table
type WorkspaceCred struct {
	ID              ID
	WorkspaceID     ID
	Name            string
	Type            string
	CommitterName   string
	CommitterEmail  string
	Username        string
	PasswordOrToken string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WorkspaceCredRepository is the interface for workspace credentials CRUD.
type WorkspaceCredRepository interface {
	Create(c *WorkspaceCred) error
	GetByID(id ID) (*WorkspaceCred, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceCred, error)
	Update(c *WorkspaceCred) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceCred, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceCredRepository returns the repository for the current driver.
func NewWorkspaceCredRepository(db *sql.DB) WorkspaceCredRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceCredRepoPostgres{db: db}
	}
	return &workspaceCredRepoSQLite{db: db}
}

func scanWorkspaceCredRows(rows *sql.Rows) ([]*WorkspaceCred, error) {
	var list []*WorkspaceCred
	for rows.Next() {
		c := &WorkspaceCred{}
		var committerName, committerEmail sql.NullString
		if err := rows.Scan(
			&c.ID,
			&c.WorkspaceID,
			&c.Name,
			&c.Type,
			&committerName,
			&committerEmail,
			&c.Username,
			&c.PasswordOrToken,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if committerName.Valid {
			c.CommitterName = committerName.String
		}
		if committerEmail.Valid {
			c.CommitterEmail = committerEmail.String
		}
		if c.Type == "" {
			c.Type = CredentialTypeGit
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

type workspaceCredRepoSQLite struct {
	db *sql.DB
}

// Create creates a new credential
func (x *workspaceCredRepoSQLite) Create(c *WorkspaceCred) error {
	if c.Type == "" {
		c.Type = CredentialTypeGit
	}
	result, err := x.db.Exec(
		`INSERT INTO workspace_credentials
		(workspace_id, name, type, committer_name, committer_email, username, password_or_token)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.WorkspaceID.String(),
		c.Name,
		c.Type,
		nullString(c.CommitterName),
		nullString(c.CommitterEmail),
		c.Username,
		c.PasswordOrToken,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = ID(strconv.FormatInt(lid, 10))
	row := x.db.QueryRow(
		`SELECT
			created_at,
			updated_at
		FROM workspace_credentials
		WHERE id = ?`,
		c.ID.String(),
	)
	_ = row.Scan(&c.CreatedAt, &c.UpdatedAt)
	return nil
}

// GetByID returns the credential for the given ID.
func (x *workspaceCredRepoSQLite) GetByID(id ID) (*WorkspaceCred, error) {
	c := &WorkspaceCred{}
	var committerName, committerEmail sql.NullString
	err := x.db.QueryRow(
		`SELECT
			id, workspace_id, name, type, committer_name, committer_email,
			username, password_or_token, created_at, updated_at
		FROM workspace_credentials
		WHERE id = ?`,
		id.String(),
	).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.Name,
		&c.Type,
		&committerName,
		&committerEmail,
		&c.Username,
		&c.PasswordOrToken,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if committerName.Valid {
		c.CommitterName = committerName.String
	}
	if committerEmail.Valid {
		c.CommitterEmail = committerEmail.String
	}
	if c.Type == "" {
		c.Type = CredentialTypeGit
	}
	return c, nil
}

// GetByIDAndWorkspace returns the credential for the given ID and workspace.
func (x *workspaceCredRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceCred, error) {
	c, err := x.GetByID(id)
	if err != nil || c == nil {
		return c, err
	}
	if c.WorkspaceID != workspaceID {
		return nil, nil
	}
	return c, nil
}

// Update updates a credential.
func (x *workspaceCredRepoSQLite) Update(c *WorkspaceCred) error {
	if c.Type == "" {
		c.Type = CredentialTypeGit
	}
	_, err := x.db.Exec(
		`UPDATE workspace_credentials
		SET
			name = ?,
			type = ?,
			committer_name = ?,
			committer_email = ?,
			username = ?,
			password_or_token = ?,
			updated_at = ?
		WHERE id = ?`,
		c.Name,
		c.Type,
		nullString(c.CommitterName),
		nullString(c.CommitterEmail),
		c.Username,
		c.PasswordOrToken,
		time.Now().UTC(),
		c.ID.String(),
	)
	return err
}

// Delete removes a credential.
func (x *workspaceCredRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_credentials WHERE id = ?`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID returns credentials for the given workspace with pagination.
func (x *workspaceCredRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceCred, error) {
	rows, err := x.db.Query(
		`SELECT
			id, workspace_id, name, type, committer_name, committer_email,
			username, password_or_token, created_at, updated_at
		FROM workspace_credentials
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
	return scanWorkspaceCredRows(rows)
}

// CountByWorkspaceID returns the number of credentials in the workspace.
func (x *workspaceCredRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_credentials
		WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

type workspaceCredRepoPostgres struct {
	db *sql.DB
}

// Create creates a new credential
func (x *workspaceCredRepoPostgres) Create(c *WorkspaceCred) error {
	if c.Type == "" {
		c.Type = CredentialTypeGit
	}
	err := x.db.QueryRow(
		`INSERT INTO workspace_credentials
		(workspace_id, name, type, committer_name, committer_email, username, password_or_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		c.WorkspaceID.String(),
		c.Name,
		c.Type,
		nullString(c.CommitterName),
		nullString(c.CommitterEmail),
		c.Username,
		c.PasswordOrToken,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return err
}

// GetByID returns the credential for the given ID.
func (x *workspaceCredRepoPostgres) GetByID(id ID) (*WorkspaceCred, error) {
	c := &WorkspaceCred{}
	var committerName, committerEmail sql.NullString
	err := x.db.QueryRow(
		`SELECT
			id, workspace_id, name, type, committer_name, committer_email,
			username, password_or_token, created_at, updated_at
		FROM workspace_credentials
		WHERE id = $1`,
		id.String(),
	).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.Name,
		&c.Type,
		&committerName,
		&committerEmail,
		&c.Username,
		&c.PasswordOrToken,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if committerName.Valid {
		c.CommitterName = committerName.String
	}
	if committerEmail.Valid {
		c.CommitterEmail = committerEmail.String
	}
	if c.Type == "" {
		c.Type = CredentialTypeGit
	}
	return c, nil
}

// GetByIDAndWorkspace returns the credential for the given ID and workspace.
func (x *workspaceCredRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceCred, error) {
	c, err := x.GetByID(id)
	if err != nil || c == nil {
		return c, err
	}
	if c.WorkspaceID != workspaceID {
		return nil, nil
	}
	return c, nil
}

// Update updates a credential.
func (x *workspaceCredRepoPostgres) Update(c *WorkspaceCred) error {
	if c.Type == "" {
		c.Type = CredentialTypeGit
	}
	_, err := x.db.Exec(
		`UPDATE workspace_credentials
		SET
			name = $1,
			type = $2,
			committer_name = $3,
			committer_email = $4,
			username = $5,
			password_or_token = $6,
			updated_at = $7
		WHERE id = $8`,
		c.Name,
		c.Type,
		nullString(c.CommitterName),
		nullString(c.CommitterEmail),
		c.Username,
		c.PasswordOrToken,
		time.Now().UTC(),
		c.ID.String(),
	)
	return err
}

// Delete removes a credential.
func (x *workspaceCredRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_credentials WHERE id = $1`,
		id.String(),
	)
	return err
}

// ListByWorkspaceID returns credentials for the given workspace with pagination.
func (x *workspaceCredRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceCred, error) {
	rows, err := x.db.Query(
		`SELECT
			id, workspace_id, name, type, committer_name, committer_email,
			username, password_or_token, created_at, updated_at
		FROM workspace_credentials
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
	return scanWorkspaceCredRows(rows)
}

// CountByWorkspaceID returns the number of credentials in the workspace.
func (x *workspaceCredRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_credentials
		WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
