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
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// CreateUserAPIKeyRequest is the body for creating an API key (name, optional expiry).
type CreateUserAPIKeyRequest struct {
	Name      string `json:"name" validate:"omitempty,max=255" label:"Name"`
	ExpiresAt string `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
}

// CreateUserAPIKeyAction creates an API key; raw key is returned only once.
func CreateUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create user API key endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	var req CreateUserAPIKeyRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate API key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateApiKey"),
		})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "invalidExpiresAtFormat"),
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
		log.Error().
			Err(err).
			Msg("Failed to create API key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateApiKey"),
		})
		return
	}

	resp := map[string]interface{}{
		"id":        apiKey.ID,
		"name":      name,
		"expiresAt": nil,
		"createdAt": apiKey.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": apiKey.UpdatedAt.UTC().Format(time.RFC3339),
		"key":       id.String(),
	}
	if apiKey.ExpiresAt != nil {
		resp["expiresAt"] = apiKey.ExpiresAt.UTC().Format(time.RFC3339)
	}

	log.Info().
		Str("userID", user.ID.String()).
		Str("apiKeyID", apiKey.ID.String()).
		Msg("User API key created")
	util.WriteJSON(w, http.StatusCreated, resp)
}

// ListUserAPIKeysAction lists your API keys (metadata only, never the secret).
func ListUserAPIKeysAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List user API keys endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	repo := db.NewAPIKeyRepository(db.GetDB())

	if _, err := repo.DeleteExpired(); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to delete expired API keys")
	}
	keys, err := repo.GetByUserID(user.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list API keys")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListApiKeys"),
		})
		return
	}

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

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"apiKeys": list,
	})
}

// GetUserAPIKeyAction returns one API key's metadata (never the key itself).
func GetUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get user API key endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidApiKeyID"),
		})
		return
	}

	repo := db.NewAPIKeyRepository(db.GetDB())

	if _, err := repo.DeleteExpired(); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to delete expired API keys")
	}
	k, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get API key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetApiKey"),
		})
		return
	}
	if k == nil || k.UserID != user.ID {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "apiKeyNotFound"),
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
	util.WriteJSON(w, http.StatusOK, resp)
}

// DeleteUserAPIKeyAction deletes one of your API keys.
func DeleteUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete user API key endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidApiKeyID"),
		})
		return
	}

	repo := db.NewAPIKeyRepository(db.GetDB())
	k, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get API key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteApiKey"),
		})
		return
	}
	if k == nil || k.UserID != user.ID {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "apiKeyNotFound"),
		})
		return
	}

	if err := repo.Delete(k.ID); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete API key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteApiKey"),
		})
		return
	}

	log.Info().
		Str("userID", user.ID.String()).
		Str("apiKeyID", k.ID.String()).
		Msg("User API key deleted")
	w.WriteHeader(http.StatusNoContent)
}
