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
	"github.com/clivern/finn/service/runner"
	"github.com/clivern/finn/service/util"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// CreateRunnerRequest represents the create runner request payload
type CreateRunnerRequest struct {
	Name             string `json:"name" validate:"required" label:"Name"`
	Tags             string `json:"tags" label:"Tags"`
	MaxParallelTasks int    `json:"maxParallelTasks" label:"Max parallel tasks"`
	Status           string `json:"status" validate:"omitempty,oneof=active inactive maintenance" label:"Status"`
	URL              string `json:"url" validate:"required" label:"URL"`
	Token            string `json:"token" validate:"required,min=1" label:"Token (UUID)"`
	PublicKey        string `json:"publicKey" label:"Public key"`
}

// UpdateRunnerRequest represents the update runner request payload
type UpdateRunnerRequest struct {
	Name             *string `json:"name" label:"Name"`
	Tags             *string `json:"tags" label:"Tags"`
	MaxParallelTasks *int    `json:"maxParallelTasks" label:"Max parallel tasks"`
	Status           *string `json:"status" validate:"omitempty,oneof=active inactive maintenance" label:"Status"`
	URL              *string `json:"url" label:"URL"`
	Token            *string `json:"token" label:"Token (UUID)"`
	PublicKey        *string `json:"publicKey" label:"Public key"`
}

// ListRunnersAction handles listing all runners (platform admin only)
func ListRunnersAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List runners endpoint called")

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

	repo := db.NewRunnerRepository(db.GetDB())
	total, err := repo.Count()
	if err != nil {
		log.Error().Err(err).Msg("Failed to count runners")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListRunners"),
		})
		return
	}

	runners, err := repo.List(limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list runners")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListRunners"),
		})
		return
	}

	list := make([]map[string]interface{}, 0, len(runners))
	for _, run := range runners {
		item := map[string]interface{}{
			"id":               run.ID,
			"token":            run.Token,
			"name":             run.Name,
			"tags":             run.Tags,
			"maxParallelTasks": run.MaxParallelTasks,
			"status":           run.Status,
			"url":              run.URL,
			"publicKey":        run.PublicKey,
			"createdAt":        run.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":        run.UpdatedAt.UTC().Format(time.RFC3339),
		}
		list = append(list, item)
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"runners": list,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetRunnerAction handles get runner by ID (platform admin only)
func GetRunnerAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get runner endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	repo := db.NewRunnerRepository(db.GetDB())
	runner, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetRunner"),
		})
		return
	}
	if runner == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerNotFound"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":               runner.ID,
		"token":            runner.Token,
		"name":             runner.Name,
		"tags":             runner.Tags,
		"maxParallelTasks": runner.MaxParallelTasks,
		"status":           runner.Status,
		"url":              runner.URL,
		"publicKey":        runner.PublicKey,
		"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// CreateRunnerAction handles runner creation (platform admin only)
func CreateRunnerAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create runner endpoint called")

	var req CreateRunnerRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	token := strings.TrimSpace(req.Token)
	if _, err := uuid.Parse(token); err != nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "tokenMustBeValidUUID"),
		})
		return
	}

	repo := db.NewRunnerRepository(db.GetDB())
	if existing, _ := repo.GetByToken(token); existing != nil {
		util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerWithTokenExists"),
		})
		return
	}

	status := db.RunnerStatusActive
	if req.Status != "" {
		status = req.Status
	}

	if req.MaxParallelTasks <= 0 {
		req.MaxParallelTasks = 10
	}

	runner := &db.Runner{
		Token:            token,
		Name:             &req.Name,
		URL:              &req.URL,
		MaxParallelTasks: req.MaxParallelTasks,
		Status:           status,
	}
	if req.Tags != "" {
		runner.Tags = &req.Tags
	}
	if req.PublicKey != "" {
		runner.PublicKey = &req.PublicKey
	}

	if err := repo.Create(runner); err != nil {
		log.Error().Err(err).Msg("Failed to create runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateRunner"),
		})
		return
	}

	log.Info().Str("runnerID", runner.ID.String()).Msg("Runner created successfully")
	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":               runner.ID,
		"token":            token,
		"name":             runner.Name,
		"tags":             runner.Tags,
		"maxParallelTasks": runner.MaxParallelTasks,
		"status":           runner.Status,
		"url":              runner.URL,
		"publicKey":        runner.PublicKey,
		"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateRunnerAction handles runner update (platform admin only)
func UpdateRunnerAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update runner endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	var req UpdateRunnerRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	repo := db.NewRunnerRepository(db.GetDB())
	runner, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get runner for update")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateRunner"),
		})
		return
	}
	if runner == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerNotFound"),
		})
		return
	}

	if req.Token != nil {
		token := strings.TrimSpace(*req.Token)
		if token == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "tokenCannotBeEmpty"),
			})
			return
		}
		if _, err := uuid.Parse(token); err != nil {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "tokenMustBeValidUUID"),
			})
			return
		}
		existing, _ := repo.GetByToken(token)
		if existing != nil && existing.ID != runner.ID {
			util.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": locale.TR(r, "runnerWithTokenExists"),
			})
			return
		}
		runner.Token = token
	}
	if req.Name != nil {
		runner.Name = req.Name
	}
	if req.Tags != nil {
		runner.Tags = req.Tags
	}
	if req.MaxParallelTasks != nil && *req.MaxParallelTasks > 0 {
		runner.MaxParallelTasks = *req.MaxParallelTasks
	}
	if req.Status != nil {
		runner.Status = *req.Status
	}
	if req.URL != nil {
		runner.URL = req.URL
	}
	if req.PublicKey != nil {
		runner.PublicKey = req.PublicKey
	}

	if err := repo.Update(runner); err != nil {
		log.Error().Err(err).Msg("Failed to update runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateRunner"),
		})
		return
	}

	log.Info().Str("runnerID", runner.ID.String()).Msg("Runner updated successfully")
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":               runner.ID,
		"token":            runner.Token,
		"name":             runner.Name,
		"tags":             runner.Tags,
		"maxParallelTasks": runner.MaxParallelTasks,
		"status":           runner.Status,
		"url":              runner.URL,
		"publicKey":        runner.PublicKey,
		"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteRunnerAction handles runner deletion (platform admin only)
func DeleteRunnerAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete runner endpoint called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	repo := db.NewRunnerRepository(db.GetDB())
	existing, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get runner for delete")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteRunner"),
		})
		return
	}
	if existing == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerNotFound"),
		})
		return
	}

	if err := repo.Delete(db.ID(idStr)); err != nil {
		log.Error().Err(err).Msg("Failed to delete runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteRunner"),
		})
		return
	}

	log.Info().Str("runnerID", idStr).Msg("Runner deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}

// PingRunnerAction pings a runner by ID using its URL, public key and token, and returns status.
func PingRunnerAction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	repo := db.NewRunnerRepository(db.GetDB())
	run, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get runner for ping")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetRunnerForPing"),
		})
		return
	}
	if run == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerNotFound"),
		})
		return
	}

	if run.URL == nil || strings.TrimSpace(*run.URL) == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerUrlNotConfigured"),
		})
		return
	}
	if run.PublicKey == nil || strings.TrimSpace(*run.PublicKey) == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "runnerPublicKeyNotConfigured"),
		})
		return
	}

	client := &runner.Client{
		BaseURL:      strings.TrimSpace(*run.URL),
		PublicKeyPEM: []byte(*run.PublicKey),
		Token:        run.Token,
	}
	if err := client.Ping(r.Context()); err != nil {
		log.Info().Err(err).Str("runnerID", idStr).Msg("Runner ping failed")
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "failed",
			"errorMessage": err.Error(),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "ok",
		"errorMessage": "",
	})
}
