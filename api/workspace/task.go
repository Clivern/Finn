// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// GetWorkspaceRepositoryTaskAction returns one clone task's status and data.
func GetWorkspaceRepositoryTaskAction(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidWorkspaceID"),
		})
		return
	}

	repoIDStr := chi.URLParam(r, "id")
	if repoIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRepositoryID"),
		})
		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	if taskIDStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid task ID",
		})
		return
	}
	log.Info().
		Str("workspaceID", workspaceIDStr).
		Str("repositoryID", repoIDStr).
		Str("taskID", taskIDStr).
		Str("userID", user.ID.String()).
		Msg("Get workspace repository clone task endpoint called")

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "workspaceNotFound"),
			})
			return
		}
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", repoIDStr).
			Str("taskID", taskIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(repoIDStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", repoIDStr).
			Str("taskID", taskIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace repository")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetRepository"),
		})
		return
	}
	if repo == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "repositoryNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceRepositoryPermission(
		user,
		repo,
		module.CanCloneWorkspaceRepository,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", repo.ID.String()).
			Str("taskID", taskIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace permission")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCheckAccess"),
		})
		return
	}
	if !allowed {
		util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": locale.TR(r, "accessDeniedToWorkspace"),
		})
		return
	}

	cloneTaskRepo := db.NewWorkspaceRepositoryCloneTaskRepository(db.GetDB())
	task, err := cloneTaskRepo.GetByID(db.ID(taskIDStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", repo.ID.String()).
			Str("taskID", taskIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get clone task")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get clone task",
		})
		return
	}
	if task == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Clone task not found",
		})
		return
	}
	if task.RepositoryID != repo.ID {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Clone task not found",
		})
		return
	}

	out := map[string]interface{}{
		"id":           task.ID,
		"repositoryId": task.RepositoryID,
		"runnerId":     task.RunnerID,
		"status":       task.Status,
		"runnerTaskId": task.RunnerTaskID,
		"createdAt":    task.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    task.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if task.Result != nil {
		out["result"] = *task.Result
	}

	util.WriteJSON(w, http.StatusOK, out)
}
