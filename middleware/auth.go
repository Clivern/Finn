// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"

	"github.com/rs/zerolog/log"
)

const (
	// ContextKeyUser is the key for storing user in context
	ContextKeyUser contextKey = "user"
)

// SessionAuth creates a session-based authentication middleware
func SessionAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkipAuth(r.URL.Path) {
				log.Info().
					Str("path", r.URL.Path).
					Msg("Skipping authentication for API route")
				next.ServeHTTP(w, r)
				return
			}

			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				user, err := db.NewUserRepository(db.GetDB()).GetByAPIKey(apiKey)
				if err != nil {
					log.Info().
						Err(err).
						Str("path", r.URL.Path).
						Msg("API key validation failed")
					util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
						"errorMessage": "Invalid API key",
					})
					return
				}
				log.Info().
					Str("path", r.URL.Path).
					Msg("API key validation successful")
				ctx := context.WithValue(r.Context(), ContextKeyUser, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			sessionToken := util.GetCookie(r, "_finn_session")
			if sessionToken == "" {
				log.Info().
					Str("path", r.URL.Path).
					Msg("No session cookie found")
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": "Not authenticated",
				})
				return
			}

			sessionManager := module.NewSessionManager(
				db.NewSessionRepository(db.GetDB()),
				db.NewUserRepository(db.GetDB()),
			)

			user, _, err := sessionManager.ValidateSession(sessionToken)
			if err != nil {
				util.DeleteCookie(w, "_finn_session")
				if user != nil {
					sessionManager.RevokeUserSessions(user.ID)
				}
				log.Info().
					Err(err).
					Str("path", r.URL.Path).
					Msg("Session validation failed")
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": "Invalid or expired session",
				})
				return
			}

			log.Info().
				Str("path", r.URL.Path).
				Msg("Session validation successful")
			ctx := context.WithValue(r.Context(), ContextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// shouldSkipAuth determines if authentication should be skipped for a given path
func shouldSkipAuth(path string) bool {
	if strings.HasPrefix(path, "/api/v1/public/") || !strings.HasPrefix(path, "/api/v1/") {
		return true
	}

	return false
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	user, ok := ctx.Value(ContextKeyUser).(*db.User)
	return user, ok
}
