// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/pkg/ai"
	"github.com/clivern/finn/pkg/util"

	"github.com/rs/zerolog/log"
)

// CreateMessageRequest is the body for sending an AI message.
type CreateMessageRequest struct {
	Content string `json:"content" validate:"required,max=10000" label:"Message content"`
}

// GetStatusAction returns whether AI is enabled.
func GetStatusAction(w http.ResponseWriter, _ *http.Request) {
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": ai.IsEnabled(),
	})
}

// GetMessagesAction returns AI messages for the current user.
func GetMessagesAction(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	repo := db.NewAIMessageRepository(db.GetDB())
	messages, err := repo.GetByUserID(user.ID, limit, offset)
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedLoadMessages"),
		})
		return
	}

	total, _ := repo.CountByUserID(user.ID)
	list := make([]map[string]interface{}, 0, len(messages))
	for _, m := range messages {
		list = append(list, map[string]interface{}{
			"id":        m.ID,
			"role":      m.Role,
			"content":   m.Content,
			"createdAt": m.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"messages": list,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// CreateMessageAction creates a new AI message.
func CreateMessageAction(w http.ResponseWriter, r *http.Request) {
	if !ai.IsEnabled() {
		util.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"errorMessage": locale.TR(r, "aiNotEnabled"),
		})
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	var req CreateMessageRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "messageContentRequired"),
		})
		return
	}

	repo := db.NewAIMessageRepository(db.GetDB())

	userMsg := &db.AIMessage{UserID: user.ID, Role: "user", Content: content}
	if err := repo.Create(userMsg); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedSaveMessage"),
		})
		return
	}

	total, _ := repo.CountByUserID(user.ID)
	contextLimit := 20
	offset := 0
	if total > int64(contextLimit) {
		offset = int(total) - contextLimit
	}
	recent, err := repo.GetByUserID(user.ID, contextLimit, offset)
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedLoadConversation"),
		})
		return
	}

	var contextParts []string
	for i := 0; i < len(recent)-1; i++ {
		m := recent[i]
		contextParts = append(contextParts, m.Role+": "+m.Content)
	}
	contextStr := strings.Join(contextParts, "\n")

	client, err := ai.NewFromConfig(r.Context())
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "aiNotConfigured"),
		})
		return
	}

	input := ai.MainPrompt(content, contextStr)
	input.Tools = client.Tools()
	input.UserID = user.ID

	response, err := client.Generate(r.Context(), ai.ModelName(), input)
	if err != nil {
		log.Error().
			Err(err).
			Str("userID", user.ID.String()).
			Msg("AI generate failed")
		msg := locale.TR(r, "failedGenerateResponse")
		if errStr := err.Error(); errStr != "" && len(errStr) < 500 {
			msg = errStr
		}
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": msg,
		})
		return
	}

	msg := &db.AIMessage{UserID: user.ID, Role: "assistant", Content: response}
	if err := repo.Create(msg); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedSaveResponse"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":        msg.ID,
		"role":      msg.Role,
		"content":   msg.Content,
		"createdAt": msg.CreatedAt.UTC().Format(time.RFC3339),
	})
}
