// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"database/sql"
	"errors"

	"github.com/clivern/lapin/db"
)

var (
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrWorkspaceUserNotFound = errors.New("workspace user not found")
)

// Workspace handles workspace management operations.
type Workspace struct {
	WorkspaceRepository     db.WorkspaceRepository
	WorkspaceUserRepository db.WorkspaceUserRepository
}

// NewWorkspace creates a new workspace module instance using db.NewWorkspaceRepository(db).
func NewWorkspace(dbConn *sql.DB) *Workspace {
	return &Workspace{WorkspaceRepository: db.NewWorkspaceRepository(dbConn), WorkspaceUserRepository: db.NewWorkspaceUserRepository(dbConn)}
}

// CreateWorkspaceOptions contains options for creating a workspace.
type CreateWorkspaceOptions struct {
	Name string
}

// CreateWorkspace creates a new workspace.
func (w *Workspace) CreateWorkspace(options *CreateWorkspaceOptions) (*db.Workspace, error) {
	workspace := &db.Workspace{Name: options.Name}
	if err := w.WorkspaceRepository.Create(workspace); err != nil {
		return nil, err
	}
	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID.
func (w *Workspace) GetWorkspace(workspaceID db.ID) (*db.Workspace, error) {
	workspace, err := w.WorkspaceRepository.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}
	return workspace, nil
}

// GetWorkspaceMembership returns the user's membership for the workspace, or nil if not a member.
func (w *Workspace) GetWorkspaceMembership(workspaceID, userID db.ID) (*db.WorkspaceUser, error) {
	return w.WorkspaceRepository.GetWorkspaceMembership(workspaceID, userID)
}

// CountWorkspaceMembers returns the number of members in a workspace.
func (w *Workspace) CountWorkspaceMembers(workspaceID db.ID) (int64, error) {
	return w.WorkspaceUserRepository.CountByWorkspace(workspaceID)
}

// ListWorkspacesOptions contains options for listing workspaces.
// UserID: nil or empty = all workspaces; set = only workspaces the user is attached to.
type ListWorkspacesOptions struct {
	Limit  int
	Offset int
	UserID *db.ID
}

// ListWorkspacesResult contains the result of listing workspaces.
type ListWorkspacesResult struct {
	Workspaces []*db.Workspace
	Total      int64
}

// ListWorkspaces retrieves workspaces with pagination. Pass UserID nil for all, or a user ID to filter by membership.
func (w *Workspace) ListWorkspaces(options *ListWorkspacesOptions) (*ListWorkspacesResult, error) {
	list, err := w.WorkspaceRepository.List(options.Limit, options.Offset, options.UserID)
	if err != nil {
		return nil, err
	}
	total, err := w.WorkspaceRepository.Count(options.UserID)
	if err != nil {
		return nil, err
	}
	return &ListWorkspacesResult{Workspaces: list, Total: total}, nil
}

// UpdateWorkspaceOptions contains options for updating a workspace.
type UpdateWorkspaceOptions struct {
	WorkspaceID db.ID
	Name        string
}

// UpdateWorkspace updates an existing workspace.
func (w *Workspace) UpdateWorkspace(options *UpdateWorkspaceOptions) (*db.Workspace, error) {
	workspace, err := w.WorkspaceRepository.GetByID(options.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}
	workspace.Name = options.Name
	if err := w.WorkspaceRepository.Update(workspace); err != nil {
		return nil, err
	}
	return workspace, nil
}

// DeleteWorkspace deletes a workspace by ID.
func (w *Workspace) DeleteWorkspace(workspaceID db.ID) error {
	workspace, err := w.WorkspaceRepository.GetByID(workspaceID)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}
	return w.WorkspaceRepository.Delete(workspaceID)
}

// AddWorkspaceUser adds a user to a workspace with the given role.
func (w *Workspace) AddWorkspaceUser(workspaceID, userID db.ID, role string) (*db.WorkspaceUser, error) {
	workspaceUser := &db.WorkspaceUser{WorkspaceID: workspaceID, UserID: userID, Role: role}
	if err := w.WorkspaceUserRepository.Create(workspaceUser); err != nil {
		return nil, err
	}
	return workspaceUser, nil
}

// ListWorkspaceUsersOptions contains options for listing workspace users
type ListWorkspaceUsersOptions struct {
	WorkspaceID db.ID
	Limit       int
	Offset      int
}

// ListWorkspaceUsersResult contains the result of listing workspace users
type ListWorkspaceUsersResult struct {
	WorkspaceUsers []*db.WorkspaceUser
	Total          int64
}

// ListWorkspaceUsers returns users in a workspace with pagination
func (w *Workspace) ListWorkspaceUsers(options *ListWorkspaceUsersOptions) (*ListWorkspaceUsersResult, error) {
	list, err := w.WorkspaceUserRepository.ListByWorkspace(options.WorkspaceID, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	total, err := w.WorkspaceUserRepository.CountByWorkspace(options.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return &ListWorkspaceUsersResult{WorkspaceUsers: list, Total: total}, nil
}

// GetWorkspaceUser returns the membership for a user in a workspace, or nil if not a member
func (w *Workspace) GetWorkspaceUser(workspaceID, userID db.ID) (*db.WorkspaceUser, error) {
	return w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceID, userID)
}

// UpdateWorkspaceUserRoleOptions contains options for updating a workspace user role
type UpdateWorkspaceUserRoleOptions struct {
	WorkspaceID db.ID
	UserID      db.ID
	Role        string
}

// UpdateWorkspaceUserRole updates the role of a user in a workspace
func (w *Workspace) UpdateWorkspaceUserRole(options *UpdateWorkspaceUserRoleOptions) (*db.WorkspaceUser, error) {
	m, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(options.WorkspaceID, options.UserID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrWorkspaceUserNotFound
	}
	m.Role = options.Role
	if err := w.WorkspaceUserRepository.Update(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoveWorkspaceUser removes a user from a workspace
func (w *Workspace) RemoveWorkspaceUser(workspaceID, userID db.ID) error {
	m, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceID, userID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrWorkspaceUserNotFound
	}
	return w.WorkspaceUserRepository.Delete(m.ID)
}
