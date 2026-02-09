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
	CanCloneWorkspaceRepository  = "CAN_CLONE_WORKSPACE_REPOSITORY"
	CanUpdateWorkspaceRepository = "CAN_UPDATE_WORKSPACE_REPOSITORY"
	CanDeleteWorkspaceRepository = "CAN_DELETE_WORKSPACE_REPOSITORY"

	CanGetWorkspaceSSHKeys   = "CAN_GET_WORKSPACE_SSH_KEYS"
	CanAddWorkspaceSSHKey    = "CAN_ADD_WORKSPACE_SSH_KEY"
	CanGetWorkspaceSSHKey    = "CAN_GET_WORKSPACE_SSH_KEY"
	CanUpdateWorkspaceSSHKey = "CAN_UPDATE_WORKSPACE_SSH_KEY"
	CanDeleteWorkspaceSSHKey = "CAN_DELETE_WORKSPACE_SSH_KEY"

	CanGetWorkspaceCreds   = "CAN_GET_WORKSPACE_CREDS"
	CanAddWorkspaceCred    = "CAN_ADD_WORKSPACE_CRED"
	CanGetWorkspaceCred    = "CAN_GET_WORKSPACE_CRED"
	CanUpdateWorkspaceCred = "CAN_UPDATE_WORKSPACE_CRED"
	CanDeleteWorkspaceCred = "CAN_DELETE_WORKSPACE_CRED"

	CanGetWorkspaceBackends   = "CAN_GET_WORKSPACE_BACKENDS"
	CanAddWorkspaceBackend    = "CAN_ADD_WORKSPACE_BACKEND"
	CanGetWorkspaceBackend    = "CAN_GET_WORKSPACE_BACKEND"
	CanUpdateWorkspaceBackend = "CAN_UPDATE_WORKSPACE_BACKEND"
	CanDeleteWorkspaceBackend = "CAN_DELETE_WORKSPACE_BACKEND"

	CanGetWorkspaceVariableGroups   = "CAN_GET_WORKSPACE_VARIABLE_GROUPS"
	CanAddWorkspaceVariableGroup    = "CAN_ADD_WORKSPACE_VARIABLE_GROUP"
	CanGetWorkspaceVariableGroup    = "CAN_GET_WORKSPACE_VARIABLE_GROUP"
	CanUpdateWorkspaceVariableGroup = "CAN_UPDATE_WORKSPACE_VARIABLE_GROUP"
	CanDeleteWorkspaceVariableGroup = "CAN_DELETE_WORKSPACE_VARIABLE_GROUP"

	CanGetWorkspaceInventories  = "CAN_GET_WORKSPACE_INVENTORIES"
	CanAddWorkspaceInventory    = "CAN_ADD_WORKSPACE_INVENTORY"
	CanGetWorkspaceInventory    = "CAN_GET_WORKSPACE_INVENTORY"
	CanUpdateWorkspaceInventory = "CAN_UPDATE_WORKSPACE_INVENTORY"
	CanDeleteWorkspaceInventory = "CAN_DELETE_WORKSPACE_INVENTORY"

	CanGetWorkspaceTemplates   = "CAN_GET_WORKSPACE_TEMPLATES"
	CanAddWorkspaceTemplate    = "CAN_ADD_WORKSPACE_TEMPLATE"
	CanGetWorkspaceTemplate    = "CAN_GET_WORKSPACE_TEMPLATE"
	CanUpdateWorkspaceTemplate = "CAN_UPDATE_WORKSPACE_TEMPLATE"
	CanDeleteWorkspaceTemplate = "CAN_DELETE_WORKSPACE_TEMPLATE"
)

// Permission checks user permissions for resources (e.g. workspace).
type Permission struct {
	UserRepository                   db.UserRepository
	WorkspaceRepository              db.WorkspaceRepository
	WorkspaceRepoRepository          db.WorkspaceRepoRepository
	WorkspaceSSHKeyRepository        db.WorkspaceSSHKeyRepository
	WorkspaceCredRepository          db.WorkspaceCredRepository
	WorkspaceInventoryRepository     db.WorkspaceInventoryRepository
	WorkspaceBackendRepository       db.WorkspaceBackendRepository
	WorkspaceVariableGroupRepository db.WorkspaceVariableGroupRepository
}

// NewPermission creates a new permission module.
func NewPermission(dbConn *sql.DB) *Permission {
	return &Permission{
		UserRepository:                   db.NewUserRepository(dbConn),
		WorkspaceRepository:              db.NewWorkspaceRepository(dbConn),
		WorkspaceRepoRepository:          db.NewWorkspaceRepoRepository(dbConn),
		WorkspaceSSHKeyRepository:        db.NewWorkspaceSSHKeyRepository(dbConn),
		WorkspaceCredRepository:          db.NewWorkspaceCredRepository(dbConn),
		WorkspaceInventoryRepository:     db.NewWorkspaceInventoryRepository(dbConn),
		WorkspaceBackendRepository:       db.NewWorkspaceBackendRepository(dbConn),
		WorkspaceVariableGroupRepository: db.NewWorkspaceVariableGroupRepository(dbConn),
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
	case CanUpdateWorkspaceRepository, CanDeleteWorkspaceRepository, CanAddWorkspaceRepository, CanCloneWorkspaceRepository:
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

// CheckWorkspaceCredPermission checks if the user has the given permission on the workspace credential.
func (p *Permission) CheckWorkspaceCredPermission(user *db.User, cred *db.WorkspaceCred, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(cred.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceCred, CanGetWorkspaceCreds:
		return true, nil
	case CanUpdateWorkspaceCred, CanDeleteWorkspaceCred, CanAddWorkspaceCred:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceBackendPermission checks if the user has the given permission on the workspace backend.
func (p *Permission) CheckWorkspaceBackendPermission(user *db.User, backend *db.WorkspaceBackend, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(backend.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceBackend, CanGetWorkspaceBackends:
		return true, nil
	case CanUpdateWorkspaceBackend, CanDeleteWorkspaceBackend, CanAddWorkspaceBackend:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceVariableGroupPermission checks if the user has the given permission on the workspace variable group.
func (p *Permission) CheckWorkspaceVariableGroupPermission(user *db.User, group *db.WorkspaceVariableGroup, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(group.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceVariableGroup, CanGetWorkspaceVariableGroups:
		return true, nil
	case CanUpdateWorkspaceVariableGroup, CanDeleteWorkspaceVariableGroup, CanAddWorkspaceVariableGroup:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceInventoryPermission checks if the user has the given permission on the workspace inventory.
func (p *Permission) CheckWorkspaceInventoryPermission(user *db.User, inventory *db.WorkspaceInventory, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(inventory.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceInventory, CanGetWorkspaceInventories:
		return true, nil
	case CanUpdateWorkspaceInventory, CanDeleteWorkspaceInventory, CanAddWorkspaceInventory:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CheckWorkspaceTemplatePermission checks if the user has the given permission on the workspace template.
func (p *Permission) CheckWorkspaceTemplatePermission(user *db.User, tmpl *db.WorkspaceTemplate, permission string) (bool, error) {
	if user.Role == db.UserRoleAdmin {
		return true, nil
	}

	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(tmpl.WorkspaceID, user.ID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspaceTemplate, CanGetWorkspaceTemplates:
		return true, nil
	case CanUpdateWorkspaceTemplate, CanDeleteWorkspaceTemplate, CanAddWorkspaceTemplate:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}
