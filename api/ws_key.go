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

// CreateWorkspaceSSHKeyRequest is the request body for creating a workspace SSH key.
type CreateWorkspaceSSHKeyRequest struct {
	Name           string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName  string `json:"committerName" validate:"required,max=100" label:"Committer name"`
	CommitterEmail string `json:"committerEmail" validate:"required,max=100" label:"Committer email"`
	PrivateKey     string `json:"privateKey" validate:"required" label:"Private key"`
	PublicKey      string `json:"publicKey" validate:"omitempty" label:"Public key"`
}

// UpdateWorkspaceSSHKeyRequest is the request body for updating a workspace SSH key.
type UpdateWorkspaceSSHKeyRequest struct {
	Name           string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName  string `json:"committerName" validate:"required,max=100" label:"Committer name"`
	CommitterEmail string `json:"committerEmail" validate:"required,max=100" label:"Committer email"`
	PrivateKey     string `json:"privateKey" validate:"omitempty" label:"Private key"`
	PublicKey      string `json:"publicKey" validate:"omitempty" label:"Public key"`
}

// CreateWorkspaceSSHKeyAction creates an SSH key in a workspace.
func CreateWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace SSH key endpoint called")

	var req CreateWorkspaceSSHKeyRequest
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
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		&db.WorkspaceSSHKey{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceSSHKey,
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

	key := &db.WorkspaceSSHKey{
		WorkspaceID:    workspace.ID,
		Name:           req.Name,
		CommitterName:  req.CommitterName,
		CommitterEmail: req.CommitterEmail,
		PrivateKey:     &req.PrivateKey,
	}
	if req.PublicKey != "" {
		key.PublicKey = &req.PublicKey
	}
	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	if err := repo.Create(key); err != nil {
		log.Error().Err(err).Msg("Failed to create workspace SSH key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create SSH key",
		})
		return
	}

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             key.ID,
		"workspaceId":    key.WorkspaceID,
		"name":           key.Name,
		"committerName":  key.CommitterName,
		"committerEmail": key.CommitterEmail,
		"publicKey":      key.PublicKey,
		"createdAt":      key.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      key.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceSSHKeysAction lists SSH keys in a workspace
func ListWorkspaceSSHKeysAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace SSH keys endpoint called")

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
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		&db.WorkspaceSSHKey{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceSSHKeys,
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
	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	list, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspace SSH keys")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list SSH keys",
		})
		return
	}
	total, _ := repo.CountByWorkspaceID(workspace.ID)
	items := make([]map[string]interface{}, 0, len(list))
	for _, k := range list {
		item := map[string]interface{}{
			"id":             k.ID,
			"workspaceId":    k.WorkspaceID,
			"name":           k.Name,
			"committerName":  k.CommitterName,
			"committerEmail": k.CommitterEmail,
			"publicKey":      k.PublicKey,
			"createdAt":      k.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":      k.UpdatedAt.UTC().Format(time.RFC3339),
		}
		items = append(items, item)
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"keys": items,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceSSHKeyAction returns a single workspace SSH key (private key not returned for security).
func GetWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace SSH key endpoint called")

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
			"errorMessage": "Invalid SSH key ID",
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	k, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace SSH key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get SSH key",
		})
		return
	}
	if k == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "SSH key not found",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		k,
		module.CanGetWorkspaceSSHKey,
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
		"id":             k.ID,
		"workspaceId":    k.WorkspaceID,
		"name":           k.Name,
		"committerName":  k.CommitterName,
		"committerEmail": k.CommitterEmail,
		"publicKey":      k.PublicKey,
		"createdAt":      k.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      k.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceSSHKeyAction updates a workspace SSH key.
func UpdateWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace SSH key endpoint called")

	var req UpdateWorkspaceSSHKeyRequest
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
			"errorMessage": "Invalid SSH key ID",
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	k, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || k == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "SSH key not found",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		k,
		module.CanUpdateWorkspaceSSHKey,
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

	k.Name = req.Name
	k.CommitterName = req.CommitterName
	k.CommitterEmail = req.CommitterEmail
	if req.PrivateKey != "" {
		k.PrivateKey = &req.PrivateKey
	}
	if req.PublicKey != "" {
		k.PublicKey = &req.PublicKey
	}
	if err := repo.Update(k); err != nil {
		log.Error().Err(err).Msg("Failed to update workspace SSH key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update SSH key",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":             k.ID,
		"workspaceId":    k.WorkspaceID,
		"name":           k.Name,
		"committerName":  k.CommitterName,
		"committerEmail": k.CommitterEmail,
		"publicKey":      k.PublicKey,
		"createdAt":      k.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      k.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceSSHKeyAction deletes a workspace SSH key.
func DeleteWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace SSH key endpoint called")

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
			"errorMessage": "Invalid SSH key ID",
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	k, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || k == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "SSH key not found",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		k,
		module.CanDeleteWorkspaceSSHKey,
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

	if err := repo.Delete(k.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete workspace SSH key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete SSH key",
		})
		return
	}

	log.Info().Str("sshKeyID", idStr).Msg("SSH key deleted successfully")
	service.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
