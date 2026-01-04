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

// CreateUserRequest is the body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=4,max=60" label:"Name"`
	Email    string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
	Password string `json:"password" validate:"required,strong_password,min=8,max=60" label:"Password"`
	Role     string `json:"role" validate:"required,oneof=admin regular" label:"Role"`
	IsActive bool   `json:"isActive" label:"Is Active"`
}

// UpdateUserRequest is the body for updating a user.
type UpdateUserRequest struct {
	Name     string `json:"name" validate:"required,min=4,max=60" label:"Name"`
	Email    string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
	Password string `json:"password" validate:"omitempty,strong_password,min=8,max=60" label:"Password"`
	Role     string `json:"role" validate:"required,oneof=admin regular" label:"Role"`
	IsActive bool   `json:"isActive" label:"Is Active"`
}

// CreateUserAction creates a new user.
func CreateUserAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create user endpoint called")

	var req CreateUserRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	user, err := userModule.CreateUser(&module.CreateUserOptions{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		IsActive: req.IsActive,
	})

	if err != nil {
		if errors.Is(err, module.ErrUserEmailAlreadyExists) {
			util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": locale.TR(r, "userWithEmailExists"),
			})
			return
		}
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to create user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateUser"),
		})
		return
	}

	log.Info().
		Str("userID", user.ID.String()).
		Msg("User created successfully")
	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"role":        user.Role,
		"isActive":    user.IsActive,
		"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// GetUserAction returns one user by ID.
func GetUserAction(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidUserID"),
		})
		return
	}

	log.Info().
		Str("userID", userIDStr).
		Msg("Get user endpoint called")

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	user, err := userModule.GetUser(db.ID(userIDStr))
	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "userNotFound"),
			})
			return
		}
		log.Error().
			Err(err).
			Str("userID", userIDStr).
			Msg("Failed to get user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetUser"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"role":        user.Role,
		"isActive":    user.IsActive,
		"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateUserAction updates a user.
func UpdateUserAction(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidUserID"),
		})
		return
	}

	var req UpdateUserRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("userID", userIDStr).
		Msg("Update user endpoint called")

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	user, err := userModule.UpdateUser(&module.UpdateUserOptions{
		UserID:   db.ID(userIDStr),
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		IsActive: req.IsActive,
	})

	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "userNotFound"),
			})
			return
		}
		if errors.Is(err, module.ErrUserEmailAlreadyExists) {
			util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": locale.TR(r, "userWithEmailExists"),
			})
			return
		}
		log.Error().
			Err(err).
			Str("userID", userIDStr).
			Msg("Failed to update user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateUser"),
		})
		return
	}

	log.Info().
		Str("userID", user.ID.String()).
		Msg("User updated successfully")
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"role":        user.Role,
		"isActive":    user.IsActive,
		"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListUsersAction returns users (paginated).
func ListUsersAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List users endpoint called")

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

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	result, err := userModule.ListUsers(&module.ListUsersOptions{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		log.Error().
			Err(err).
			Int("limit", limit).
			Int("offset", offset).
			Msg("Failed to list users")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListUsers"),
		})
		return
	}

	users := make([]map[string]interface{}, 0, len(result.Users))
	for _, user := range result.Users {
		users = append(users, map[string]interface{}{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"role":        user.Role,
			"isActive":    user.IsActive,
			"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
		})
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

// DeleteUserAction deletes a user.
func DeleteUserAction(w http.ResponseWriter, r *http.Request) {
	currentUser, _ := middleware.GetUserFromContext(r.Context())

	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidUserID"),
		})
		return
	}
	userID := db.ID(userIDStr)

	log.Info().
		Str("userID", userIDStr).
		Str("actorID", currentUser.ID.String()).
		Msg("Delete user endpoint called")

	if currentUser.ID == userID {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "cannotDeleteOwnAccount"),
		})
		return
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	err := userModule.DeleteUser(userID)

	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "userNotFound"),
			})
			return
		}
		log.Error().
			Err(err).
			Str("userID", userIDStr).
			Str("actorID", currentUser.ID.String()).
			Msg("Failed to delete user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteUser"),
		})
		return
	}

	log.Info().
		Str("userID", userID.String()).
		Str("actorID", currentUser.ID.String()).
		Msg("User deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
