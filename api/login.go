// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// LoginRequest is the body you send to log in.
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email" label:"Email"`
	Password   string `json:"password" validate:"required" label:"Password"`
	RememberMe bool   `json:"rememberMe" validate:"omitempty,boolean" label:"Remember Me"`
}

// LoginAction logs the user in and creates a session.
func LoginAction(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("email", req.Email).
		Msg("Login endpoint called")

	userRepo := db.NewUserRepository(db.GetDB())
	sessionRepo := db.NewSessionRepository(db.GetDB())
	authModule := module.NewAuth(userRepo)
	user, err := authModule.Login(req.Email, req.Password)
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidUsernameOrPassword"),
		})
		return
	}

	if !user.IsActive {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "accountNotActive"),
		})
		return
	}

	optionRepo := db.NewOptionRepository(db.GetDB())
	maintenanceOpt, err := optionRepo.Get("maintenance_mode")

	if err != nil {
		log.Error().
			Err(err).
			Str("userID", user.ID.String()).
			Msg("Failed to get maintenance_mode option")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckMaintenance"),
		})
		return
	}

	if maintenanceOpt != nil && maintenanceOpt.Value == "1" && user.Role != db.UserRoleAdmin {
		util.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"errorMessage": locale.TR(r, "maintenanceModeEnabled"),
		})
		return
	}

	sessionManager := module.NewSessionManager(sessionRepo, userRepo)
	session, err := sessionManager.CreateSession(
		user.ID,
		time.Hour*24*7,
		r.RemoteAddr,
		r.UserAgent(),
	)
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateSession"),
		})
		return
	}

	var cookieOptions *util.CookieOptions
	if viper.GetBool("app.tls.status") || strings.HasPrefix(viper.GetString("app.hostname"), "https://") {
		cookieOptions = util.SecureCookieOptions()
	} else {
		cookieOptions = util.DefaultCookieOptions()
	}
	if req.RememberMe {
		cookieOptions.MaxAge = int((time.Hour * 24 * 30) / time.Second)
	} else {
		cookieOptions.MaxAge = 0
	}

	util.SetCookie(w, "_finn_session", session.Token, cookieOptions)
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": locale.TR(r, "loginSuccessful"),
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
