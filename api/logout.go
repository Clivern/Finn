// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/service/util"

	"github.com/rs/zerolog/log"
)

// LogoutAction handles logout requests.
func LogoutAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Logout endpoint called")

	util.DeleteCookie(w, "_finn_session")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"successMessage": locale.TR(r, "logoutSuccessful"),
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
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": locale.TR(r, "logoutSuccessful"),
	})
}
