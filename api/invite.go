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

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	inviteExpiryDays    = 7
	inviteStatusPending = "pending"
)

// CreateUserInviteRequest is the body you send when creating an invite.
type CreateUserInviteRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=255" label:"Email"`
	Role  string `json:"role" validate:"required,oneof=admin regular" label:"Role"`
}

// CreateUserInviteAction creates a new user invite.
func CreateUserInviteAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create user invite endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	var req CreateUserInviteRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())
	existingUser, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to check existing user for invite")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateInvite"),
		})
		return
	}
	if existingUser != nil {
		util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": locale.TR(r, "userWithEmailAlreadyRegistered"),
		})
		return
	}

	inviteRepo := db.NewUserInviteRepository(db.GetDB())
	count, err := inviteRepo.CountPendingByEmail(req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to count pending invites for email")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateInvite"),
		})
		return
	}
	if count > 0 {
		util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": locale.TR(r, "pendingInviteAlreadyExists"),
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
		log.Error().
			Err(err).
			Msg("Failed to create user invite")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateInvite"),
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
	mailer := util.NewMailer(optRepo)
	if err := mailer.SendInviteEmail(req.Email, inviteLink, platformName); err != nil {
		log.Warn().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to send invite email; invite was created, link available in UI")
	}

	log.Info().
		Str("inviteID", invite.ID.String()).
		Str("email", req.Email).
		Msg("User invite created successfully")

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        invite.ID,
		"email":     invite.Email,
		"role":      invite.Role,
		"token":     invite.Token,
		"status":    invite.Status,
		"expiresAt": invite.ExpiresAt.UTC().Format(time.RFC3339),
		"createdAt": invite.CreatedAt.UTC().Format(time.RFC3339),
	})
}

// ListInvitesAction returns all invites (admin only).
func ListInvitesAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List invites endpoint called")

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
		log.Error().
			Err(err).
			Msg("Failed to count invites")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListInvites"),
		})
		return
	}

	repo.MarkExpiredAsExpired()
	invites, err := repo.ListPaginated(limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list invites")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListInvites"),
		})
		return
	}

	list := make([]map[string]interface{}, 0, len(invites))
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
			"updatedAt":     inv.UpdatedAt.UTC().Format(time.RFC3339),
		}
		if inv.AcceptedAt != nil {
			item["acceptedAt"] = inv.AcceptedAt.UTC().Format(time.RFC3339)
		}
		list = append(list, item)
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"invites": list,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetUserInviteAction returns one invite by ID.
func GetUserInviteAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get user invite endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidInviteID"),
		})
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())

	repo.MarkExpiredAsExpired()
	invite, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get user invite")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetInvite"),
		})
		return
	}
	if invite == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "inviteNotFound"),
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
		"updatedAt": invite.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if invite.AcceptedAt != nil {
		resp["acceptedAt"] = invite.AcceptedAt.UTC().Format(time.RFC3339)
	}
	util.WriteJSON(w, http.StatusOK, resp)
}

// DeleteUserInviteAction deletes an invite by ID.
func DeleteUserInviteAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete user invite endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidInviteID"),
		})
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())
	err := repo.Delete(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete user invite")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteInvite"),
		})
		return
	}

	log.Info().
		Str("inviteID", idStr).
		Msg("User invite deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, nil)
}

// GetUserInviteByTokenAction returns invite details by token (used on the signup page).
func GetUserInviteByTokenAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get user invite by token endpoint called")

	token := chi.URLParam(r, "token")
	if token == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidToken"),
		})
		return
	}

	repo := db.NewUserInviteRepository(db.GetDB())

	repo.MarkExpiredAsExpired()
	invite, err := repo.GetByToken(token)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get user invite by token")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetInvite"),
		})
		return
	}
	if invite == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "inviteNotFound"),
		})
		return
	}
	if invite.Status != inviteStatusPending {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "inviteNoLongerValid"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":        invite.ID,
		"email":     invite.Email,
		"role":      invite.Role,
		"status":    invite.Status,
		"expiresAt": invite.ExpiresAt.UTC().Format(time.RFC3339),
		"createdAt": invite.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": invite.UpdatedAt.UTC().Format(time.RFC3339),
	})
}
