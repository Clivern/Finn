// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/clivern/tivo/db"
	"github.com/clivern/tivo/middleware"
	"github.com/clivern/tivo/module"
	"github.com/clivern/tivo/service"
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
	log.Debug().Msg("Create workspace endpoint called")

	var req CreateWorkspaceRequest
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

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.CreateWorkspace(&module.CreateWorkspaceOptions{Name: req.Name})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create workspace",
		})
		return
	}

	log.Info().Str("workspaceID", workspace.ID.String()).Msg("Workspace created successfully")

	wsModule.AddWorkspaceUser(workspace.ID, user.ID, db.UserRoleOwner)
	log.Info().Str("workspaceID", workspace.ID.String()).Str("userID", user.ID.String()).Msg("User added to workspace as owner")

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        workspace.ID,
		"name":      workspace.Name,
		"createdAt": workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": workspace.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspacesAction handles workspace listing requests with pagination.
func ListWorkspacesAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List workspaces endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
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
	// By default load only user's workspaces; admins can pass ?all=true to get all workspaces
	if user.Role != db.UserRoleAdmin || r.URL.Query().Get("all") != "true" {
		opts.UserID = &user.ID
	}

	wsModule := module.NewWorkspace(db.GetDB())
	result, err := wsModule.ListWorkspaces(opts)

	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspaces")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list workspaces",
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

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"workspaces": workspaceList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceAction handles get workspace by ID requests. Allowed for app admins and workspace members.
func GetWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get workspace endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(idStr))
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
	allowed, err := permModule.CheckWorkspacePermission(user, workspace, module.CanGetWorkspace)
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
		"id":        workspace.ID,
		"name":      workspace.Name,
		"createdAt": workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": workspace.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceAction handles workspace update requests. Allowed for app admins and workspace admin/owner.
func UpdateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update workspace endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	var req UpdateWorkspaceRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(idStr))
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
	allowed, err := permModule.CheckWorkspacePermission(user, workspace, module.CanUpdateWorkspace)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "You do not have permission to update this workspace",
		})
		return
	}

	workspace, err = wsModule.UpdateWorkspace(&module.UpdateWorkspaceOptions{
		WorkspaceID: db.ID(idStr),
		Name:        req.Name,
	})
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update workspace",
		})
		return
	}

	log.Info().Str("workspaceID", workspace.ID.String()).Msg("Workspace updated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":        workspace.ID,
		"name":      workspace.Name,
		"createdAt": workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": workspace.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceAction handles workspace deletion requests. Allowed for app admins and workspace admin/owner.
func DeleteWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Delete workspace endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(idStr))
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
	allowed, err := permModule.CheckWorkspacePermission(user, workspace, module.CanDeleteWorkspace)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "You do not have permission to delete this workspace",
		})
		return
	}

	err = wsModule.DeleteWorkspace(db.ID(idStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to delete workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete workspace",
		})
		return
	}

	log.Info().Str("workspaceID", idStr).Msg("Workspace deleted successfully")
	service.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
