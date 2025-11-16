// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/lapin/db"
	"github.com/clivern/lapin/middleware"
	"github.com/clivern/lapin/module"
	"github.com/clivern/lapin/service"

	"github.com/rs/zerolog/log"
)

// LogoutAction handles logout requests.
func LogoutAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Logout endpoint called")

	service.DeleteCookie(w, "_lapin_session")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"successMessage": "Logout successful",
		})
		return
	}

	sessionManager := module.NewSessionManager(
		db.NewSessionRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	sessionManager.CleanupExpiredSessions()
	if err := sessionManager.RevokeUserSessions(user.ID); err != nil {
		log.Error().Err(err).Msg("Failed to revoke session")
	}
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "Logout successful",
	})
}
