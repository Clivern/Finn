// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"database/sql"

	"github.com/clivern/finn/db"
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

	CanGetWorkspaceRepositories  = "CAN_GET_WORKSPACE_REPOSITORIES"
	CanAddWorkspaceRepository    = "CAN_ADD_WORKSPACE_REPOSITORY"
	CanGetWorkspaceRepository    = "CAN_GET_WORKSPACE_REPOSITORY"
	CanUpdateWorkspaceRepository = "CAN_UPDATE_WORKSPACE_REPOSITORY"
	CanDeleteWorkspaceRepository = "CAN_DELETE_WORKSPACE_REPOSITORY"

	CanGetWorkspaceSSHKeys   = "CAN_GET_WORKSPACE_SSH_KEYS"
	CanAddWorkspaceSSHKey    = "CAN_ADD_WORKSPACE_SSH_KEY"
	CanGetWorkspaceSSHKey    = "CAN_GET_WORKSPACE_SSH_KEY"
	CanUpdateWorkspaceSSHKey = "CAN_UPDATE_WORKSPACE_SSH_KEY"
	CanDeleteWorkspaceSSHKey = "CAN_DELETE_WORKSPACE_SSH_KEY"

	CanGetWorkspaceGitCreds   = "CAN_GET_WORKSPACE_GIT_CREDS"
	CanAddWorkspaceGitCred    = "CAN_ADD_WORKSPACE_GIT_CRED"
	CanGetWorkspaceGitCred    = "CAN_GET_WORKSPACE_GIT_CRED"
	CanUpdateWorkspaceGitCred = "CAN_UPDATE_WORKSPACE_GIT_CRED"
	CanDeleteWorkspaceGitCred = "CAN_DELETE_WORKSPACE_GIT_CRED"
)

// Permission checks user permissions for resources (e.g. workspace).
type Permission struct {
	UserRepository             db.UserRepository
	WorkspaceRepository        db.WorkspaceRepository
	WorkspaceRepoRepository    db.WorkspaceRepoRepository
	WorkspaceSSHKeyRepository  db.WorkspaceSSHKeyRepository
	WorkspaceGitCredRepository db.WorkspaceGitCredRepository
}

// NewPermission creates a new permission module.
func NewPermission(dbConn *sql.DB) *Permission {
	return &Permission{
		UserRepository:             db.NewUserRepository(dbConn),
		WorkspaceRepository:        db.NewWorkspaceRepository(dbConn),
		WorkspaceRepoRepository:    db.NewWorkspaceRepoRepository(dbConn),
		WorkspaceSSHKeyRepository:  db.NewWorkspaceSSHKeyRepository(dbConn),
		WorkspaceGitCredRepository: db.NewWorkspaceGitCredRepository(dbConn),
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

// CheckWorkspaceRepositoryPermission checks if the user has the given permission on the workspace repository.
func (p *Permission) CheckWorkspaceRepositoryPermission(user *db.User, workspaceRepo *db.WorkspaceRepo, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(workspaceRepo.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceRepository, CanGetWorkspaceRepositories:
		return true, nil
	case CanUpdateWorkspaceRepository, CanDeleteWorkspaceRepository, CanAddWorkspaceRepository:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceSSHKeyPermission checks if the user has the given permission on the workspace SSH key.
func (p *Permission) CheckWorkspaceSSHKeyPermission(user *db.User, workspaceSSHKey *db.WorkspaceSSHKey, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(workspaceSSHKey.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceSSHKey, CanGetWorkspaceSSHKeys:
		return true, nil
	case CanUpdateWorkspaceSSHKey, CanDeleteWorkspaceSSHKey, CanAddWorkspaceSSHKey:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceGitCredPermission checks if the user has the given permission on the workspace Git credential.
func (p *Permission) CheckWorkspaceGitCredPermission(user *db.User, workspaceGitCred *db.WorkspaceGitCred, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(workspaceGitCred.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceGitCred, CanGetWorkspaceGitCreds:
		return true, nil
	case CanUpdateWorkspaceGitCred, CanDeleteWorkspaceGitCred, CanAddWorkspaceGitCred:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}
