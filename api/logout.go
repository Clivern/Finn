// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/met/db"
	"github.com/clivern/met/middleware"
	"github.com/clivern/met/module"
	"github.com/clivern/met/service"

	"github.com/rs/zerolog/log"
)

// LogoutAction handles logout requests.
// Returns 200 even when not authenticated (idempotent: "already logged out").
func LogoutAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Logout endpoint called")

	service.DeleteCookie(w, "_met_session")

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
