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
	log.Info().Msg("Add workspace user endpoint called")

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

	var req AddWorkspaceUserRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
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
	allowed, err := permModule.CheckWorkspaceUserPermission(
		user,
		&db.WorkspaceUser{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceUser,
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
			"errorMessage": locale.TR(r, "noPermissionAddUsersToWorkspace"),
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	targetUser, err := userRepo.GetByID(db.ID(req.UserID))
	if err != nil || targetUser == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "userNotFound"),
		})
		return
	}

	existing, _ := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(req.UserID))
	if existing != nil {
		util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": locale.TR(r, "userAlreadyMemberOfWorkspace"),
		})
		return
	}

	workspaceUser, err := wsModule.AddWorkspaceUser(
		db.ID(workspaceIDStr), db.ID(req.UserID), req.Role,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add workspace user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedAddUserToWorkspace"),
		})
		return
	}

	log.Info().
		Str("workspaceID", workspaceIDStr).
		Str("userID", targetUser.ID.String()).
		Str("role", workspaceUser.Role).
		Msg("User added to workspace successfully")

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          workspaceUser.ID,
		"userId":      workspaceUser.UserID,
		"name":        targetUser.Name,
		"email":       targetUser.Email,
		"role":        workspaceUser.Role,
		"isActive":    targetUser.IsActive,
		"avatar":      util.GravatarURL(targetUser.Email, 80),
		"lastLoginAt": targetUser.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   workspaceUser.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   workspaceUser.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceUsersAction lists users in a workspace
func ListWorkspaceUsersAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace users endpoint called")

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
	allowed, err := permModule.CheckWorkspaceUserPermission(
		user,
		&db.WorkspaceUser{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceUsers,
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

	result, err := wsModule.ListWorkspaceUsers(&module.ListWorkspaceUsersOptions{
		WorkspaceID: db.ID(workspaceIDStr),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspace users")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListWorkspaceUsers"),
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	users := make([]map[string]interface{}, 0, len(result.WorkspaceUsers))
	for _, wu := range result.WorkspaceUsers {
		u, _ := userRepo.GetByID(wu.UserID)
		if u != nil {
			users = append(users, map[string]interface{}{
				"id":          wu.ID,
				"userId":      wu.UserID,
				"name":        u.Name,
				"email":       u.Email,
				"role":        wu.Role,
				"isActive":    u.IsActive,
				"avatar":      util.GravatarURL(u.Email, 80),
				"lastLoginAt": u.LastLoginAt.UTC().Format(time.RFC3339),
				"createdAt":   wu.CreatedAt.UTC().Format(time.RFC3339),
				"updatedAt":   wu.UpdatedAt.UTC().Format(time.RFC3339),
			})
		}
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceUserAction returns a single workspace user membership
func GetWorkspaceUserAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace user endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	userIDStr := chi.URLParam(r, "userID")
	if workspaceIDStr == "" || userIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceIDOrUserID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	_, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
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

	workspaceUser, err := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspaceUser"),
		})
		return
	}
	if workspaceUser == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "userNotMemberOfWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceUserPermission(
		user,
		workspaceUser,
		module.CanGetWorkspaceUser,
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

	userRepo := db.NewUserRepository(db.GetDB())
	targetUser, _ := userRepo.GetByID(workspaceUser.UserID)

	if targetUser == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "userNotFound"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          workspaceUser.ID,
		"userId":      workspaceUser.UserID,
		"name":        targetUser.Name,
		"email":       targetUser.Email,
		"role":        workspaceUser.Role,
		"isActive":    targetUser.IsActive,
		"avatar":      util.GravatarURL(targetUser.Email, 80),
		"lastLoginAt": targetUser.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   workspaceUser.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   workspaceUser.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceUserRoleAction updates a workspace user's role
func UpdateWorkspaceUserRoleAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace user role endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	userIDStr := chi.URLParam(r, "userID")
	if workspaceIDStr == "" || userIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceIDOrUserID"),
		})
		return
	}

	var req UpdateWorkspaceUserRoleRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	_, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
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

	workspaceUser, err := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspaceUser"),
		})
		return
	}
	if workspaceUser == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "userNotMemberOfWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceUserPermission(
		user,
		workspaceUser,
		module.CanUpdateWorkspaceUser,
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
			"errorMessage": locale.TR(r, "noPermissionUpdateUsersInWorkspace"),
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
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "userNotMemberOfWorkspace"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update workspace user role")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateWorkspaceUserRole"),
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	targetUser, _ := userRepo.GetByID(workspaceUser.UserID)

	if targetUser == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "userNotFound"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          workspaceUser.ID,
		"userId":      workspaceUser.UserID,
		"name":        targetUser.Name,
		"email":       targetUser.Email,
		"role":        workspaceUser.Role,
		"isActive":    targetUser.IsActive,
		"avatar":      util.GravatarURL(targetUser.Email, 80),
		"lastLoginAt": targetUser.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   workspaceUser.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   workspaceUser.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// RemoveWorkspaceUserAction removes a user from a workspace
func RemoveWorkspaceUserAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Remove workspace user endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	userIDStr := chi.URLParam(r, "userID")
	if workspaceIDStr == "" || userIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceIDOrUserID"),
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	_, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
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

	workspaceUser, err := wsModule.GetWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspaceUser"),
		})
		return
	}
	if workspaceUser == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "userNotMemberOfWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceUserPermission(
		user,
		workspaceUser,
		module.CanDeleteWorkspaceUser,
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
			"errorMessage": locale.TR(r, "noPermissionRemoveUsersFromWorkspace"),
		})
		return
	}

	err = wsModule.RemoveWorkspaceUser(db.ID(workspaceIDStr), db.ID(userIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceUserNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "userNotMemberOfWorkspace"),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to remove workspace user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedRemoveUserFromWorkspace"),
		})
		return
	}

	log.Info().
		Str("workspaceID", workspaceIDStr).
		Str("userID", userIDStr).
		Msg("User removed from workspace")

	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}

// ListAllUsersAction lists all users for the workspace add-user form. Only callable by users with
func ListAllUsersAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List all users (add form) endpoint called")

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
			"errorMessage": locale.TR(r, "noPermissionAddUsersToWorkspace"),
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
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListUsersForWorkspace"),
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
			"id":          u.ID,
			"name":        u.Name,
			"email":       u.Email,
			"isActive":    u.IsActive,
			"avatar":      util.GravatarURL(u.Email, 80),
			"lastLoginAt": u.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":   u.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   u.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"users": userList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}
