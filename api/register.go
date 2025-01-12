// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/service"

	"github.com/rs/zerolog/log"
)

const inviteStatusAccepted = "accepted"

// RegisterRequest represents the register request body (invite token + name + password)
type RegisterRequest struct {
	Token    string `json:"token" validate:"required,min=1" label:"Token"`
	Name     string `json:"name" validate:"required,min=1,max=255" label:"Name"`
	Password string `json:"password" validate:"required,strong_password,min=8,max=60" label:"Password"`
}

// RegisterAction handles register requests. Requires a valid invite token.
func RegisterAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Register endpoint called")

	var req RegisterRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())
	invite, err := repo.GetByToken(req.Token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user invite by token")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to complete registration",
		})
		return
	}
	if invite == nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid or expired invite token",
		})
		return
	}

	if invite.Status != inviteStatusPending {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invite is no longer valid",
		})
		return
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		repo.MarkExpiredAsExpired()
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invite has expired",
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	existing, err := userRepo.GetByEmail(invite.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to complete registration",
		})
		return
	}
	if existing != nil {
		service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": "An account with this email already exists",
		})
		return
	}

	hashedPassword, err := service.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to complete registration",
		})
		return
	}

	now := time.Now().UTC()
	user := &db.User{
		Name:        req.Name,
		Email:       invite.Email,
		Password:    hashedPassword,
		Role:        invite.Role,
		IsActive:    true,
		LastLoginAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := userRepo.Create(user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to complete registration",
		})
		return
	}

	if err := repo.UpdateStatus(invite.ID, inviteStatusAccepted, &now); err != nil {
		log.Error().Err(err).Msg("Failed to mark invite as accepted")
		// User already created; continue and return success
	}

	log.Info().
		Str("userID", user.ID.String()).
		Str("email", user.Email).
		Msg("User registered successfully")

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "Registration successful. Please sign in.",
	})
}
