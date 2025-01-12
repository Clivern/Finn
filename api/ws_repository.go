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
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceRepositoryRequest is the request body for creating a workspace repository.
type CreateWorkspaceRepositoryRequest struct {
	Name             string `json:"name" validate:"required,max=100" label:"Name"`
	GitURL           string `json:"gitUrl" validate:"required,max=500" label:"Git URL"`
	GitBranch        string `json:"gitBranch" validate:"max=100" label:"Git branch"`
	GitTag           string `json:"gitTag" validate:"max=100" label:"Git tag"`
	SSHKeyID         string `json:"sshKeyId" validate:"omitempty" label:"SSH key ID"`
	GitCredentialsID string `json:"gitCredentialsId" validate:"omitempty" label:"Git credentials ID"`
}

// UpdateWorkspaceRepositoryRequest is the request body for updating a workspace repository.
type UpdateWorkspaceRepositoryRequest struct {
	Name             string `json:"name" validate:"required,max=100" label:"Name"`
	GitURL           string `json:"gitUrl" validate:"required,max=500" label:"Git URL"`
	GitBranch        string `json:"gitBranch" validate:"max=100" label:"Git branch"`
	GitTag           string `json:"gitTag" validate:"max=100" label:"Git tag"`
	SSHKeyID         string `json:"sshKeyId" validate:"omitempty" label:"SSH key ID"`
	GitCredentialsID string `json:"gitCredentialsId" validate:"omitempty" label:"Git credentials ID"`
}

// CreateWorkspaceRepositoryAction creates a repository in a workspace.
func CreateWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace repository endpoint called")

	var req CreateWorkspaceRepositoryRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
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
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	if req.GitTag == "" && req.GitBranch == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Git branch or tag is required",
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
		log.Error().Err(err).Msg("Failed to create workspace repository")
		service.WriteJSON(
			w,
			http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": "Failed to create repository",
			},
		)
		return
	}
	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
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

// ListWorkspaceRepositoriesAction lists repositories in a workspace.
func ListWorkspaceRepositoriesAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace repositories endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
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
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
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
	list, err := repoDB.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workspace repositories")
		service.WriteJSON(
			w,
			http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": "Failed to list repositories",
			},
		)
		return
	}

	total, _ := repoDB.CountByWorkspaceID(workspace.ID)
	items := make([]map[string]interface{}, 0, len(list))
	for _, repo := range list {
		items = append(items, map[string]interface{}{
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
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"repositories": items,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceRepositoryAction returns a single workspace repository.
func GetWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace repository endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid repository ID",
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workspace repository")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get repository",
		})
		return
	}
	if repo == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Repository not found",
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
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
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

// UpdateWorkspaceRepositoryAction updates a workspace repository.
func UpdateWorkspaceRepositoryAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace repository endpoint called")

	var req UpdateWorkspaceRepositoryRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid repository ID",
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || repo == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Repository not found",
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
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	if req.GitTag == "" && req.GitBranch == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Git branch or tag is required",
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
		log.Error().Err(err).Msg("Failed to update workspace repository")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update repository",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
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
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	workspaceIDStr := chi.URLParam(r, "workspaceID")
	if workspaceIDStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid workspace ID",
		})
		return
	}

	wsModule := module.NewWorkspace(db.GetDB())
	workspace, err := wsModule.GetWorkspace(db.ID(workspaceIDStr))
	if err != nil {
		if errors.Is(err, module.ErrWorkspaceNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Workspace not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get workspace")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get workspace",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid repository ID",
		})
		return
	}

	repoDB := db.NewWorkspaceRepoRepository(db.GetDB())
	repo, err := repoDB.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || repo == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Repository not found",
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
		log.Error().Err(err).Msg("Failed to check workspace permission")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to check access",
		})
		return
	}
	if !allowed {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied to this workspace",
		})
		return
	}

	if err := repoDB.Delete(repo.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete workspace repository")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete repository",
		})
		return
	}

	log.Info().Str("repositoryID", idStr).Msg("Repository deleted successfully")
	service.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
