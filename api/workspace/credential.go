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
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceCredRequest is the body for creating a workspace credential
type CreateWorkspaceCredRequest struct {
	Type            string `json:"type" validate:"required,oneof=git credential none,max=50" label:"Type"`
	Name            string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName   string `json:"committerName" validate:"omitempty,max=100" label:"Committer name"`
	CommitterEmail  string `json:"committerEmail" validate:"omitempty,max=100" label:"Committer email"`
	Username        string `json:"username" validate:"omitempty,max=100" label:"Username"`
	PasswordOrToken string `json:"passwordOrToken" validate:"omitempty,max=500" label:"Password or token"`
}

// UpdateWorkspaceCredRequest is the body for updating a workspace credential.
type UpdateWorkspaceCredRequest struct {
	Type            string `json:"type" validate:"required,oneof=git credential none,max=50" label:"Type"`
	Name            string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName   string `json:"committerName" validate:"omitempty,max=100" label:"Committer name"`
	CommitterEmail  string `json:"committerEmail" validate:"omitempty,max=100" label:"Committer email"`
	Username        string `json:"username" validate:"omitempty,max=100" label:"Username"`
	PasswordOrToken string `json:"passwordOrToken" validate:"omitempty,max=500" label:"Password or token"`
}

// CreateWorkspaceCredAction creates a credential in the workspace
func CreateWorkspaceCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace credential endpoint called")

	var req CreateWorkspaceCredRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	credType := req.Type
	if credType == "" {
		credType = db.CredentialTypeGit
	}
	switch credType {
	case db.CredentialTypeGit:
		if req.CommitterName == "" || req.CommitterEmail == "" || req.Username == "" || req.PasswordOrToken == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "gitCredentialRequiresCommitterAndUsernamePassword"),
			})
			return
		}
	case db.CredentialTypeCredential:
		if req.Username == "" || req.PasswordOrToken == "" {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "credentialTypeRequiresUsernamePassword"),
			})
			return
		}
	case db.CredentialTypeNone:
		// only name required
	default:
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "invalidCredentialType"),
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
	allowed, err := permModule.CheckWorkspaceCredPermission(
		user,
		&db.WorkspaceCred{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceCred,
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

	token, err := util.EncryptToBase64String(
		req.PasswordOrToken,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt credential password")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateGitCredentials"),
		})
		return
	}

	cred := &db.WorkspaceCred{
		WorkspaceID:     workspace.ID,
		Name:            req.Name,
		Type:            credType,
		CommitterName:   req.CommitterName,
		CommitterEmail:  req.CommitterEmail,
		Username:        req.Username,
		PasswordOrToken: token,
	}

	repo := db.NewWorkspaceCredRepository(db.GetDB())
	if err := repo.Create(cred); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create workspace credential")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateGitCredentials"),
		})
		return
	}

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             cred.ID,
		"workspaceId":    cred.WorkspaceID,
		"name":           cred.Name,
		"type":           cred.Type,
		"committerName":  cred.CommitterName,
		"committerEmail": cred.CommitterEmail,
		"username":       cred.Username,
		"createdAt":      cred.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      cred.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceCredsAction lists credentials in the workspace.
func ListWorkspaceCredsAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace credentials endpoint called")

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
	allowed, err := permModule.CheckWorkspaceCredPermission(
		user,
		&db.WorkspaceCred{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceCreds,
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

	repo := db.NewWorkspaceCredRepository(db.GetDB())
	result, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to list workspace credentials")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListGitCredentials"),
		})
		return
	}

	total, _ := repo.CountByWorkspaceID(workspace.ID)
	credentials := make([]map[string]interface{}, 0, len(result))
	for _, c := range result {
		credentials = append(credentials, map[string]interface{}{
			"id":             c.ID,
			"workspaceId":    c.WorkspaceID,
			"name":           c.Name,
			"type":           c.Type,
			"committerName":  c.CommitterName,
			"committerEmail": c.CommitterEmail,
			"username":       c.Username,
			"createdAt":      c.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":      c.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"credentials": credentials,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceCredAction returns one credential
func GetWorkspaceCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace credential endpoint called")

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
			"errorMessage": locale.TR(r, "invalidGitCredentialsID"),
		})
		return
	}

	repo := db.NewWorkspaceCredRepository(db.GetDB())
	c, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("credID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace credential")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetGitCredentials"),
		})
		return
	}
	if c == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "gitCredentialsNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceCredPermission(
		user,
		c,
		module.CanGetWorkspaceCred,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("credID", idStr).
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

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":             c.ID,
		"workspaceId":    c.WorkspaceID,
		"name":           c.Name,
		"type":           c.Type,
		"committerName":  c.CommitterName,
		"committerEmail": c.CommitterEmail,
		"username":       c.Username,
		"createdAt":      c.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      c.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceCredAction updates a workspace credential.
func UpdateWorkspaceCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace credential endpoint called")

	var req UpdateWorkspaceCredRequest
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
			"errorMessage": locale.TR(r, "invalidGitCredentialsID"),
		})
		return
	}

	repo := db.NewWorkspaceCredRepository(db.GetDB())
	c, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || c == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "gitCredentialsNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceCredPermission(
		user,
		c,
		module.CanUpdateWorkspaceCred,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("credID", idStr).
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

	credType := req.Type
	if credType != "" {
		c.Type = credType
	}
	c.Name = req.Name
	c.CommitterName = req.CommitterName
	c.CommitterEmail = req.CommitterEmail
	c.Username = req.Username
	if req.PasswordOrToken != "" {
		token, err := util.EncryptToBase64String(
			req.PasswordOrToken,
			util.GetEncryptionKey(),
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("credID", idStr).
				Str("workspaceID", workspaceIDStr).
				Str("userID", user.ID.String()).
				Msg("Failed to encrypt credential password")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedUpdateGitCredentials"),
			})
			return
		}
		c.PasswordOrToken = token
	}

	if err := repo.Update(c); err != nil {
		log.Error().
			Err(err).
			Str("credID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to update workspace credential")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateGitCredentials"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":             c.ID,
		"workspaceId":    c.WorkspaceID,
		"name":           c.Name,
		"type":           c.Type,
		"committerName":  c.CommitterName,
		"committerEmail": c.CommitterEmail,
		"username":       c.Username,
		"createdAt":      c.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      c.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceCredAction deletes a workspace credential.
func DeleteWorkspaceCredAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace credential endpoint called")

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
			"errorMessage": locale.TR(r, "invalidGitCredentialsID"),
		})
		return
	}

	repo := db.NewWorkspaceCredRepository(db.GetDB())
	c, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || c == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "gitCredentialsNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceCredPermission(
		user,
		c,
		module.CanDeleteWorkspaceCred,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("credID", idStr).
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

	if err := repo.Delete(c.ID); err != nil {
		log.Error().
			Err(err).
			Str("credID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to delete workspace credential")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteGitCredentials"),
		})
		return
	}

	log.Info().
		Str("credID", idStr).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Credential deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
