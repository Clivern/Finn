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
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceInventoryRequest is the body for creating an Ansible inventory.
type CreateWorkspaceInventoryRequest struct {
	Name               string `json:"name" validate:"required,max=100" label:"Name"`
	Type               string `json:"type" validate:"required,oneof=static static_yaml file,max=50" label:"Type"`
	Inventory          string `json:"inventory" validate:"omitempty,max=1000000" label:"Inventory"`
	InventoryPath      string `json:"inventoryPath" validate:"omitempty,max=512" label:"Path to Inventory file"`
	SSHKeyID           string `json:"sshKeyId" validate:"omitempty" label:"User Credentials"`
	CredentialID       string `json:"credentialId" validate:"omitempty" label:"User Credentials (Optional)"`
	BecomeKeyID        string `json:"becomeKeyId" validate:"omitempty" label:"Sudo Credentials (Key)"`
	BecomeCredentialID string `json:"becomeCredentialId" validate:"omitempty" label:"Sudo Credentials (Optional)"`
	RepositoryID       string `json:"repositoryId" validate:"omitempty" label:"Repository (Optional)"`
}

// UpdateWorkspaceInventoryRequest is the body for updating an Ansible inventory.
type UpdateWorkspaceInventoryRequest struct {
	Name               string `json:"name" validate:"required,max=100" label:"Name"`
	Type               string `json:"type" validate:"required,oneof=static static_yaml file,max=50" label:"Type"`
	Inventory          string `json:"inventory" validate:"omitempty,max=1000000" label:"Inventory"`
	InventoryPath      string `json:"inventoryPath" validate:"omitempty,max=512" label:"Path to Inventory file"`
	SSHKeyID           string `json:"sshKeyId" validate:"omitempty" label:"User Credentials"`
	CredentialID       string `json:"credentialId" validate:"omitempty" label:"User Credentials (Optional)"`
	BecomeKeyID        string `json:"becomeKeyId" validate:"omitempty" label:"Sudo Credentials (Key)"`
	BecomeCredentialID string `json:"becomeCredentialId" validate:"omitempty" label:"Sudo Credentials (Optional)"`
	RepositoryID       string `json:"repositoryId" validate:"omitempty" label:"Repository (Optional)"`
}

// CreateWorkspaceInventoryAction creates an inventory in the workspace.
func CreateWorkspaceInventoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace inventory endpoint called")

	var req CreateWorkspaceInventoryRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	switch req.Type {
	case db.InventoryTypeFile:
		if req.InventoryPath == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "inventoryFilePathRequired"),
			})
			return
		}
	case db.InventoryTypeStatic, db.InventoryTypeStaticYAML:
		if req.Inventory == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "inventoryContentRequired"),
			})
			return
		}
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
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceInventoryPermission(
		user,
		&db.WorkspaceInventory{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceInventory,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace inventory permission")
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

	inventory := &db.WorkspaceInventory{
		WorkspaceID:   workspace.ID,
		Name:          req.Name,
		Inventory:     req.Inventory,
		InventoryPath: req.InventoryPath,
		Type:          req.Type,
	}

	if req.SSHKeyID != "" {
		sid := db.ID(req.SSHKeyID)
		inventory.SSHKeyID = &sid
	}
	if req.CredentialID != "" {
		cid := db.ID(req.CredentialID)
		inventory.CredentialID = &cid
	}
	if req.BecomeKeyID != "" {
		bid := db.ID(req.BecomeKeyID)
		inventory.BecomeKeyID = &bid
	}
	if req.BecomeCredentialID != "" {
		cid := db.ID(req.BecomeCredentialID)
		inventory.BecomeCredentialID = &cid
	}
	if req.RepositoryID != "" {
		rid := db.ID(req.RepositoryID)
		inventory.RepositoryID = &rid
	}

	repo := db.NewWorkspaceInventoryRepository(db.GetDB())
	if err := repo.Create(inventory); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create workspace inventory")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateInventory"),
		})
		return
	}

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":                 inventory.ID,
		"workspaceId":        inventory.WorkspaceID,
		"name":               inventory.Name,
		"type":               inventory.Type,
		"inventory":          inventory.Inventory,
		"inventoryPath":      inventory.InventoryPath,
		"sshKeyId":           inventory.SSHKeyID,
		"credentialId":       inventory.CredentialID,
		"becomeKeyId":        inventory.BecomeKeyID,
		"becomeCredentialId": inventory.BecomeCredentialID,
		"repositoryId":       inventory.RepositoryID,
		"createdAt":          inventory.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":          inventory.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceInventoriesAction lists inventories in the workspace.
func ListWorkspaceInventoriesAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace inventories endpoint called")

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
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceInventoryPermission(
		user,
		&db.WorkspaceInventory{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceInventories,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace inventory permission")
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

	repo := db.NewWorkspaceInventoryRepository(db.GetDB())
	result, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to list workspace inventories")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListInventories"),
		})
		return
	}

	total, _ := repo.CountByWorkspaceID(workspace.ID)
	inventories := make([]map[string]interface{}, 0, len(result))
	for _, inventory := range result {
		inventories = append(inventories, map[string]interface{}{
			"id":                 inventory.ID,
			"workspaceId":        inventory.WorkspaceID,
			"name":               inventory.Name,
			"type":               inventory.Type,
			"inventory":          inventory.Inventory,
			"inventoryPath":      inventory.InventoryPath,
			"sshKeyId":           inventory.SSHKeyID,
			"credentialId":       inventory.CredentialID,
			"becomeKeyId":        inventory.BecomeKeyID,
			"becomeCredentialId": inventory.BecomeCredentialID,
			"repositoryId":       inventory.RepositoryID,
			"createdAt":          inventory.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":          inventory.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"inventories": inventories,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceInventoryAction returns one inventory by ID.
func GetWorkspaceInventoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace inventory endpoint called")

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
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidInventoryID"),
		})
		return
	}

	repo := db.NewWorkspaceInventoryRepository(db.GetDB())
	inventory, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || inventory == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "inventoryNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceInventoryPermission(
		user,
		inventory,
		module.CanGetWorkspaceInventory,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("inventoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace inventory permission")
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
		"id":                 inventory.ID,
		"workspaceId":        inventory.WorkspaceID,
		"name":               inventory.Name,
		"type":               inventory.Type,
		"inventory":          inventory.Inventory,
		"inventoryPath":      inventory.InventoryPath,
		"sshKeyId":           inventory.SSHKeyID,
		"credentialId":       inventory.CredentialID,
		"becomeKeyId":        inventory.BecomeKeyID,
		"becomeCredentialId": inventory.BecomeCredentialID,
		"repositoryId":       inventory.RepositoryID,
		"createdAt":          inventory.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":          inventory.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceInventoryAction updates a workspace inventory.
func UpdateWorkspaceInventoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace inventory endpoint called")

	var req UpdateWorkspaceInventoryRequest
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
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidInventoryID"),
		})
		return
	}

	repo := db.NewWorkspaceInventoryRepository(db.GetDB())
	inventory, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || inventory == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "inventoryNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceInventoryPermission(
		user,
		inventory,
		module.CanUpdateWorkspaceInventory,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("inventoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace inventory permission")
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

	inventory.Name = req.Name

	switch req.Type {
	case db.InventoryTypeFile:
		if req.InventoryPath == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "inventoryFilePathRequired"),
			})
			return
		}
	case db.InventoryTypeStatic, db.InventoryTypeStaticYAML:
		if req.Inventory == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "inventoryContentRequired"),
			})
			return
		}
	}

	inventory.Type = req.Type

	if req.Inventory != "" {
		inventory.Inventory = req.Inventory
	}
	if req.InventoryPath != "" {
		inventory.InventoryPath = req.InventoryPath
	}
	if req.SSHKeyID != "" {
		sid := db.ID(req.SSHKeyID)
		inventory.SSHKeyID = &sid
	}
	if req.CredentialID != "" {
		cid := db.ID(req.CredentialID)
		inventory.CredentialID = &cid
	}
	if req.BecomeKeyID != "" {
		bid := db.ID(req.BecomeKeyID)
		inventory.BecomeKeyID = &bid
	}
	if req.BecomeCredentialID != "" {
		cid := db.ID(req.BecomeCredentialID)
		inventory.BecomeCredentialID = &cid
	}
	if req.RepositoryID != "" {
		rid := db.ID(req.RepositoryID)
		inventory.RepositoryID = &rid
	}

	if err := repo.Update(inventory); err != nil {
		log.Error().
			Err(err).
			Str("inventoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to update workspace inventory")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateInventory"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":                 inventory.ID,
		"workspaceId":        inventory.WorkspaceID,
		"name":               inventory.Name,
		"type":               inventory.Type,
		"inventory":          inventory.Inventory,
		"inventoryPath":      inventory.InventoryPath,
		"sshKeyId":           inventory.SSHKeyID,
		"credentialId":       inventory.CredentialID,
		"becomeKeyId":        inventory.BecomeKeyID,
		"becomeCredentialId": inventory.BecomeCredentialID,
		"repositoryId":       inventory.RepositoryID,
		"createdAt":          inventory.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":          inventory.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceInventoryAction deletes a workspace inventory.
func DeleteWorkspaceInventoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace inventory endpoint called")

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
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidInventoryID"),
		})
		return
	}

	repo := db.NewWorkspaceInventoryRepository(db.GetDB())
	inventory, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || inventory == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "inventoryNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceInventoryPermission(
		user,
		inventory,
		module.CanDeleteWorkspaceInventory,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("inventoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace inventory permission")
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

	if err := repo.Delete(inventory.ID); err != nil {
		log.Error().
			Err(err).
			Str("inventoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to delete workspace inventory")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteInventory"),
		})
		return
	}

	log.Info().
		Str("inventoryID", idStr).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Inventory deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
