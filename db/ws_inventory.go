// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

const InventoryTypeStatic = "static"
const InventoryTypeFile = "file"
const InventoryTypeStaticYAML = "static_yaml"

// WorkspaceInventory is a single row in the workspace_inventories table.
type WorkspaceInventory struct {
	ID                 ID
	WorkspaceID        ID
	Name               string
	Inventory          string
	InventoryPath      string
	Type               string
	SSHKeyID           *ID
	CredentialID       *ID
	BecomeKeyID        *ID
	BecomeCredentialID *ID
	RepositoryID       *ID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// WorkspaceInventoryRepository is the interface for workspace inventory CRUD.
type WorkspaceInventoryRepository interface {
	Create(i *WorkspaceInventory) error
	GetByID(id ID) (*WorkspaceInventory, error)
	GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceInventory, error)
	Update(i *WorkspaceInventory) error
	Delete(id ID) error
	ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceInventory, error)
	CountByWorkspaceID(workspaceID ID) (int64, error)
}

// NewWorkspaceInventoryRepository returns the repository for the current driver.
func NewWorkspaceInventoryRepository(db *sql.DB) WorkspaceInventoryRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &workspaceInventoryRepoPostgres{db: db}
	}
	return &workspaceInventoryRepoSQLite{db: db}
}

func scanWorkspaceInventoryRows(rows *sql.Rows) ([]*WorkspaceInventory, error) {
	var list []*WorkspaceInventory
	for rows.Next() {
		i := &WorkspaceInventory{}
		var inventory, inventoryPath sql.NullString
		var sshKeyID, credentialID, becomeKeyID, becomeCredentialID, repositoryID interface{}

		if err := rows.Scan(
			&i.ID,
			&i.WorkspaceID,
			&i.Name,
			&inventory,
			&inventoryPath,
			&i.Type,
			&sshKeyID,
			&credentialID,
			&becomeKeyID,
			&becomeCredentialID,
			&repositoryID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if inventory.Valid {
			i.Inventory = inventory.String
		}
		if inventoryPath.Valid {
			i.InventoryPath = inventoryPath.String
		}
		if i.Type == "" {
			i.Type = InventoryTypeStatic
		}

		if sshKeyID != nil {
			id := ID("")
			_ = id.Scan(sshKeyID)
			i.SSHKeyID = &id
		}
		if credentialID != nil {
			id := ID("")
			_ = id.Scan(credentialID)
			i.CredentialID = &id
		}
		if becomeKeyID != nil {
			id := ID("")
			_ = id.Scan(becomeKeyID)
			i.BecomeKeyID = &id
		}
		if becomeCredentialID != nil {
			id := ID("")
			_ = id.Scan(becomeCredentialID)
			i.BecomeCredentialID = &id
		}
		if repositoryID != nil {
			id := ID("")
			_ = id.Scan(repositoryID)
			i.RepositoryID = &id
		}

		list = append(list, i)
	}
	return list, rows.Err()
}

type workspaceInventoryRepoSQLite struct {
	db *sql.DB
}

func (x *workspaceInventoryRepoSQLite) Create(i *WorkspaceInventory) error {
	if i.Type == "" {
		i.Type = InventoryTypeStatic
	}

	result, err := x.db.Exec(
		`INSERT INTO workspace_inventories
		(
			workspace_id,
			name,
			inventory,
			inventory_path,
			type,
			ssh_key_id,
			credential_id,
			become_key_id,
			become_credential_id,
			repository_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.WorkspaceID.String(),
		i.Name,
		nullString(i.Inventory),
		nullString(i.InventoryPath),
		i.Type,
		i.SSHKeyID.ValueOrNilPtr(),
		i.CredentialID.ValueOrNilPtr(),
		i.BecomeKeyID.ValueOrNilPtr(),
		i.BecomeCredentialID.ValueOrNilPtr(),
		i.RepositoryID.ValueOrNilPtr(),
	)
	if err != nil {
		return err
	}

	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	i.ID = ID(strconv.FormatInt(lid, 10))

	row := x.db.QueryRow(
		`SELECT created_at, updated_at
		FROM workspace_inventories
		WHERE id = ?`,
		i.ID.String(),
	)
	_ = row.Scan(&i.CreatedAt, &i.UpdatedAt)
	return nil
}

func (x *workspaceInventoryRepoSQLite) GetByID(id ID) (*WorkspaceInventory, error) {
	i := &WorkspaceInventory{}
	var inventory, inventoryPath sql.NullString
	var sshKeyID, credentialID, becomeKeyID, becomeCredentialID, repositoryID interface{}

	err := x.db.QueryRow(
		`SELECT
			id,
			workspace_id,
			name,
			inventory,
			inventory_path,
			type,
			ssh_key_id,
			credential_id,
			become_key_id,
			become_credential_id,
			repository_id,
			created_at,
			updated_at
		FROM workspace_inventories
		WHERE id = ?`,
		id.String(),
	).Scan(
		&i.ID,
		&i.WorkspaceID,
		&i.Name,
		&inventory,
		&inventoryPath,
		&i.Type,
		&sshKeyID,
		&credentialID,
		&becomeKeyID,
		&becomeCredentialID,
		&repositoryID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if inventory.Valid {
		i.Inventory = inventory.String
	}
	if inventoryPath.Valid {
		i.InventoryPath = inventoryPath.String
	}
	if i.Type == "" {
		i.Type = InventoryTypeStatic
	}

	if sshKeyID != nil {
		idVal := ID("")
		_ = idVal.Scan(sshKeyID)
		i.SSHKeyID = &idVal
	}
	if credentialID != nil {
		idVal := ID("")
		_ = idVal.Scan(credentialID)
		i.CredentialID = &idVal
	}
	if becomeKeyID != nil {
		idVal := ID("")
		_ = idVal.Scan(becomeKeyID)
		i.BecomeKeyID = &idVal
	}
	if becomeCredentialID != nil {
		idVal := ID("")
		_ = idVal.Scan(becomeCredentialID)
		i.BecomeCredentialID = &idVal
	}
	if repositoryID != nil {
		idVal := ID("")
		_ = idVal.Scan(repositoryID)
		i.RepositoryID = &idVal
	}

	return i, nil
}

func (x *workspaceInventoryRepoSQLite) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceInventory, error) {
	i, err := x.GetByID(id)
	if err != nil || i == nil {
		return i, err
	}
	if i.WorkspaceID != workspaceID {
		return nil, nil
	}
	return i, nil
}

func (x *workspaceInventoryRepoSQLite) Update(i *WorkspaceInventory) error {
	if i.Type == "" {
		i.Type = InventoryTypeStatic
	}

	_, err := x.db.Exec(
		`UPDATE workspace_inventories
		SET
			name = ?,
			inventory = ?,
			inventory_path = ?,
			type = ?,
			ssh_key_id = ?,
			credential_id = ?,
			become_key_id = ?,
			become_credential_id = ?,
			repository_id = ?,
			updated_at = ?
		WHERE id = ?`,
		i.Name,
		nullString(i.Inventory),
		nullString(i.InventoryPath),
		i.Type,
		i.SSHKeyID.ValueOrNilPtr(),
		i.CredentialID.ValueOrNilPtr(),
		i.BecomeKeyID.ValueOrNilPtr(),
		i.BecomeCredentialID.ValueOrNilPtr(),
		i.RepositoryID.ValueOrNilPtr(),
		time.Now().UTC(),
		i.ID.String(),
	)
	return err
}

func (x *workspaceInventoryRepoSQLite) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_inventories WHERE id = ?`,
		id.String(),
	)
	return err
}

func (x *workspaceInventoryRepoSQLite) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceInventory, error) {
	rows, err := x.db.Query(
		`SELECT
			id,
			workspace_id,
			name,
			inventory,
			inventory_path,
			type,
			ssh_key_id,
			credential_id,
			become_key_id,
			become_credential_id,
			repository_id,
			created_at,
			updated_at
		FROM workspace_inventories
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
	return scanWorkspaceInventoryRows(rows)
}

func (x *workspaceInventoryRepoSQLite) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_inventories
		WHERE workspace_id = ?`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}

type workspaceInventoryRepoPostgres struct {
	db *sql.DB
}

func (x *workspaceInventoryRepoPostgres) Create(i *WorkspaceInventory) error {
	if i.Type == "" {
		i.Type = InventoryTypeStatic
	}

	err := x.db.QueryRow(
		`INSERT INTO workspace_inventories
		(
			workspace_id,
			name,
			inventory,
			inventory_path,
			type,
			ssh_key_id,
			credential_id,
			become_key_id,
			become_credential_id,
			repository_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`,
		i.WorkspaceID.String(),
		i.Name,
		nullString(i.Inventory),
		nullString(i.InventoryPath),
		i.Type,
		i.SSHKeyID.ValueOrNilPtr(),
		i.CredentialID.ValueOrNilPtr(),
		i.BecomeKeyID.ValueOrNilPtr(),
		i.BecomeCredentialID.ValueOrNilPtr(),
		i.RepositoryID.ValueOrNilPtr(),
	).Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)

	return err
}

func (x *workspaceInventoryRepoPostgres) GetByID(id ID) (*WorkspaceInventory, error) {
	i := &WorkspaceInventory{}
	var inventory, inventoryPath sql.NullString
	var sshKeyID, credentialID, becomeKeyID, becomeCredentialID, repositoryID interface{}

	err := x.db.QueryRow(
		`SELECT
			id,
			workspace_id,
			name,
			inventory,
			inventory_path,
			type,
			ssh_key_id,
			credential_id,
			become_key_id,
			become_credential_id,
			repository_id,
			created_at,
			updated_at
		FROM workspace_inventories
		WHERE id = $1`,
		id.String(),
	).Scan(
		&i.ID,
		&i.WorkspaceID,
		&i.Name,
		&inventory,
		&inventoryPath,
		&i.Type,
		&sshKeyID,
		&credentialID,
		&becomeKeyID,
		&becomeCredentialID,
		&repositoryID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if inventory.Valid {
		i.Inventory = inventory.String
	}
	if inventoryPath.Valid {
		i.InventoryPath = inventoryPath.String
	}
	if i.Type == "" {
		i.Type = InventoryTypeStatic
	}

	if sshKeyID != nil {
		idVal := ID("")
		_ = idVal.Scan(sshKeyID)
		i.SSHKeyID = &idVal
	}
	if credentialID != nil {
		idVal := ID("")
		_ = idVal.Scan(credentialID)
		i.CredentialID = &idVal
	}
	if becomeKeyID != nil {
		idVal := ID("")
		_ = idVal.Scan(becomeKeyID)
		i.BecomeKeyID = &idVal
	}
	if becomeCredentialID != nil {
		idVal := ID("")
		_ = idVal.Scan(becomeCredentialID)
		i.BecomeCredentialID = &idVal
	}
	if repositoryID != nil {
		idVal := ID("")
		_ = idVal.Scan(repositoryID)
		i.RepositoryID = &idVal
	}

	return i, nil
}

func (x *workspaceInventoryRepoPostgres) GetByIDAndWorkspace(id, workspaceID ID) (*WorkspaceInventory, error) {
	i, err := x.GetByID(id)
	if err != nil || i == nil {
		return i, err
	}
	if i.WorkspaceID != workspaceID {
		return nil, nil
	}
	return i, nil
}

func (x *workspaceInventoryRepoPostgres) Update(i *WorkspaceInventory) error {
	if i.Type == "" {
		i.Type = InventoryTypeStatic
	}

	_, err := x.db.Exec(
		`UPDATE workspace_inventories
		SET
			name = $1,
			inventory = $2,
			inventory_path = $3,
			type = $4,
			ssh_key_id = $5,
			credential_id = $6,
			become_key_id = $7,
			become_credential_id = $8,
			repository_id = $9,
			updated_at = $10
		WHERE id = $11`,
		i.Name,
		nullString(i.Inventory),
		nullString(i.InventoryPath),
		i.Type,
		i.SSHKeyID.ValueOrNilPtr(),
		i.CredentialID.ValueOrNilPtr(),
		i.BecomeKeyID.ValueOrNilPtr(),
		i.BecomeCredentialID.ValueOrNilPtr(),
		i.RepositoryID.ValueOrNilPtr(),
		time.Now().UTC(),
		i.ID.String(),
	)
	return err
}

func (x *workspaceInventoryRepoPostgres) Delete(id ID) error {
	_, err := x.db.Exec(
		`DELETE FROM workspace_inventories WHERE id = $1`,
		id.String(),
	)
	return err
}

func (x *workspaceInventoryRepoPostgres) ListByWorkspaceID(workspaceID ID, limit, offset int) ([]*WorkspaceInventory, error) {
	rows, err := x.db.Query(
		`SELECT
			id,
			workspace_id,
			name,
			inventory,
			inventory_path,
			type,
			ssh_key_id,
			credential_id,
			become_key_id,
			become_credential_id,
			repository_id,
			created_at,
			updated_at
		FROM workspace_inventories
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
	return scanWorkspaceInventoryRows(rows)
}

func (x *workspaceInventoryRepoPostgres) CountByWorkspaceID(workspaceID ID) (int64, error) {
	var count int64
	err := x.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_inventories
		WHERE workspace_id = $1`,
		workspaceID.String(),
	).Scan(&count)
	return count, err
}
