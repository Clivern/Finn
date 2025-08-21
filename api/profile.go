// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"

	"github.com/rs/zerolog/log"
)

// UpdateProfileRequest is the body for updating your profile.
type UpdateProfileRequest struct {
	Name            string `json:"name" validate:"omitempty,max=255" label:"Name"`
	Email           string `json:"email" validate:"omitempty,email,max=255" label:"Email"`
	CurrentPassword string `json:"currentPassword" validate:"omitempty,max=60" label:"Current password"`
	NewPassword     string `json:"newPassword" validate:"omitempty,strong_password,min=8,max=60" label:"New password"`
}

// GetProfileAction returns the current user's profile.
func GetProfileAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get profile endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"role":        user.Role,
			"isActive":    user.IsActive,
			"name":        user.Name,
			"avatar":      util.GravatarURL(user.Email, 80),
			"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
		},
	})
}

// UpdateProfileAction updates the current user's profile.
func UpdateProfileAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update profile endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	var req UpdateProfileRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	if (req.Email != "" && req.Email != user.Email) || (req.NewPassword != "") {
		if req.CurrentPassword == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "currentPasswordRequired"),
			})
			return
		}
		if !util.ComparePassword(user.Password, req.CurrentPassword) {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "currentPasswordIncorrect"),
			})
			return
		}
	}

	name := user.Name
	if req.Name != "" {
		name = req.Name
	}
	email := user.Email
	if req.Email != "" {
		email = req.Email
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	opts := &module.UpdateUserOptions{
		UserID:   user.ID,
		Name:     name,
		Email:    email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}
	if req.NewPassword != "" {
		opts.Password = req.NewPassword
	}

	updated, err := userModule.UpdateUser(opts)
	if err != nil {
		if err == module.ErrUserEmailAlreadyExists {
			util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": locale.TR(r, "accountWithEmailExists"),
			})
			return
		}
		log.Error().
			Err(err).
			Msg("Failed to update profile")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateProfile"),
		})
		return
	}

	log.Info().
		Str("userID", updated.ID.String()).
		Msg("Profile updated successfully")
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":          updated.ID,
			"email":       updated.Email,
			"role":        updated.Role,
			"isActive":    updated.IsActive,
			"name":        updated.Name,
			"avatar":      util.GravatarURL(updated.Email, 80),
			"lastLoginAt": updated.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":   updated.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   updated.UpdatedAt.UTC().Format(time.RFC3339),
		},
		"successMessage": locale.TR(r, "profileUpdatedSuccessfully"),
	})
}
