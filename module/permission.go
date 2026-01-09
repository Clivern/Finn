// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"database/sql"

	"github.com/clivern/tivo/db"
)

// Workspace permission constants. Use with Permission.Has for workspace-scoped checks.
const (
	CanGetWorkspace    = "CAN_GET_WORKSPACE"
	CanUpdateWorkspace = "CAN_UPDATE_WORKSPACE"
	CanDeleteWorkspace = "CAN_DELETE_WORKSPACE"

	CanGetWorkspaceUsers   = "CAN_GET_WORKSPACE_USERS"
	CanAddWorkspaceUser    = "CAN_ADD_WORKSPACE_USER"
	CanGetWorkspaceUser    = "CAN_GET_WORKSPACE_USER"
	CanUpdateWorkspaceUser = "CAN_UPDATE_WORKSPACE_USER"
	CanDeleteWorkspaceUser = "CAN_DELETE_WORKSPACE_USER"
)

// Permission checks user permissions for resources (e.g. workspace).
type Permission struct {
	UserRepository      db.UserRepository
	WorkspaceRepository db.WorkspaceRepository
}

// NewPermission creates a new permission module.
func NewPermission(dbConn *sql.DB) *Permission {
	return &Permission{
		UserRepository:      db.NewUserRepository(dbConn),
		WorkspaceRepository: db.NewWorkspaceRepository(dbConn),
	}
}

// CheckWorkspacePermission checks if the user has the given permission on the workspace.
func (p *Permission) CheckWorkspacePermission(user *db.User, workspace *db.Workspace, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(workspace.ID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspace:
		return true, nil
	case CanUpdateWorkspace, CanDeleteWorkspace:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceUserPermission checks if the user has the given permission on the workspace user.
func (p *Permission) CheckWorkspaceUserPermission(user *db.User, workspaceUser *db.WorkspaceUser, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(workspaceUser.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceUser, CanGetWorkspaceUsers:
		return true, nil
	case CanUpdateWorkspaceUser, CanDeleteWorkspaceUser, CanAddWorkspaceUser:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}
