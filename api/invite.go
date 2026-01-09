// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/tivo/db"
	"github.com/clivern/tivo/middleware"
	"github.com/clivern/tivo/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	inviteExpiryDays    = 7
	inviteStatusPending = "pending"
)

// CreateUserInviteRequest represents the create user invite request payload
type CreateUserInviteRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=255" label:"Email"`
	Role  string `json:"role" validate:"required,oneof=admin regular" label:"Role"`
}

// CreateUserInviteAction handles user invite creation requests
func CreateUserInviteAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create user invite endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	var req CreateUserInviteRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	existingUser, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing user for invite")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create invite",
		})
		return
	}
	if existingUser != nil {
		service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": "A user with this email is already registered",
		})
		return
	}

	inviteRepo := db.NewUserInviteRepository(db.GetDB())
	count, err := inviteRepo.CountPendingByEmail(req.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count pending invites for email")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create invite",
		})
		return
	}
	if count > 0 {
		service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": "A pending invite already exists for this email",
		})
		return
	}

	tokenUUID, _ := uuid.NewRandom()
	token := tokenUUID.String()

	expiresAt := time.Now().UTC().Add(inviteExpiryDays * 24 * time.Hour)
	invite := &db.UserInvite{
		Email:         req.Email,
		Role:          req.Role,
		Token:         token,
		Status:        inviteStatusPending,
		InviterUserID: user.ID,
		ExpiresAt:     expiresAt,
	}
	if err := inviteRepo.Create(invite); err != nil {
		log.Error().Err(err).Msg("Failed to create user invite")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create invite",
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

	inviteLink := strings.TrimSuffix(platformURL, "/") + "/register?token=" + url.QueryEscape(invite.Token)
	mailer := service.NewMailer(optRepo)
	if err := mailer.SendInviteEmail(req.Email, inviteLink, platformName); err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("Failed to send invite email; invite was created, link available in UI")
	}

	log.Info().
		Str("inviteID", invite.ID.String()).
		Str("email", req.Email).
		Msg("User invite created successfully")

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        invite.ID,
		"email":     invite.Email,
		"role":      invite.Role,
		"token":     invite.Token,
		"status":    invite.Status,
		"expiresAt": invite.ExpiresAt.UTC().Format(time.RFC3339),
		"createdAt": invite.CreatedAt.UTC().Format(time.RFC3339),
	})
}

// ListInvitesAction handles listing all user invites (admin only). Admins see all invites.
func ListInvitesAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List invites endpoint called")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 50
	offset := 0
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	repo := db.NewUserInviteRepository(db.GetDB())
	total, err := repo.Count()
	if err != nil {
		log.Error().Err(err).Msg("Failed to count invites")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list invites",
		})
		return
	}

	repo.MarkExpiredAsExpired()
	invites, err := repo.ListPaginated(limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list invites")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list invites",
		})
		return
	}

	inviteList := make([]map[string]interface{}, 0, len(invites))
	for _, inv := range invites {
		item := map[string]interface{}{
			"id":            inv.ID,
			"email":         inv.Email,
			"role":          inv.Role,
			"token":         inv.Token,
			"status":        inv.Status,
			"inviterUserId": inv.InviterUserID,
			"expiresAt":     inv.ExpiresAt.UTC().Format(time.RFC3339),
			"createdAt":     inv.CreatedAt.UTC().Format(time.RFC3339),
		}
		if inv.AcceptedAt != nil {
			item["acceptedAt"] = inv.AcceptedAt.UTC().Format(time.RFC3339)
		}
		inviteList = append(inviteList, item)
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"invites": inviteList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetUserInviteAction handles get user invite by ID requests
func GetUserInviteAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get user invite endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid invite ID",
		})
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())

	repo.MarkExpiredAsExpired()
	invite, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user invite")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get invite",
		})
		return
	}
	if invite == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Invite not found",
		})
		return
	}

	resp := map[string]interface{}{
		"id":        invite.ID,
		"email":     invite.Email,
		"role":      invite.Role,
		"status":    invite.Status,
		"expiresAt": invite.ExpiresAt.UTC().Format(time.RFC3339),
		"createdAt": invite.CreatedAt.UTC().Format(time.RFC3339),
	}
	if invite.AcceptedAt != nil {
		resp["acceptedAt"] = invite.AcceptedAt.UTC().Format(time.RFC3339)
	}
	service.WriteJSON(w, http.StatusOK, resp)
}

// DeleteUserInviteAction handles user invite deletion requests
func DeleteUserInviteAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete user invite endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid invite ID",
		})
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())
	err := repo.Delete(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user invite")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete invite",
		})
		return
	}

	log.Info().Str("inviteID", idStr).Msg("User invite deleted successfully")
	service.WriteJSON(w, http.StatusNoContent, nil)
}

// GetUserInviteByTokenAction returns invite details by token (public, for accept page)
func GetUserInviteByTokenAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get user invite by token endpoint called")

	token := chi.URLParam(r, "token")
	if token == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid token",
		})
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())

	repo.MarkExpiredAsExpired()
	invite, err := repo.GetByToken(token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user invite by token")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get invite",
		})
		return
	}
	if invite == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Invite not found",
		})
		return
	}
	if invite.Status != inviteStatusPending {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invite is no longer valid",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"email":     invite.Email,
		"role":      invite.Role,
		"expiresAt": invite.ExpiresAt.UTC().Format(time.RFC3339),
	})
}
