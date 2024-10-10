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

// CreateWorkspaceRequest represents the create workspace request payload
type CreateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255" label:"Name"`
}

// UpdateWorkspaceRequest represents the update workspace request payload
type UpdateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255" label:"Name"`
}

// CreateWorkspaceAction handles workspace creation requests
func CreateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace endpoint called")

	var req CreateWorkspaceRequest
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

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.CreateWorkspace(&module.CreateWorkspaceOptions{Name: req.Name})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateWorkspace"),
		})
		return
	}

	log.Info().Str("workspaceID", workspace.ID.String()).Msg("Workspace created successfully")
	wsModule.AddWorkspaceUser(workspace.ID, user.ID, db.UserRoleOwner)
	log.Info().Str("workspaceID", workspace.ID.String()).Str("userID", user.ID.String()).Msg("User added to workspace as owner")

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           workspace.ID,
		"name":         workspace.Name,
		"membersCount": 1,
		"role":         db.UserRoleOwner,
		"createdAt":    workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspacesAction handles workspace listing requests with pagination
func ListWorkspacesAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspaces endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	opts := &module.ListWorkspacesOptions{Limit: limit, Offset: offset}
	if user.Role != db.UserRoleAdmin || r.URL.Query().Get("all") != "true" {
		opts.UserID = &user.ID
	}

	wsModule := module.NewWorkspace(db.GetDB())
	result, err := wsModule.ListWorkspaces(opts)

	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspaces")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListWorkspaces"),
		})
		return
	}

	workspaceList := make([]map[string]interface{}, 0, len(result.Workspaces))
	for _, w := range result.Workspaces {
		membersCount, _ := wsModule.CountWorkspaceMembers(w.ID)
		item := map[string]interface{}{
			"id":           w.ID,
			"name":         w.Name,
			"membersCount": membersCount,
			"createdAt":    w.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":    w.UpdatedAt.UTC().Format(time.RFC3339),
		}
		if membership, _ := wsModule.GetWorkspaceMembership(w.ID, user.ID); membership != nil {
			item["role"] = membership.Role
		} else if user.Role == db.UserRoleAdmin {
			item["role"] = db.UserRoleAdmin
		}
		workspaceList = append(workspaceList, item)
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"workspaces": workspaceList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceAction handles get workspace by ID requests
func GetWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(idStr))
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
	allowed, err := permModule.CheckWorkspacePermission(
		user,
		workspace,
		module.CanGetWorkspace,
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

	membersCount, _ := wsModule.CountWorkspaceMembers(workspace.ID)

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           workspace.ID,
		"name":         workspace.Name,
		"membersCount": membersCount,
		"createdAt":    workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceAction handles workspace update requests
func UpdateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	var req UpdateWorkspaceRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(idStr))
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
	allowed, err := permModule.CheckWorkspacePermission(
		user,
		workspace,
		module.CanUpdateWorkspace,
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
			"errorMessage": locale.TR(r, "noPermissionUpdateWorkspace"),
		})
		return
	}

	workspace, err = wsModule.UpdateWorkspace(&module.UpdateWorkspaceOptions{
		WorkspaceID: db.ID(idStr),
		Name:        req.Name,
	})
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateWorkspace"),
		})
		return
	}

	membersCount, _ := wsModule.CountWorkspaceMembers(workspace.ID)

	log.Info().Str("workspaceID", workspace.ID.String()).Msg("Workspace updated successfully")
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           workspace.ID,
		"name":         workspace.Name,
		"membersCount": membersCount,
		"createdAt":    workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceAction handles workspace deletion requests
func DeleteWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(idStr))
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
	allowed, err := permModule.CheckWorkspacePermission(
		user,
		workspace,
		module.CanDeleteWorkspace,
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
			"errorMessage": locale.TR(r, "noPermissionDeleteWorkspace"),
		})
		return
	}

	log.Info().Str("workspaceID", idStr).Msg("Deleting workspace")

	err = wsModule.DeleteWorkspace(db.ID(idStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to delete workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteWorkspace"),
		})
		return
	}

	log.Info().Str("workspaceID", idStr).Msg("Workspace deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
