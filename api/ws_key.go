// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/service/util"
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
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckAccess"),
		})
		return
	}
	if !allowed {
		util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": locale.TR(r, "accessDeniedToWorkspace"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateSshKey"),
		})
		return
	}

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
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
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckAccess"),
		})
		return
	}
	if !allowed {
		util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": locale.TR(r, "accessDeniedToWorkspace"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListSshKeys"),
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

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
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
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidSshKeyID"),
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	k, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace SSH key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetSshKey"),
		})
		return
	}
	if k == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "sshKeyNotFound"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckAccess"),
		})
		return
	}
	if !allowed {
		util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": locale.TR(r, "accessDeniedToWorkspace"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
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
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidSshKeyID"),
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	k, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || k == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "sshKeyNotFound"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckAccess"),
		})
		return
	}
	if !allowed {
		util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": locale.TR(r, "accessDeniedToWorkspace"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateSshKey"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
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
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidSshKeyID"),
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	k, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || k == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "sshKeyNotFound"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckAccess"),
		})
		return
	}
	if !allowed {
		util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": locale.TR(r, "accessDeniedToWorkspace"),
		})
		return
	}

	if err := repo.Delete(k.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete workspace SSH key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteSshKey"),
		})
		return
	}

	log.Info().Str("sshKeyID", idStr).Msg("SSH key deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
