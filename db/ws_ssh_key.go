// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// WorkspaceSSHKey is a single row in the workspace_ssh_keys table.
type WorkspaceSSHKey struct {
	ID             ID
	WorkspaceID    ID
	Name           string
	CommitterName  string
	CommitterEmail string
	PrivateKey     string
	PublicKey      *string
	Passphrase     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// WorkspaceSSHKeyRepository is the interface for workspace SSH key CRUD.
type WorkspaceSSHKeyRepository interface {
	Create(k *WorkspaceSSHKey) error
	GetByID(id ID) (*WorkspaceSSHKey, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceSSHKey, error)
	Update(k *WorkspaceSSHKey) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceSSHKey, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceSSHKeyRepository returns the repository for the current driver.
func NewWorkspaceSSHKeyRepository(db *sql.DB) WorkspaceSSHKeyRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceSSHKeyRepoPostgres{db: db}
	}
	return &workspaceSSHKeyRepoSQLite{db: db}
}

func scanWorkspaceSSHKeyRows(rows *sql.Rows) ([]*WorkspaceSSHKey, error) {
	var list []*WorkspaceSSHKey
	for rows.Next() {
		k := &WorkspaceSSHKey{}
		if err := rows.Scan(
			&k.ID,
			&k.WorkspaceID,
			&k.Name,
			&k.CommitterName,
			&k.CommitterEmail,
			&k.PrivateKey,
			&k.PublicKey,
			&k.Passphrase,
			&k.CreatedAt,
			&k.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, k)
	}
	return list, rows.Err()
}

type workspaceSSHKeyRepoSQLite struct {
	db *sql.DB
}

func (x *workspaceSSHKeyRepoSQLite) Create(k *WorkspaceSSHKey) error {
	result, err := x.db.Exec(
		`INSERT INTO workspace_ssh_keys (workspace_id, name, committer_name, committer_email, private_key, public_key, passphrase)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		k.WorkspaceID.String(),
		k.Name,
		k.CommitterName,
		k.CommitterEmail,
		k.PrivateKey,
		k.PublicKey,
		k.Passphrase,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	k.ID = ID(strconv.FormatInt(lid, 10))
	row := x.db.QueryRow(
		`SELECT created_at, updated_at FROM workspace_ssh_keys WHERE id = ?`,
		k.ID.String(),
	)
	_ = row.Scan(&k.CreatedAt, &k.UpdatedAt)
	return nil
}

func (x *workspaceSSHKeyRepoSQLite) GetByID(id ID) (*WorkspaceSSHKey, error) {
	k := &WorkspaceSSHKey{}
	err := x.db.QueryRow(
		`SELECT id, workspace_id, name, committer_name, committer_email, private_key, public_key, passphrase, created_at, updated_at
		FROM workspace_ssh_keys WHERE id = ?`,
		id.String(),
	).Scan(
		&k.ID,
		&k.WorkspaceID,
		&k.Name,
		&k.CommitterName,
		&k.CommitterEmail,
		&k.PrivateKey,
		&k.PublicKey,
		&k.Passphrase,
		&k.CreatedAt,
		&k.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return k, nil
}

func (x *workspaceSSHKeyRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceSSHKey, error) {
	k, err := x.GetByID(id)
	if err != nil || k == nil {
		return k, err
	}
	if k.WorkspaceID != workspaceID {
		return nil, nil
	}
	return k, nil
}

func (x *workspaceSSHKeyRepoSQLite) Update(k *WorkspaceSSHKey) error {
	_, err := x.db.Exec(
		`UPDATE workspace_ssh_keys SET name = ?, committer_name = ?, committer_email = ?, private_key = ?, public_key = ?, passphrase = ?, updated_at = ?
		WHERE id = ?`,
		k.Name,
		k.CommitterName,
		k.CommitterEmail,
		k.PrivateKey,
		k.PublicKey,
		k.Passphrase,
		time.Now().UTC(),
		k.ID.String(),
	)
	return err
}

func (x *workspaceSSHKeyRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_ssh_keys WHERE id = ?`,
		id.String(),
	)
	return err
}

func (x *workspaceSSHKeyRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceSSHKey, error) {
	rows, err := x.db.Query(
		`SELECT id, workspace_id, name, committer_name, committer_email, private_key, public_key, passphrase, created_at, updated_at
		FROM workspace_ssh_keys WHERE workspace_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceSSHKeyRows(rows)
}

func (x *workspaceSSHKeyRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_ssh_keys WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

type workspaceSSHKeyRepoPostgres struct {
	db *sql.DB
}

func (x *workspaceSSHKeyRepoPostgres) Create(k *WorkspaceSSHKey) error {
	err := x.db.QueryRow(
		`INSERT INTO workspace_ssh_keys (workspace_id, name, committer_name, committer_email, private_key, public_key, passphrase)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`,
		k.WorkspaceID.String(),
		k.Name,
		k.CommitterName,
		k.CommitterEmail,
		k.PrivateKey,
		k.PublicKey,
		k.Passphrase,
	).Scan(&k.ID, &k.CreatedAt, &k.UpdatedAt)
	return err
}

func (x *workspaceSSHKeyRepoPostgres) GetByID(id ID) (*WorkspaceSSHKey, error) {
	k := &WorkspaceSSHKey{}
	err := x.db.QueryRow(
		`SELECT id, workspace_id, name, committer_name, committer_email, private_key, public_key, passphrase, created_at, updated_at
		FROM workspace_ssh_keys WHERE id = $1`,
		id.String(),
	).Scan(
		&k.ID,
		&k.WorkspaceID,
		&k.Name,
		&k.CommitterName,
		&k.CommitterEmail,
		&k.PrivateKey,
		&k.PublicKey,
		&k.Passphrase,
		&k.CreatedAt,
		&k.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return k, nil
}

func (x *workspaceSSHKeyRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceSSHKey, error) {
	k, err := x.GetByID(id)
	if err != nil || k == nil {
		return k, err
	}
	if k.WorkspaceID != workspaceID {
		return nil, nil
	}
	return k, nil
}

func (x *workspaceSSHKeyRepoPostgres) Update(k *WorkspaceSSHKey) error {
	_, err := x.db.Exec(
		`UPDATE workspace_ssh_keys SET name = $1, committer_name = $2, committer_email = $3, private_key = $4, public_key = $5, passphrase = $6, updated_at = $7 WHERE id = $8`,
		k.Name,
		k.CommitterName,
		k.CommitterEmail,
		k.PrivateKey,
		k.PublicKey,
		k.Passphrase,
		time.Now().UTC(),
		k.ID.String(),
	)
	return err
}

func (x *workspaceSSHKeyRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_ssh_keys WHERE id = $1`,
		id.String(),
	)
	return err
}

func (x *workspaceSSHKeyRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceSSHKey, error) {
	rows, err := x.db.Query(
		`SELECT id, workspace_id, name, committer_name, committer_email, private_key, public_key, passphrase, created_at, updated_at
		FROM workspace_ssh_keys WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		workspaceID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkspaceSSHKeyRows(rows)
}

func (x *workspaceSSHKeyRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_ssh_keys WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
