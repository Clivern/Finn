// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// CreateRunnerRequest is the body for creating a runner.
type CreateRunnerRequest struct {
	Name             string            `json:"name" validate:"required" label:"Name"`
	Tags             map[string]string `json:"tags" label:"Tags"`
	MaxParallelTasks int               `json:"maxParallelTasks" validate:"required,min=1" label:"Max parallel tasks"`
	Status           string            `json:"status" validate:"required,oneof=active inactive maintenance" label:"Status"`
	URL              string            `json:"url" validate:"required" label:"URL"`
	Token            string            `json:"token" validate:"required,min=1" label:"Token (UUID)"`
}

// UpdateRunnerRequest is the body for updating a runner.
type UpdateRunnerRequest struct {
	Name             string            `json:"name" validate:"required" label:"Name"`
	Tags             map[string]string `json:"tags" label:"Tags"`
	MaxParallelTasks int               `json:"maxParallelTasks" validate:"required,min=1" label:"Max parallel tasks"`
	Status           string            `json:"status" validate:"required,oneof=active inactive maintenance" label:"Status"`
	URL              string            `json:"url" validate:"required" label:"URL"`
	Token            string            `json:"token" validate:"required,min=1" label:"Token (UUID)"`
}

// ListRunnersAction returns all runners (admin only).
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
		log.Error().
			Err(err).
			Int("limit", limit).
			Int("offset", offset).
			Msg("Failed to count runners")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListRunners"),
		})
		return
	}

	result, err := repo.List(limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Int("limit", limit).
			Int("offset", offset).
			Msg("Failed to list runners")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListRunners"),
		})
		return
	}

	runners := make([]map[string]interface{}, 0, len(result))
	for _, runner := range result {
		runners = append(runners, map[string]interface{}{
			"id":               runner.ID,
			"token":            runner.Token,
			"name":             runner.Name,
			"tags":             runnerTagsFromStr(runner.Tags),
			"maxParallelTasks": runner.MaxParallelTasks,
			"status":           runner.Status,
			"url":              runner.URL,
			"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"runners": runners,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetRunnerAction returns one runner by ID (admin only).
func GetRunnerAction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	log.Info().
		Str("runnerID", idStr).
		Msg("Get runner endpoint called")

	repo := db.NewRunnerRepository(db.GetDB())
	runner, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("runnerID", idStr).
			Msg("Failed to get runner")
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
		"tags":             runnerTagsFromStr(runner.Tags),
		"maxParallelTasks": runner.MaxParallelTasks,
		"status":           runner.Status,
		"url":              runner.URL,
		"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// CreateRunnerAction creates a runner (admin only).
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
		Name:             req.Name,
		URL:              req.URL,
		MaxParallelTasks: req.MaxParallelTasks,
		Status:           status,
		Tags:             runnerTagsToJSON(normalizeRunnerTags(req.Tags)),
	}

	if err := repo.Create(runner); err != nil {
		log.Error().
			Err(err).
			Str("name", req.Name).
			Str("url", req.URL).
			Msg("Failed to create runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateRunner"),
		})
		return
	}

	log.Info().
		Str("runnerID", runner.ID.String()).
		Msg("Runner created successfully")
	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":               runner.ID,
		"token":            token,
		"name":             runner.Name,
		"tags":             runnerTagsFromStr(runner.Tags),
		"maxParallelTasks": runner.MaxParallelTasks,
		"status":           runner.Status,
		"url":              runner.URL,
		"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateRunnerAction updates a runner (admin only).
func UpdateRunnerAction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	log.Info().
		Str("runnerID", idStr).
		Msg("Update runner endpoint called")

	var req UpdateRunnerRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	repo := db.NewRunnerRepository(db.GetDB())
	runner, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("runnerID", idStr).
			Msg("Failed to get runner for update")
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

	if req.Token != "" {
		token := strings.TrimSpace(req.Token)
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

	runner.Name = req.Name
	runner.Tags = runnerTagsToJSON(normalizeRunnerTags(req.Tags))
	runner.MaxParallelTasks = req.MaxParallelTasks
	runner.Status = req.Status
	runner.URL = req.URL

	if err := repo.Update(runner); err != nil {
		log.Error().
			Err(err).
			Str("runnerID", idStr).
			Msg("Failed to update runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateRunner"),
		})
		return
	}

	log.Info().
		Str("runnerID", runner.ID.String()).
		Msg("Runner updated successfully")
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":               runner.ID,
		"token":            runner.Token,
		"name":             runner.Name,
		"tags":             runnerTagsFromStr(runner.Tags),
		"maxParallelTasks": runner.MaxParallelTasks,
		"status":           runner.Status,
		"url":              runner.URL,
		"createdAt":        runner.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        runner.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteRunnerAction deletes a runner (admin only).
func DeleteRunnerAction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRunnerID"),
		})
		return
	}

	log.Info().
		Str("runnerID", idStr).
		Msg("Delete runner endpoint called")

	repo := db.NewRunnerRepository(db.GetDB())
	existing, err := repo.GetByID(db.ID(idStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("runnerID", idStr).
			Msg("Failed to get runner for delete")
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
		log.Error().
			Err(err).
			Str("runnerID", idStr).
			Msg("Failed to delete runner")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteRunner"),
		})
		return
	}

	log.Info().
		Str("runnerID", idStr).
		Msg("Runner deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}

// PingRunnerAction pings a runner to check it's reachable.
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
		log.Error().
			Err(err).
			Str("runnerID", idStr).
			Msg("Failed to get runner for ping")
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

	client, err := runner.NewGRPCClient(
		r.Context(),
		strings.TrimSpace(run.URL),
		run.Token,
	)

	if err != nil {
		log.Info().
			Err(err).
			Str("runnerID", idStr).
			Msg("Runner ping: grpc client failed")
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}
	defer client.Close()

	if err := client.Ping(r.Context()); err != nil {
		log.Info().
			Err(err).
			Str("runnerID", idStr).
			Msg("Runner ping failed")
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "ok",
		"errorMessage": "",
	})
}

func normalizeRunnerTags(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string)
	for k, v := range in {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		out[k] = strings.TrimSpace(v)
	}
	return out
}

func runnerTagsToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func runnerTagsFromStr(s string) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return map[string]string{}
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil || m == nil {
		return map[string]string{}
	}
	return m
}
