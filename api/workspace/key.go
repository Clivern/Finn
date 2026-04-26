// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateWorkspaceSSHKeyRequest is the body for creating an SSH key.
type CreateWorkspaceSSHKeyRequest struct {
	Name           string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName  string `json:"committerName" validate:"required,max=100" label:"Committer name"`
	CommitterEmail string `json:"committerEmail" validate:"required,max=100" label:"Committer email"`
	PrivateKey     string `json:"privateKey" validate:"required" label:"Private key"`
	PublicKey      string `json:"publicKey" validate:"omitempty" label:"Public key"`
	Passphrase     string `json:"passphrase" validate:"omitempty,max=200" label:"Passphrase"`
}

// UpdateWorkspaceSSHKeyRequest is the body for updating an SSH key.
type UpdateWorkspaceSSHKeyRequest struct {
	Name           string `json:"name" validate:"required,max=100" label:"Name"`
	CommitterName  string `json:"committerName" validate:"required,max=100" label:"Committer name"`
	CommitterEmail string `json:"committerEmail" validate:"required,max=100" label:"Committer email"`
	PrivateKey     string `json:"privateKey" validate:"omitempty" label:"Private key"`
	PublicKey      string `json:"publicKey" validate:"omitempty" label:"Public key"`
	Passphrase     string `json:"passphrase" validate:"omitempty,max=200" label:"Passphrase"`
}

// CreateWorkspaceSSHKeyAction creates an SSH key in the workspace.
func CreateWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace SSH key endpoint called")

	var req CreateWorkspaceSSHKeyRequest
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
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		&db.WorkspaceSSHKey{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceSSHKey,
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

	privateKeyEnc, err := util.EncryptToBase64String(
		req.PrivateKey,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt SSH private key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateSshKey"),
		})
		return
	}
	passphraseEnc, err := util.EncryptToBase64String(
		req.Passphrase,
		util.GetEncryptionKey(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encrypt SSH key passphrase")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateSshKey"),
		})
		return
	}

	key := &db.WorkspaceSSHKey{
		WorkspaceID:    workspace.ID,
		Name:           req.Name,
		CommitterName:  req.CommitterName,
		CommitterEmail: req.CommitterEmail,
		PrivateKey:     privateKeyEnc,
		Passphrase:     passphraseEnc,
	}
	if req.PublicKey != "" {
		key.PublicKey = &req.PublicKey
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	if err := repo.Create(key); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create workspace SSH key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateSshKey"),
		})
		return
	}

	util.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             key.ID,
		"workspaceId":    key.WorkspaceID,
		"name":           key.Name,
		"committerName":  key.CommitterName,
		"committerEmail": key.CommitterEmail,
		"publicKey":      key.PublicKey,
		"createdAt":      key.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      key.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// ListWorkspaceSSHKeysAction lists SSH keys in the workspace.
func ListWorkspaceSSHKeysAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace SSH keys endpoint called")

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
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		&db.WorkspaceSSHKey{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceSSHKeys,
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

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	result, err := repo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to list workspace SSH keys")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListSshKeys"),
		})
		return
	}

	total, _ := repo.CountByWorkspaceID(workspace.ID)
	keys := make([]map[string]interface{}, 0, len(result))
	for _, key := range result {
		keys = append(keys, map[string]interface{}{
			"id":             key.ID,
			"workspaceId":    key.WorkspaceID,
			"name":           key.Name,
			"committerName":  key.CommitterName,
			"committerEmail": key.CommitterEmail,
			"createdAt":      key.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":      key.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"keys": keys,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceSSHKeyAction returns one SSH key.
func GetWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace SSH key endpoint called")

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
			"errorMessage": locale.TR(r, "invalidSshKeyID"),
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	key, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("sshKeyID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace SSH key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetSshKey"),
		})
		return
	}
	if key == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "sshKeyNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		key,
		module.CanGetWorkspaceSSHKey,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("sshKeyID", idStr).
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
		"id":             key.ID,
		"workspaceId":    key.WorkspaceID,
		"name":           key.Name,
		"committerName":  key.CommitterName,
		"committerEmail": key.CommitterEmail,
		"publicKey":      key.PublicKey,
		"createdAt":      key.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      key.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateWorkspaceSSHKeyAction updates a workspace SSH key.
func UpdateWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace SSH key endpoint called")

	var req UpdateWorkspaceSSHKeyRequest
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
			"errorMessage": locale.TR(r, "invalidSshKeyID"),
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	key, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || key == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "sshKeyNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		key,
		module.CanUpdateWorkspaceSSHKey,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("sshKeyID", idStr).
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

	key.Name = req.Name
	key.CommitterName = req.CommitterName
	key.CommitterEmail = req.CommitterEmail

	if req.PrivateKey != "" {
		privateKeyEnc, err := util.EncryptToBase64String(
			req.PrivateKey,
			util.GetEncryptionKey(),
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("sshKeyID", idStr).
				Str("workspaceID", workspaceIDStr).
				Str("userID", user.ID.String()).
				Msg("Failed to encrypt SSH private key")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedUpdateSshKey"),
			})
			return
		}
		key.PrivateKey = privateKeyEnc
	}
	if req.PublicKey != "" {
		key.PublicKey = &req.PublicKey
	}

	if strings.TrimSpace(req.Passphrase) != "" {
		passphraseEnc, err := util.EncryptToBase64String(
			req.Passphrase,
			util.GetEncryptionKey(),
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("sshKeyID", idStr).
				Str("workspaceID", workspaceIDStr).
				Str("userID", user.ID.String()).
				Msg("Failed to encrypt SSH key passphrase")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedUpdateSshKey"),
			})
			return
		}
		key.Passphrase = passphraseEnc
	}

	if err := repo.Update(key); err != nil {
		log.Error().
			Err(err).
			Str("sshKeyID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to update workspace SSH key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateSshKey"),
		})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":             key.ID,
		"workspaceId":    key.WorkspaceID,
		"name":           key.Name,
		"committerName":  key.CommitterName,
		"committerEmail": key.CommitterEmail,
		"publicKey":      key.PublicKey,
		"createdAt":      key.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      key.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteWorkspaceSSHKeyAction deletes a workspace SSH key.
func DeleteWorkspaceSSHKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace SSH key endpoint called")

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
			"errorMessage": locale.TR(r, "invalidSshKeyID"),
		})
		return
	}

	repo := db.NewWorkspaceSSHKeyRepository(db.GetDB())
	key, err := repo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || key == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "sshKeyNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceSSHKeyPermission(
		user,
		key,
		module.CanDeleteWorkspaceSSHKey,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("sshKeyID", idStr).
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

	if err := repo.Delete(key.ID); err != nil {
		log.Error().
			Err(err).
			Str("sshKeyID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to delete workspace SSH key")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteSshKey"),
		})
		return
	}

	log.Info().
		Str("sshKeyID", idStr).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("SSH key deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
