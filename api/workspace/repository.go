// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/mgmt"
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceRepositoryRequest is the body for creating a repo.
type CreateWorkspaceRepositoryRequest struct {
	Name             string `json:"name" validate:"required,max=100" label:"Name"`
	GitURL           string `json:"gitUrl" validate:"required,max=500" label:"Git URL"`
	GitBranch        string `json:"gitBranch" validate:"max=100" label:"Git branch"`
	GitTag           string `json:"gitTag" validate:"max=100" label:"Git tag"`
	SSHKeyID         string `json:"sshKeyId" validate:"omitempty" label:"SSH key ID"`
	GitCredentialsID string `json:"gitCredentialsId" validate:"omitempty" label:"Git credentials ID"`
}

// UpdateWorkspaceRepositoryRequest is the body for updating a repo.
type UpdateWorkspaceRepositoryRequest struct {
	Name             string `json:"name" validate:"required,max=100" label:"Name"`
	GitURL           string `json:"gitUrl" validate:"required,max=500" label:"Git URL"`
	GitBranch        string `json:"gitBranch" validate:"max=100" label:"Git branch"`
	GitTag           string `json:"gitTag" validate:"max=100" label:"Git tag"`
	SSHKeyID         string `json:"sshKeyId" validate:"omitempty" label:"SSH key ID"`
	GitCredentialsID string `json:"gitCredentialsId" validate:"omitempty" label:"Git credentials ID"`
}

// CreateWorkspaceRepositoryAction creates a repository in the workspace.
func CreateWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace repository endpoint called")

	var req CreateWorkspaceRepositoryRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

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
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceRepositoryPermission(
		user,
		&db.WorkspaceRepo{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceRepository,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
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

	if req.GitTag == "" && req.GitBranch == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "gitBranchOrTagRequired"),
		})
		return
	}

	repo := &db.WorkspaceRepo{
		WorkspaceID: workspace.ID,
		Name:        req.Name,
		GitURL:      req.GitURL,
		GitBranch:   req.GitBranch,
		GitTag:      req.GitTag,
	}
	if req.SSHKeyID != "" {
		sid := db.ID(req.SSHKeyID)
		repo.SSHKeyID = &sid
	}
	if req.GitCredentialsID != "" {
		gid := db.ID(req.GitCredentialsID)
		repo.GitCredentialsID = &gid
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	if err := repoDB.Create(repo); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create workspace repository")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateRepository"),
		})
		return
	}

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":               repo.ID,
		"workspaceId":      repo.WorkspaceID,
		"name":             repo.Name,
		"gitUrl":           repo.GitURL,
		"gitBranch":        repo.GitBranch,
		"gitTag":           repo.GitTag,
		"sshKeyId":         repo.SSHKeyID,
		"latestStatus":     "unknown",
		"gitCredentialsId": repo.GitCredentialsID,
		"createdAt":        repo.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        repo.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceRepositoriesAction lists repositories in the workspace.
func ListWorkspaceRepositoriesAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace repositories endpoint called")

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
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceRepositoryPermission(
		user,
		&db.WorkspaceRepo{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceRepositories,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
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

	limit, offset := 50, 0
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

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	result, err := repoDB.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to list workspace repositories")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListRepositories"),
		})
		return
	}

	total, _ := repoDB.CountByWorkspaceID(workspace.ID)
	cloneTaskRepo := db.NewWorkspaceRepositoryCloneTaskRepository(db.GetDB())
	repos := make([]map[string]interface{}, 0, len(result))

	for _, repo := range result {
		latestStatus, _ := cloneTaskRepo.LatestStatusByRepositoryID(repo.ID)
		repos = append(repos, map[string]interface{}{
			"id":               repo.ID,
			"workspaceId":      repo.WorkspaceID,
			"name":             repo.Name,
			"gitUrl":           repo.GitURL,
			"gitBranch":        repo.GitBranch,
			"gitTag":           repo.GitTag,
			"sshKeyId":         repo.SSHKeyID,
			"latestStatus":     latestStatus,
			"gitCredentialsId": repo.GitCredentialsID,
			"createdAt":        repo.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":        repo.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"repositories": repos,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceRepositoryAction returns one repository by ID.
func GetWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace repository endpoint called")

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
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRepositoryID"),
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	cloneTaskRepo := db.NewWorkspaceRepositoryCloneTaskRepository(db.GetDB())

	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("repositoryID", idStr).
			Str("workspaceID", workspaceIDStr).
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
		module.CanGetWorkspaceRepository,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("repositoryID", idStr).
			Str("workspaceID", workspaceIDStr).
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
	latestStatus, _ := cloneTaskRepo.LatestStatusByRepositoryID(repo.ID)

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":               repo.ID,
		"workspaceId":      repo.WorkspaceID,
		"name":             repo.Name,
		"gitUrl":           repo.GitURL,
		"gitBranch":        repo.GitBranch,
		"gitTag":           repo.GitTag,
		"sshKeyId":         repo.SSHKeyID,
		"latestStatus":     latestStatus,
		"gitCredentialsId": repo.GitCredentialsID,
		"createdAt":        repo.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        repo.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceRepositoryAction updates a workspace repository.
func UpdateWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace repository endpoint called")

	var req UpdateWorkspaceRepositoryRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

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
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRepositoryID"),
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || repo == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "repositoryNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceRepositoryPermission(
		user,
		repo,
		module.CanUpdateWorkspaceRepository,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("repositoryID", idStr).
			Str("workspaceID", workspaceIDStr).
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

	if req.GitTag == "" && req.GitBranch == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "gitBranchOrTagRequired"),
		})
		return
	}

	repo.Name = req.Name
	repo.GitURL = req.GitURL
	repo.GitBranch = req.GitBranch
	repo.GitTag = req.GitTag
	repo.SSHKeyID = nil
	repo.GitCredentialsID = nil
	if req.SSHKeyID != "" {
		sid := db.ID(req.SSHKeyID)
		repo.SSHKeyID = &sid
	}
	if req.GitCredentialsID != "" {
		gid := db.ID(req.GitCredentialsID)
		repo.GitCredentialsID = &gid
	}
	if err := repoDB.Update(repo); err != nil {
		log.Error().
			Err(err).
			Str("repositoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to update workspace repository")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateRepository"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":               repo.ID,
		"workspaceId":      repo.WorkspaceID,
		"name":             repo.Name,
		"gitUrl":           repo.GitURL,
		"gitBranch":        repo.GitBranch,
		"gitTag":           repo.GitTag,
		"sshKeyId":         repo.SSHKeyID,
		"gitCredentialsId": repo.GitCredentialsID,
		"createdAt":        repo.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        repo.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceRepositoryAction deletes a workspace repository.
func DeleteWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace repository endpoint called")

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
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRepositoryID"),
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || repo == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "repositoryNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceRepositoryPermission(
		user,
		repo,
		module.CanDeleteWorkspaceRepository,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("repositoryID", idStr).
			Str("workspaceID", workspaceIDStr).
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

	if err := repoDB.Delete(repo.ID); err != nil {
		log.Error().
			Err(err).
			Str("repositoryID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to delete workspace repository")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteRepository"),
		})
		return
	}

	log.Info().
		Str("repositoryID", idStr).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Repository deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}

// CloneWorkspaceRepositoryAction triggers a clone on a runner and returns the task.
func CloneWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Clone workspace repository endpoint called")

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

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidRepositoryID"),
		})
		return
	}

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
			Str("repositoryID", idStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", idStr).
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
			Str("repositoryID", idStr).
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

	async := mgmt.NewAsync(
		db.NewRunnerRepository(db.GetDB()),
		repoDB,
		db.NewWorkspaceGitCredRepository(db.GetDB()),
		db.NewWorkspaceSSHKeyRepository(db.GetDB()),
	)
	runner, err := async.GetActiveRunner(r.Context())
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", idStr).
			Str("userID", user.ID.String()).
			Msg("No active runner for clone")
		util.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"errorMessage": "No active runner available",
		})
		return
	}

	runnerTaskID, err := async.Clone(r.Context(), repo, runner)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Str("repositoryID", idStr).
			Msg("Clone request to runner failed")
		util.WriteJSON(w, http.StatusBadGateway, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	cloneTask := &db.WorkspaceRepositoryCloneTask{
		RepositoryID: repo.ID,
		RunnerID:     &runner.ID,
		Status:       "pending",
		RunnerTaskID: runnerTaskID,
	}
	cloneTaskRepo := db.NewWorkspaceRepositoryCloneTaskRepository(db.GetDB())
	if err := cloneTaskRepo.Create(cloneTask); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", idStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create clone task")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create clone task",
		})
		return
	}

	if w := mgmt.DefaultWorker(); w != nil {
		w.Enqueue(mgmt.FollowUpTask{ID: cloneTask.ID, Type: "clone"})
	}

	util.WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"id":           cloneTask.ID,
		"repositoryId": cloneTask.RepositoryID,
		"runnerId":     cloneTask.RunnerID,
		"status":       cloneTask.Status,
		"runnerTaskId": cloneTask.RunnerTaskID,
		"createdAt":    cloneTask.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    cloneTask.UpdatedAt.UTC().Format(time.RFC3339),
	})
}
