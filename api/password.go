// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/clivern/lapin/db"
	"github.com/clivern/lapin/module"
	"github.com/clivern/lapin/service"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

// ForgotPasswordRequest represents the forgot password request body
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=255" label:"Email"`
}

// ResetPasswordRequest represents the reset password request body
type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,min=1" label:"Token"`
	Password string `json:"password" validate:"required,strong_password,min=8,max=60" label:"Password"`
}

// ForgotPasswordAction handles forgot password requests.
func ForgotPasswordAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Forgot password endpoint called")

	var req ForgotPasswordRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	user, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to look up user for forgot password")
		service.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"successMessage": "If an account exists for this email, you will receive a password reset link shortly.",
		})
		return
	}
	if user == nil {
		service.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"successMessage": "If an account exists for this email, you will receive a password reset link shortly.",
		})
		return
	}

	tokenUUID, _ := uuid.NewRandom()
	token := tokenUUID.String()

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	resetRepo := db.NewPasswordResetTokenRepository(db.GetDB())
	resetToken := &db.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := resetRepo.Create(resetToken); err != nil {
		log.Error().Err(err).Msg("Failed to create password reset token")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to process request",
		})
		return
	}

	optRepo := db.NewOptionRepository(db.GetDB())
	var platformURL, platformName string
	if o, _ := optRepo.Get("platform_url"); o != nil {
		platformURL = strings.TrimSpace(o.Value)
	}
	if o, _ := optRepo.Get("platform_name"); o != nil {
		platformName = strings.TrimSpace(o.Value)
	}
	resetLink := strings.TrimSuffix(platformURL, "/") + "/reset-password/" + url.PathEscape(token)
	mailer := service.NewMailer(optRepo)

	if err := mailer.SendPasswordResetEmail(user.Email, resetLink, platformName); err != nil {
		log.Warn().Err(err).Str("email", user.Email).Msg("Failed to send password reset email; token was created")
	}

	log.Info().
		Str("userID", user.ID.String()).
		Str("email", user.Email).
		Msg("Password reset token created")

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "If an account exists for this email, you will receive a password reset link shortly.",
	})
}

// ResetPasswordAction handles reset password requests (token + new password)
func ResetPasswordAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Reset password endpoint called")

	var req ResetPasswordRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	resetRepo := db.NewPasswordResetTokenRepository(db.GetDB())
	tok, err := resetRepo.GetByToken(req.Token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to look up password reset token")
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "This reset link is invalid or has expired.",
		})
		return
	}
	if tok == nil || time.Now().UTC().After(tok.ExpiresAt) {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "This reset link is invalid or has expired.",
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	user, err := userRepo.GetByID(tok.UserID)
	if err != nil || user == nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "This reset link is invalid or has expired.",
		})
		return
	}

	userModule := module.NewUser(userRepo)
	_, err = userModule.UpdateUser(&module.UpdateUserOptions{
		UserID:   tok.UserID,
		Name:     user.Name,
		Email:    user.Email,
		Password: req.Password,
		Role:     user.Role,
		IsActive: user.IsActive,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update password on reset")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to reset password",
		})
		return
	}

	if err := resetRepo.DeleteByToken(req.Token); err != nil {
		log.Error().Err(err).Msg("Failed to delete password reset token")
	}

	log.Info().Str("userID", tok.UserID.String()).Msg("Password reset successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "Your password has been reset successfully.",
	})
}
