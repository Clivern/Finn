// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/tivo/db"
	"github.com/clivern/tivo/middleware"
	"github.com/clivern/tivo/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// CreateUserAPIKeyRequest represents the create API key request payload
type CreateUserAPIKeyRequest struct {
	Name      string `json:"name" validate:"omitempty,max=255" label:"Name"`
	ExpiresAt string `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
}

// CreateUserAPIKeyAction creates a new API key for the current user. The raw key is returned only once.
func CreateUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create user API key endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	var req CreateUserAPIKeyRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate API key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create API key",
		})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "Invalid expiresAt format",
			})
			return
		}
		expiresAt = &t
	}

	var name *string
	if req.Name != "" {
		name = &req.Name
	}

	apiKey := &db.APIKey{
		UserID:    user.ID,
		Name:      name,
		Key:       id.String(),
		ExpiresAt: expiresAt,
	}
	repo := db.NewAPIKeyRepository(db.GetDB())
	if err := repo.Create(apiKey); err != nil {
		log.Error().Err(err).Msg("Failed to create API key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create API key",
		})
		return
	}

	// Refetch to get createdAt from DB
	created, _ := repo.GetByID(apiKey.ID)
	if created == nil {
		created = apiKey
	}

	resp := map[string]interface{}{
		"id":        created.ID,
		"name":      name,
		"expiresAt": nil,
		"createdAt": created.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": created.UpdatedAt.UTC().Format(time.RFC3339),
		"key":       id.String(),
	}
	if created.ExpiresAt != nil {
		resp["expiresAt"] = created.ExpiresAt.UTC().Format(time.RFC3339)
	}

	log.Info().Str("userID", user.ID.String()).Str("apiKeyID", created.ID.String()).Msg("User API key created")
	service.WriteJSON(w, http.StatusCreated, resp)
}

// ListUserAPIKeysAction lists API keys for the current user (metadata only, never the key).
func ListUserAPIKeysAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List user API keys endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	repo := db.NewAPIKeyRepository(db.GetDB())

	if _, err := repo.DeleteExpired(); err != nil {
		log.Warn().Err(err).Msg("Failed to delete expired API keys")
	}
	keys, err := repo.GetByUserID(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list API keys")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list API keys",
		})
		return
	}

	// Only return non-expired keys, or all if you prefer; returning all so user can revoke expired
	list := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		item := map[string]interface{}{
			"id":        k.ID,
			"name":      k.Name,
			"createdAt": k.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt": k.UpdatedAt.UTC().Format(time.RFC3339),
		}
		if k.ExpiresAt != nil {
			item["expiresAt"] = k.ExpiresAt.UTC().Format(time.RFC3339)
		} else {
			item["expiresAt"] = nil
		}
		list = append(list, item)
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"apiKeys": list,
	})
}

// GetUserAPIKeyAction returns a single API key metadata for the current user.
func GetUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get user API key endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid API key ID",
		})
		return
	}

	repo := db.NewAPIKeyRepository(db.GetDB())

	if _, err := repo.DeleteExpired(); err != nil {
		log.Warn().Err(err).Msg("Failed to delete expired API keys")
	}
	k, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get API key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get API key",
		})
		return
	}
	if k == nil || k.UserID != user.ID {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "API key not found",
		})
		return
	}

	resp := map[string]interface{}{
		"id":        k.ID,
		"name":      k.Name,
		"createdAt": k.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": k.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if k.ExpiresAt != nil {
		resp["expiresAt"] = k.ExpiresAt.UTC().Format(time.RFC3339)
	} else {
		resp["expiresAt"] = nil
	}
	service.WriteJSON(w, http.StatusOK, resp)
}

// DeleteUserAPIKeyAction deletes an API key for the current user.
func DeleteUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete user API key endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid API key ID",
		})
		return
	}

	repo := db.NewAPIKeyRepository(db.GetDB())
	k, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get API key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete API key",
		})
		return
	}
	if k == nil || k.UserID != user.ID {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "API key not found",
		})
		return
	}

	if err := repo.Delete(k.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete API key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete API key",
		})
		return
	}

	log.Info().Str("userID", user.ID.String()).Str("apiKeyID", k.ID.String()).Msg("User API key deleted")
	w.WriteHeader(http.StatusNoContent)
}
