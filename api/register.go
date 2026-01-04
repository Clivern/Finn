// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/pkg/util"

	"github.com/rs/zerolog/log"
)

const inviteStatusAccepted = "accepted"

// RegisterRequest is the body for signing up.
type RegisterRequest struct {
	Token    string `json:"token" validate:"required,min=1" label:"Token"`
	Name     string `json:"name" validate:"required,min=1,max=255" label:"Name"`
	Password string `json:"password" validate:"required,strong_password,min=8,max=60" label:"Password"`
}

// RegisterAction creates an account using a valid invite token.
func RegisterAction(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("token", req.Token).
		Msg("Register endpoint called")

	repo := db.NewUserInviteRepository(db.GetDB())
	invite, err := repo.GetByToken(req.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("token", req.Token).
			Msg("Failed to get user invite by token")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCompleteRegistration"),
		})
		return
	}
	if invite == nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidOrExpiredInviteToken"),
		})
		return
	}

	if invite.Status != inviteStatusPending {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "inviteNoLongerValid"),
		})
		return
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		repo.MarkExpiredAsExpired()
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "inviteExpired"),
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	existing, err := userRepo.GetByEmail(invite.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("inviteID", invite.ID.String()).
			Str("token", req.Token).
			Msg("Failed to check existing user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCompleteRegistration"),
		})
		return
	}
	if existing != nil {
		util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": locale.TR(r, "accountWithEmailAlreadyExists"),
		})
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Str("inviteID", invite.ID.String()).
			Msg("Failed to hash password")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCompleteRegistration"),
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
		log.Error().
			Err(err).
			Str("inviteID", invite.ID.String()).
			Str("email", user.Email).
			Msg("Failed to create user")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCompleteRegistration"),
		})
		return
	}

	if err := repo.UpdateStatus(invite.ID, inviteStatusAccepted, &now); err != nil {
		log.Error().
			Err(err).
			Str("inviteID", invite.ID.String()).
			Str("userID", user.ID.String()).
			Msg("Failed to mark invite as accepted")
		// User already created; continue and return success
	}

	log.Info().
		Str("userID", user.ID.String()).
		Str("email", user.Email).
		Msg("User registered successfully")

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": locale.TR(r, "registrationSuccessful"),
	})
}
