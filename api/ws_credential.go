// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/clivern/lapin/db"
	"github.com/clivern/lapin/middleware"
	"github.com/clivern/lapin/module"
	"github.com/clivern/lapin/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceGitCredRequest is the request body for creating workspace git credentials.
type CreateWorkspaceGitCredRequest struct {
	Name            string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName   string `json:"committerName" validate:"required,max=100" label:"Committer name"`
	CommitterEmail  string `json:"committerEmail" validate:"required,max=100" label:"Committer email"`
	Username        string `json:"username" validate:"omitempty,max=100" label:"Username"`
	PasswordOrToken string `json:"passwordOrToken" validate:"required" label:"Password or token"`
}

// UpdateWorkspaceGitCredRequest is the request body for updating workspace git credentials.
type UpdateWorkspaceGitCredRequest struct {
	Name            string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName   string `json:"committerName" validate:"required,max=100" label:"Committer name"`
	CommitterEmail  string `json:"committerEmail" validate:"required,max=100" label:"Committer email"`
	Username        string `json:"username" validate:"omitempty,max=100" label:"Username"`
	PasswordOrToken string `json:"passwordOrToken" validate:"omitempty" label:"Password or token"`
}

// CreateWorkspaceGitCredAction creates git credentials in a workspace.
func CreateWorkspaceGitCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace git credential endpoint called")

	var req CreateWorkspaceGitCredRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceGitCredPermission(
		user,
		&db.WorkspaceGitCred{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceGitCred,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	cred := &db.WorkspaceGitCred{
		WorkspaceID:     workspace.ID,
		Name:            req.Name,
		CommitterName:   &req.CommitterName,
		CommitterEmail:  &req.CommitterEmail,
		PasswordOrToken: &req.PasswordOrToken,
	}
	if req.Username != "" {
		cred.Username = &req.Username
	}
	repo := db.NewWorkspaceGitCredRepository(db.GetDB())
	if err := repo.Create(cred); err != nil {
		log.Error().Err(err).Msg("Failed to create workspace git credentials")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create git credentials",
		})
		return
	}

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             cred.ID,
		"workspaceId":    cred.WorkspaceID,
		"name":           cred.Name,
		"committerName":  cred.CommitterName,
		"committerEmail": cred.CommitterEmail,
		"username":       cred.Username,
		"createdAt":      cred.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      cred.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceGitCredsAction lists git credentials in a workspace
func ListWorkspaceGitCredsAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace git credentials endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceGitCredPermission(
		user,
		&db.WorkspaceGitCred{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceGitCreds,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	limit, offset := 50, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	repo := db.NewWorkspaceGitCredRepository(db.GetDB())
	list, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspace git credentials")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list git credentials",
		})
		return
	}
	total, _ := repo.CountByWorkspaceID(workspace.ID)
	items := make([]map[string]interface{}, 0, len(list))
	for _, c := range list {
		items = append(items, map[string]interface{}{
			"id":             c.ID,
			"workspaceId":    c.WorkspaceID,
			"name":           c.Name,
			"committerName":  c.CommitterName,
			"committerEmail": c.CommitterEmail,
			"username":       c.Username,
			"createdAt":      c.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":      c.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"credentials": items,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceGitCredAction returns a single workspace git credential (password/token never returned).
func GetWorkspaceGitCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace git credential endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid git credentials ID",
		})
		return
	}

	repo := db.NewWorkspaceGitCredRepository(db.GetDB())
	c, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace git credentials")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get git credentials",
		})
		return
	}
	if c == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Git credentials not found",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceGitCredPermission(
		user,
		c,
		module.CanGetWorkspaceGitCred,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":             c.ID,
		"workspaceId":    c.WorkspaceID,
		"name":           c.Name,
		"committerName":  c.CommitterName,
		"committerEmail": c.CommitterEmail,
		"username":       c.Username,
		"createdAt":      c.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      c.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceGitCredAction updates workspace git credentials.
func UpdateWorkspaceGitCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace git credential endpoint called")

	var req UpdateWorkspaceGitCredRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid git credentials ID",
		})
		return
	}

	repo := db.NewWorkspaceGitCredRepository(db.GetDB())
	c, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || c == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Git credentials not found",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceGitCredPermission(
		user,
		c,
		module.CanUpdateWorkspaceGitCred,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	c.Name = req.Name
	c.CommitterName = &req.CommitterName
	c.CommitterEmail = &req.CommitterEmail
	if req.Username != "" {
		c.Username = &req.Username
	} else {
		c.Username = nil
	}
	if req.PasswordOrToken != "" {
		c.PasswordOrToken = &req.PasswordOrToken
	}
	if err := repo.Update(c); err != nil {
		log.Error().Err(err).Msg("Failed to update workspace git credentials")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update git credentials",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":             c.ID,
		"workspaceId":    c.WorkspaceID,
		"name":           c.Name,
		"committerName":  c.CommitterName,
		"committerEmail": c.CommitterEmail,
		"username":       c.Username,
		"createdAt":      c.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      c.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceGitCredAction deletes workspace git credentials.
func DeleteWorkspaceGitCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace git credential endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid git credentials ID",
		})
		return
	}

	repo := db.NewWorkspaceGitCredRepository(db.GetDB())
	c, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || c == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Git credentials not found",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceGitCredPermission(
		user,
		c,
		module.CanDeleteWorkspaceGitCred,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	if err := repo.Delete(c.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete workspace git credentials")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete git credentials",
		})
		return
	}

	log.Info().Str("gitCredID", idStr).Msg("Git credentials deleted successfully")
	service.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
