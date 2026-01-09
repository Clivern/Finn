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

// AddWorkspaceUserRequest represents the add workspace user request payload
type AddWorkspaceUserRequest struct {
	UserID string `json:"userId" validate:"required,min=1" label:"User ID"`
	Role   string `json:"role" validate:"required,oneof=admin owner readonly" label:"Role"`
}

// UpdateWorkspaceUserRoleRequest represents the update workspace user role request payload
type UpdateWorkspaceUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin owner readonly" label:"Role"`
}

// AddWorkspaceUserAction adds a user to a workspace
func AddWorkspaceUserAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Add workspace user endpoint called")

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

	var req AddWorkspaceUserRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
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
	allowed, err := permModule.CheckWorkspaceUserPermission(user, &db.WorkspaceUser{WorkspaceID: workspace.ID}, module.CanAddWorkspaceUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "You do not have permission to add users to this workspace",
		})
		return
	}

	// Ensure target user exists
	userRepo := db.NewUserRepository(db.GetDB())
	targetUser, err := userRepo.GetByID(db.ID(req.UserID))
	if err != nil || targetUser == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "User not found",
		})
		return
	}

	existing, _ := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(req.UserID))
	if existing != nil {
		service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": "User is already a member of this workspace",
		})
		return
	}

	workspaceUser, err := wsModule.AddWorkspaceUser(db.ID(workspaceIDStr), db.ID(req.UserID), req.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add workspace user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to add user to workspace",
		})
		return
	}

	log.Info().
		Str("workspaceID", workspaceIDStr).
		Str("userID", targetUser.ID.String()).
		Msg("User added to workspace")

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          workspaceUser.ID,
		"workspaceId": workspaceUser.WorkspaceID,
		"userId":      workspaceUser.UserID,
		"role":        workspaceUser.Role,
		"createdAt":   workspaceUser.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   workspaceUser.UpdatedAt.UTC().Format(time.RFC3339),
		"user": map[string]interface{}{
			"id":    targetUser.ID,
			"name":  targetUser.Name,
			"email": targetUser.Email,
		},
	})
}

// ListWorkspaceUsersAction lists users in a workspace
func ListWorkspaceUsersAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List workspace users endpoint called")

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

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 50
	offset := 0
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
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
	allowed, err := permModule.CheckWorkspaceUserPermission(user, &db.WorkspaceUser{WorkspaceID: workspace.ID}, module.CanGetWorkspaceUsers)
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

	result, err := wsModule.ListWorkspaceUsers(&module.ListWorkspaceUsersOptions{
		WorkspaceID: db.ID(workspaceIDStr),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspace users")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list workspace users",
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	workspaceUserList := make([]map[string]interface{}, 0, len(result.WorkspaceUsers))
	for _, wu := range result.WorkspaceUsers {
		u, _ := userRepo.GetByID(wu.UserID)
		item := map[string]interface{}{
			"id":          wu.ID,
			"workspaceId": wu.WorkspaceID,
			"userId":      wu.UserID,
			"role":        wu.Role,
			"createdAt":   wu.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   wu.UpdatedAt.UTC().Format(time.RFC3339),
		}
		if u != nil {
			item["user"] = map[string]interface{}{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			}
		}
		workspaceUserList = append(workspaceUserList, item)
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"workspaceUsers": workspaceUserList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceUserAction returns a single workspace user membership
func GetWorkspaceUserAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get workspace user endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	userIDStr := chi.URLParam(r, "userID")
	if workspaceIDStr == "" || userIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID or user ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	_, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
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

	workspaceUser, err := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace user",
		})
		return
	}
	if workspaceUser == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "User is not a member of this workspace",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceUserPermission(user, workspaceUser, module.CanGetWorkspaceUser)
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

	userRepo := db.NewUserRepository(db.GetDB())
	targetUser, _ := userRepo.GetByID(workspaceUser.UserID)
	resp := map[string]interface{}{
		"id":          workspaceUser.ID,
		"workspaceId": workspaceUser.WorkspaceID,
		"userId":      workspaceUser.UserID,
		"role":        workspaceUser.Role,
		"createdAt":   workspaceUser.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   workspaceUser.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if targetUser != nil {
		resp["user"] = map[string]interface{}{
			"id":    targetUser.ID,
			"name":  targetUser.Name,
			"email": targetUser.Email,
		}
	}
	service.WriteJSON(w, http.StatusOK, resp)
}

// UpdateWorkspaceUserRoleAction updates a workspace user's role
func UpdateWorkspaceUserRoleAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update workspace user role endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	userIDStr := chi.URLParam(r, "userID")
	if workspaceIDStr == "" || userIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID or user ID",
		})
		return
	}

	var req UpdateWorkspaceUserRoleRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	_, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
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

	workspaceUser, err := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace user",
		})
		return
	}
	if workspaceUser == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "User is not a member of this workspace",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceUserPermission(user, workspaceUser, module.CanUpdateWorkspaceUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "You do not have permission to update users in this workspace",
		})
		return
	}

	workspaceUser, err = wsModule.UpdateWorkspaceUserRole(&module.UpdateWorkspaceUserRoleOptions{
		WorkspaceID: db.ID(workspaceIDStr),
		UserID:      db.ID(userIDStr),
		Role:        req.Role,
	})
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User is not a member of this workspace",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update workspace user role")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update workspace user role",
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	targetUser, _ := userRepo.GetByID(workspaceUser.UserID)
	resp := map[string]interface{}{
		"id":          workspaceUser.ID,
		"workspaceId": workspaceUser.WorkspaceID,
		"userId":      workspaceUser.UserID,
		"role":        workspaceUser.Role,
		"createdAt":   workspaceUser.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   workspaceUser.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if targetUser != nil {
		resp["user"] = map[string]interface{}{
			"id":    targetUser.ID,
			"name":  targetUser.Name,
			"email": targetUser.Email,
		}
	}
	service.WriteJSON(w, http.StatusOK, resp)
}

// RemoveWorkspaceUserAction removes a user from a workspace
func RemoveWorkspaceUserAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Remove workspace user endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	userIDStr := chi.URLParam(r, "userID")
	if workspaceIDStr == "" || userIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID or user ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	_, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
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

	workspaceUser, err := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace user",
		})
		return
	}
	if workspaceUser == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "User is not a member of this workspace",
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceUserPermission(user, workspaceUser, module.CanDeleteWorkspaceUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "You do not have permission to remove users from this workspace",
		})
		return
	}

	err = wsModule.RemoveWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User is not a member of this workspace",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to remove workspace user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to remove user from workspace",
		})
		return
	}

	log.Info().
		Str("workspaceID", workspaceIDStr).
		Str("userID", userIDStr).
		Msg("User removed from workspace")

	w.WriteHeader(http.StatusNoContent)
}

// ListAllUsersAction lists all users for the workspace add-user form. Only callable by users with
func ListAllUsersAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List all users (add form) endpoint called")

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

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	excludeMembers := r.URL.Query().Get("excludeMembers") == "true" || r.URL.Query().Get("excludeMembers") == "1"

	limit := 50
	offset := 0
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
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
			"errorMessage": "You do not have permission to add users to this workspace",
		})
		return
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	result, err := userModule.ListUsers(&module.ListUsersOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list users",
		})
		return
	}

	if excludeMembers {
		membersResult, err := wsModule.ListWorkspaceUsers(&module.ListWorkspaceUsersOptions{
			WorkspaceID: db.ID(workspaceIDStr),
			Limit:       10000,
			Offset:      0,
		})
		if err == nil {
			memberIDs := make(map[db.ID]struct{})
			for _, wu := range membersResult.WorkspaceUsers {
				memberIDs[wu.UserID] = struct{}{}
			}
			filtered := result.Users[:0]
			for _, u := range result.Users {
				if _, in := memberIDs[u.ID]; !in {
					filtered = append(filtered, u)
				}
			}
			result.Users = filtered
			result.Total = int64(len(filtered))
		}
	}

	userList := make([]map[string]interface{}, 0, len(result.Users))
	for _, u := range result.Users {
		userList = append(userList, map[string]interface{}{
			"id":       u.ID,
			"name":     u.Name,
			"email":    u.Email,
			"isActive": u.IsActive,
		})
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"users": userList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}
