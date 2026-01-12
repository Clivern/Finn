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

// CreateWorkspaceVariableGroupRequest is the body for creating a variable group.
type CreateWorkspaceVariableGroupRequest struct {
	Name                       string `json:"name" validate:"required,max=255" label:"Name"`
	ExtraVariables             string `json:"extraVariables" validate:"omitempty,json_object" label:"Extra variables"`
	EnvironmentVariables       string `json:"environmentVariables" validate:"omitempty,json_object" label:"Environment variables"`
	SecretExtraVariables       string `json:"secretExtraVariables" validate:"omitempty,json_object" label:"Secret extra variables"`
	SecretEnvironmentVariables string `json:"secretEnvironmentVariables" validate:"omitempty,json_object" label:"Secret environment variables"`
}

// UpdateWorkspaceVariableGroupRequest is the body for updating a variable group.
type UpdateWorkspaceVariableGroupRequest struct {
	Name                       string `json:"name" validate:"required,max=255" label:"Name"`
	ExtraVariables             string `json:"extraVariables" validate:"omitempty,json_object" label:"Extra variables"`
	EnvironmentVariables       string `json:"environmentVariables" validate:"omitempty,json_object" label:"Environment variables"`
	SecretExtraVariables       string `json:"secretExtraVariables" validate:"omitempty,json_object" label:"Secret extra variables"`
	SecretEnvironmentVariables string `json:"secretEnvironmentVariables" validate:"omitempty,json_object" label:"Secret environment variables"`
}

// CreateWorkspaceVariableGroupAction creates a variable group in the workspace.
func CreateWorkspaceVariableGroupAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace variable group endpoint called")

	var req CreateWorkspaceVariableGroupRequest
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

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceVariableGroupPermission(
		user,
		&db.WorkspaceVariableGroup{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceVariableGroup,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace permission")
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

	secretExtraVariables, err := util.EncryptToBase64String(
		req.SecretExtraVariables,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt secret extra variables")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedEncryptSecretExtraVariables"),
		})
		return
	}

	secretEnvironmentVariables, err := util.EncryptToBase64String(
		req.SecretEnvironmentVariables,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt secret environment variables")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedEncryptSecretEnvironmentVariables"),
		})
		return
	}

	repo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
	group := &db.WorkspaceVariableGroup{
		WorkspaceID:                workspace.ID,
		Name:                       req.Name,
		ExtraVariables:             req.ExtraVariables,
		EnvironmentVariables:       req.EnvironmentVariables,
		SecretExtraVariables:       secretExtraVariables,
		SecretEnvironmentVariables: secretEnvironmentVariables,
	}

	if err := repo.Create(group); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create workspace variable group")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateVariableGroup"),
		})
		return
	}

	log.Info().
		Str("groupID", group.ID.String()).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Workspace variable group created successfully")

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":                         group.ID,
		"workspaceId":                group.WorkspaceID,
		"name":                       group.Name,
		"extraVariables":             group.ExtraVariables,
		"environmentVariables":       group.EnvironmentVariables,
		"secretExtraVariables":       group.SecretExtraVariables,
		"secretEnvironmentVariables": group.SecretEnvironmentVariables,
		"createdAt":                  group.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":                  group.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceVariableGroupsAction lists variable groups in the workspace.
func ListWorkspaceVariableGroupsAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace variable groups endpoint called")

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
	allowed, err := permModule.CheckWorkspaceVariableGroupPermission(
		user,
		&db.WorkspaceVariableGroup{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceVariableGroups,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace permission")
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

	repo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
	result, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to list workspace variable groups")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListVariableGroups"),
		})
		return
	}

	total, _ := repo.CountByWorkspaceID(workspace.ID)
	groups := make([]map[string]interface{}, 0, len(result))
	for _, g := range result {
		secretExtraVariables := util.DecryptFromBase64String(
			g.SecretExtraVariables,
			util.GetEncryptionKey(),
		)
		secretEnvironmentVariables := util.DecryptFromBase64String(
			g.SecretEnvironmentVariables,
			util.GetEncryptionKey(),
		)
		groups = append(groups, map[string]interface{}{
			"id":                         g.ID,
			"workspaceId":                g.WorkspaceID,
			"name":                       g.Name,
			"extraVariables":             g.ExtraVariables,
			"environmentVariables":       g.EnvironmentVariables,
			"secretExtraVariables":       secretExtraVariables,
			"secretEnvironmentVariables": secretEnvironmentVariables,
			"createdAt":                  g.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":                  g.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceVariableGroupAction returns one variable group.
func GetWorkspaceVariableGroupAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace variable group endpoint called")

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
			"errorMessage": locale.TR(r, "invalidVariableGroupID"),
		})
		return
	}

	repo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
	group, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace variable group")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetVariableGroup"),
		})
		return
	}
	if group == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "variableGroupNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceVariableGroupPermission(
		user,
		group,
		module.CanGetWorkspaceVariableGroup,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace permission")
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
		"id":                         group.ID,
		"workspaceId":                group.WorkspaceID,
		"name":                       group.Name,
		"extraVariables":             group.ExtraVariables,
		"environmentVariables":       group.EnvironmentVariables,
		"secretExtraVariables":       group.SecretExtraVariables,
		"secretEnvironmentVariables": group.SecretEnvironmentVariables,
		"createdAt":                  group.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":                  group.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceVariableGroupAction updates a workspace variable group.
func UpdateWorkspaceVariableGroupAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace variable group endpoint called")

	var req UpdateWorkspaceVariableGroupRequest
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
			"errorMessage": locale.TR(r, "invalidVariableGroupID"),
		})
		return
	}

	repo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
	group, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || group == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "variableGroupNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceVariableGroupPermission(
		user,
		group,
		module.CanUpdateWorkspaceVariableGroup,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace permission")
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

	secretExtraVariables, err := util.EncryptToBase64String(
		req.SecretExtraVariables,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt secret extra variables")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedEncryptSecretExtraVariables"),
		})
		return
	}

	secretEnvironmentVariables, err := util.EncryptToBase64String(
		req.SecretEnvironmentVariables,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt secret environment variables")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedEncryptSecretEnvironmentVariables"),
		})
		return
	}

	group.Name = req.Name
	group.ExtraVariables = req.ExtraVariables
	group.EnvironmentVariables = req.EnvironmentVariables
	group.SecretExtraVariables = secretExtraVariables
	group.SecretEnvironmentVariables = secretEnvironmentVariables

	if err := repo.Update(group); err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to update workspace variable group")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateVariableGroup"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":                         group.ID,
		"workspaceId":                group.WorkspaceID,
		"name":                       group.Name,
		"extraVariables":             group.ExtraVariables,
		"environmentVariables":       group.EnvironmentVariables,
		"secretExtraVariables":       group.SecretExtraVariables,
		"secretEnvironmentVariables": group.SecretEnvironmentVariables,
		"createdAt":                  group.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":                  group.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceVariableGroupAction deletes a workspace variable group.
func DeleteWorkspaceVariableGroupAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace variable group endpoint called")

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
			"errorMessage": locale.TR(r, "invalidVariableGroupID"),
		})
		return
	}

	repo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
	group, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || group == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "variableGroupNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceVariableGroupPermission(
		user,
		group,
		module.CanDeleteWorkspaceVariableGroup,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace permission")
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

	if err := repo.Delete(group.ID); err != nil {
		log.Error().
			Err(err).
			Str("groupID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to delete workspace variable group")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteVariableGroup"),
		})
		return
	}

	log.Info().
		Str("groupID", idStr).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Workspace variable group deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
