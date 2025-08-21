// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceBackendRequest is the body for creating a backend.
type CreateWorkspaceBackendRequest struct {
	Name     string `json:"name" validate:"required,max=255" label:"Name"`
	Slug     string `json:"slug" validate:"omitempty,max=255" label:"Slug"`
	Username string `json:"username" validate:"required,max=255" label:"Username"`
	Password string `json:"password" validate:"required,max=255" label:"Password"`
}

// UpdateWorkspaceBackendRequest is the body for updating a backend.
type UpdateWorkspaceBackendRequest struct {
	Name     string `json:"name" validate:"required,max=255" label:"Name"`
	Slug     string `json:"slug" validate:"omitempty,max=255" label:"Slug"`
	Username string `json:"username" validate:"required,max=255" label:"Username"`
	Password string `json:"password" validate:"omitempty,max=255" label:"Password"`
}

// LockBackendStateRequest is the body for locking backend state.
type LockBackendStateRequest struct {
	ID        string `json:"ID"`
	Operation string `json:"Operation"`
	Who       string `json:"Who"`
	Version   string `json:"Version"`
	Path      string `json:"Path"`
	Info      string `json:"Info"`
	Created   string `json:"Created"`
}

// lockMu is a mutex for locking the backend state.
var lockMu sync.Mutex

// lockInfoResponse builds the lock JSON for Terraform.
func lockInfoResponse(lock *db.WorkspaceBackendLock) map[string]interface{} {
	id := lock.LockID
	if id == "" {
		id = lock.ID.String()
	}

	return map[string]interface{}{
		"ID":        id,
		"Operation": lock.Operation,
		"Who":       lock.Who,
		"Version":   lock.Version,
		"Path":      lock.Path,
		"Info":      lock.Info,
		"Created":   lock.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// CreateWorkspaceBackendAction creates a backend in the workspace.
func CreateWorkspaceBackendAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace backend endpoint called")

	var req CreateWorkspaceBackendRequest
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
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceBackendPermission(
		user,
		&db.WorkspaceBackend{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceBackend,
	)
	if err != nil {
		log.Error().
			Err(err).
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

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	var slug string
	for {
		slug = util.GenerateRandomSlug()
		existing, err := repo.GetBySlug(slug)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to check backend slug")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedCreateBackend"),
			})
			return
		}
		if existing == nil {
			break
		}
	}

	backend := &db.WorkspaceBackend{
		WorkspaceID: workspace.ID,
		Name:        req.Name,
		Slug:        slug,
		Username:    req.Username,
		Password:    req.Password,
	}
	if err := repo.Create(backend); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create workspace backend")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateBackend"),
		})
		return
	}

	lockRepo := db.NewWorkspaceBackendLockRepository(db.GetDB())
	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        backend.ID,
		"name":      backend.Name,
		"slug":      backend.Slug,
		"username":  backend.Username,
		"locked":    lockRepo.IsLocked(backend.ID),
		"createdAt": backend.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": backend.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceBackendsAction lists backends in the workspace.
func ListWorkspaceBackendsAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace backends endpoint called")

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
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceBackendPermission(
		user,
		&db.WorkspaceBackend{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceBackends,
	)
	if err != nil {
		log.Error().
			Err(err).
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

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	list, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list workspace backends")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListBackends"),
		})
		return
	}

	lockRepo := db.NewWorkspaceBackendLockRepository(db.GetDB())
	total, _ := repo.CountByWorkspaceID(workspace.ID)
	items := make([]map[string]interface{}, 0, len(list))
	for _, b := range list {
		items = append(items, map[string]interface{}{
			"id":        b.ID,
			"name":      b.Name,
			"slug":      b.Slug,
			"username":  b.Username,
			"locked":    lockRepo.IsLocked(b.ID),
			"createdAt": b.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt": b.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"backends": items,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceBackendAction returns one backend by ID.
func GetWorkspaceBackendAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace backend endpoint called")

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
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidBackendID"),
		})
		return
	}

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	backend, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get workspace backend")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetBackend"),
		})
		return
	}
	if backend == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "backendNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceBackendPermission(user, backend, module.CanGetWorkspaceBackend)
	if err != nil {
		log.Error().
			Err(err).
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

	lockRepo := db.NewWorkspaceBackendLockRepository(db.GetDB())
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":        backend.ID,
		"name":      backend.Name,
		"slug":      backend.Slug,
		"username":  backend.Username,
		"locked":    lockRepo.IsLocked(backend.ID),
		"createdAt": backend.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": backend.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// RevealWorkspaceBackendPasswordAction returns the backend password (use with care).
func RevealWorkspaceBackendPasswordAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Reveal workspace backend password endpoint called")

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
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidBackendID"),
		})
		return
	}

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	backend, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get workspace backend")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetBackend"),
		})
		return
	}
	if backend == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "backendNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceBackendPermission(user, backend, module.CanGetWorkspaceBackend)
	if err != nil {
		log.Error().
			Err(err).
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

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"password": backend.Password,
	})
}

// UpdateWorkspaceBackendAction updates a workspace backend.
func UpdateWorkspaceBackendAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace backend endpoint called")

	var req UpdateWorkspaceBackendRequest
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
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidBackendID"),
		})
		return
	}

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	backend, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || backend == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "backendNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceBackendPermission(user, backend, module.CanUpdateWorkspaceBackend)
	if err != nil {
		log.Error().
			Err(err).
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

	backend.Name = req.Name
	backend.Username = req.Username
	if req.Password != "" {
		backend.Password = req.Password
	}

	if err := repo.Update(backend); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to update workspace backend")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateBackend"),
		})
		return
	}

	lockRepo := db.NewWorkspaceBackendLockRepository(db.GetDB())
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":        backend.ID,
		"name":      backend.Name,
		"slug":      backend.Slug,
		"username":  backend.Username,
		"locked":    lockRepo.IsLocked(backend.ID),
		"createdAt": backend.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": backend.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceBackendAction deletes a workspace backend.
func DeleteWorkspaceBackendAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace backend endpoint called")

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
			Msg("Failed to get workspace")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetWorkspace"),
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidBackendID"),
		})
		return
	}

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	backend, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || backend == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "backendNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceBackendPermission(user, backend, module.CanDeleteWorkspaceBackend)
	if err != nil {
		log.Error().
			Err(err).
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

	if err := repo.Delete(backend.ID); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete workspace backend")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteBackend"),
		})
		return
	}

	log.Info().
		Str("backendID", idStr).
		Msg("Backend deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}

// getBackendBySlugAndBasicAuth resolves backend by slug and validates Basic Auth
func getBackendBySlugAndBasicAuth(w http.ResponseWriter, r *http.Request) *db.WorkspaceBackend {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidBackendID"),
		})
		return nil
	}

	repo := db.NewWorkspaceBackendRepository(db.GetDB())
	backend, err := repo.GetBySlug(slug)
	if err != nil {
		log.Error().
			Err(err).
			Str("slug", slug).
			Msg("Failed to get backend by slug")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetBackend"),
		})
		return nil
	}
	if backend == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "backendNotFound"),
		})
		return nil
	}

	user, pass, ok := r.BasicAuth()
	userMatch := ok && subtle.ConstantTimeCompare([]byte(user), []byte(backend.Username)) == 1
	passMatch := ok && subtle.ConstantTimeCompare([]byte(pass), []byte(backend.Password)) == 1
	if !userMatch || !passMatch {
		w.Header().Set("WWW-Authenticate", `Basic realm="Backend"`)
		util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": locale.TR(r, "notAuthenticated"),
		})
		return nil
	}
	return backend
}

// CreateBackendStateAction saves state for a backend.
func CreateBackendStateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create backend state endpoint called")

	backend := getBackendBySlugAndBasicAuth(w, r)
	if backend == nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedToReadBody"),
		})
		return
	}
	checksum := r.Header.Get("Content-MD5")
	if checksum == "" && len(body) > 0 {
		hash := md5.Sum(body)
		checksum = base64.StdEncoding.EncodeToString(hash[:])
	}

	repo := db.NewWorkspaceBackendStateRepository(db.GetDB())
	state := &db.WorkspaceBackendState{
		BackendID: backend.ID,
		Name:      "_tf_state_",
		Value:     string(body),
		Checksum:  checksum,
	}
	if err := repo.Create(state); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create backend state")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateBackend"),
		})
		return
	}

	if state.Checksum != "" {
		w.Header().Set("Content-MD5", state.Checksum)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(state.Value)); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to write state response")
	}
}

// GetBackendStateAction returns the latest state for a backend.
func GetBackendStateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get backend state endpoint called")

	backend := getBackendBySlugAndBasicAuth(w, r)
	if backend == nil {
		return
	}

	repo := db.NewWorkspaceBackendStateRepository(db.GetDB())
	state, err := repo.GetLatestByBackendID(backend.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get backend state")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetBackend"),
		})
		return
	}

	if state == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if state.Checksum != "" {
		w.Header().Set("Content-MD5", state.Checksum)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(state.Value))
}

// LockBackendStateAction creates a lock for a backend.
func LockBackendStateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Lock backend state endpoint called")

	backend := getBackendBySlugAndBasicAuth(w, r)
	if backend == nil {
		return
	}
	var req LockBackendStateRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedProcessRequest"),
		})
		return
	}

	lockMu.Lock()
	defer lockMu.Unlock()

	repo := db.NewWorkspaceBackendLockRepository(db.GetDB())
	active, err := repo.GetAnyActiveByBackendID(backend.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get active lock")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetBackend"),
		})
		return
	}
	if active != nil {
		util.WriteJSON(w, http.StatusLocked, lockInfoResponse(active))
		return
	}

	lock := &db.WorkspaceBackendLock{
		BackendID: backend.ID,
		LockID:    req.ID,
		Operation: req.Operation,
		Info:      req.Info,
		Who:       req.Who,
		Version:   req.Version,
		Path:      req.Path,
		IsActive:  true,
	}
	if err := repo.Create(lock); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create backend lock")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateBackend"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, lockInfoResponse(lock))
}

// UnlockBackendStateAction releases the active lock for a backend.
func UnlockBackendStateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Unlock backend state endpoint called")

	backend := getBackendBySlugAndBasicAuth(w, r)
	if backend == nil {
		return
	}

	repo := db.NewWorkspaceBackendLockRepository(db.GetDB())
	lock, err := repo.GetAnyActiveByBackendID(backend.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get lock")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetBackend"),
		})
		return
	}
	if lock == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "backendNotFound"),
		})
		return
	}
	if err := repo.SetInactive(lock.ID); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to unlock backend")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateBackend"),
		})
		return
	}
	w.WriteHeader(http.StatusOK)
}
