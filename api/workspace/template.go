// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
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

type Script struct {
	Engine        string `json:"engine" validate:"omitempty,oneof=bash sh python3 node" label:"Script engine"`
	EngineVersion string `json:"engineVersion" validate:"omitempty,max=100" label:"Script engine version"`
	ScriptPath    string `json:"scriptPath" validate:"omitempty,max=10000" label:"Script path"`
}

type IaC struct {
	Engine        string `json:"engine" validate:"omitempty,oneof=terraform opentofu" label:"IaC engine"`
	EngineVersion string `json:"engineVersion" validate:"omitempty,max=100" label:"IaC engine version"`
	AutoApprove   bool   `json:"autoApprove"`
	BackendID     string `json:"backendID" validate:"omitempty,max=64" label:"IaC backend"`
	LogLevel      string `json:"logLevel" validate:"omitempty,oneof=TRACE DEBUG INFO WARN ERROR" label:"IaC log level"`
}

type Ansible struct {
	EngineVersion string `json:"engineVersion" validate:"omitempty,max=100" label:"Ansible engine version"`
	InventoryID   string `json:"inventoryID" validate:"omitempty,max=64" label:"Ansible inventory"`
	PlaybookPath  string `json:"playbookPath" validate:"omitempty,max=4096" label:"Ansible playbook path"`
	VaultPath     string `json:"vaultPath" validate:"omitempty,max=4096" label:"Ansible vault path"`
	VaultPassword string `json:"vaultPassword" validate:"omitempty,max=4096" label:"Ansible vault password"`
	LogLevel      string `json:"logLevel" validate:"omitempty,oneof=0 1 2 3 4 5" label:"Ansible log level"`
}

// CreateWorkspaceTemplateRequest is the body for creating a workspace template.
type CreateWorkspaceTemplateRequest struct {
	Name                string            `json:"name" validate:"required,max=100" label:"Name"`
	Description         string            `json:"description" validate:"required,max=10000" label:"Description"`
	Engine              string            `json:"engine" validate:"required,oneof=script iac ansible" label:"Engine"`
	RepositoryID        string            `json:"repositoryId" validate:"required,max=64" label:"Repository"`
	WorkingDirectory    string            `json:"workingDirectory" validate:"omitempty,max=500" label:"Working directory"`
	BranchOverride      string            `json:"branchOverride" validate:"omitempty,max=100" label:"Branch override"`
	TagOverride         string            `json:"tagOverride" validate:"omitempty,max=100" label:"Tag override"`
	RunnerTags          map[string]string `json:"runnerTags"`
	AllowConcurrentRuns bool              `json:"allowConcurrentRuns"`
	MaxConcurrentRuns   int               `json:"maxConcurrentRuns"`
	ApprovalRequired    bool              `json:"approvalRequired"`
	Script              Script            `json:"script"`
	IaC                 IaC               `json:"iac"`
	Ansible             Ansible           `json:"ansible"`
	ExtraArgs           string            `json:"extraArgs" validate:"omitempty,max=10000" label:"Extra arguments"`
	GroupVarsID         string            `json:"groupVarsId" validate:"omitempty,max=64" label:"Variable group"`
}

// UpdateWorkspaceTemplateRequest is the body for updating a workspace template.
type UpdateWorkspaceTemplateRequest struct {
	Name                string            `json:"name" validate:"required,max=100" label:"Name"`
	Description         string            `json:"description" validate:"required,max=10000" label:"Description"`
	Engine              string            `json:"engine" validate:"required,oneof=script iac ansible" label:"Engine"`
	RepositoryID        string            `json:"repositoryId" validate:"required,max=64" label:"Repository"`
	WorkingDirectory    string            `json:"workingDirectory" validate:"omitempty,max=500" label:"Working directory"`
	BranchOverride      string            `json:"branchOverride" validate:"omitempty,max=100" label:"Branch override"`
	TagOverride         string            `json:"tagOverride" validate:"omitempty,max=100" label:"Tag override"`
	RunnerTags          map[string]string `json:"runnerTags"`
	AllowConcurrentRuns bool              `json:"allowConcurrentRuns"`
	MaxConcurrentRuns   int               `json:"maxConcurrentRuns"`
	ApprovalRequired    bool              `json:"approvalRequired"`
	Script              Script            `json:"script"`
	IaC                 IaC               `json:"iac"`
	Ansible             Ansible           `json:"ansible"`
	ExtraArgs           string            `json:"extraArgs" validate:"omitempty,max=10000" label:"Extra arguments"`
	GroupVarsID         string            `json:"groupVarsId" validate:"omitempty,max=64" label:"Variable group"`
}

// CreateWorkspaceTemplateAction creates a workspace template.
func CreateWorkspaceTemplateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create workspace template endpoint called")

	var req CreateWorkspaceTemplateRequest
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
	allowed, err := permModule.CheckWorkspaceTemplatePermission(
		user,
		&db.WorkspaceTemplate{WorkspaceID: workspace.ID},
		module.CanAddWorkspaceTemplate,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace template permission")
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

	repoRepo := db.NewWorkspaceRepoRepository(db.GetDB())
	wsRepo, err := repoRepo.GetByIDAndWorkspace(db.ID(req.RepositoryID), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", req.RepositoryID).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace repository for template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetRepository"),
		})
		return
	}
	if wsRepo == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "repositoryNotFound"),
		})
		return
	}

	if req.GroupVarsID != "" {
		gvRepo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
		gv, err := gvRepo.GetByIDAndWorkspace(db.ID(req.GroupVarsID), workspace.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceID", workspaceIDStr).
				Str("groupVarsID", req.GroupVarsID).
				Str("userID", user.ID.String()).
				Msg("Failed to get variable group for template")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedGetVariableGroup"),
			})
			return
		}
		if gv == nil {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "variableGroupNotFound"),
			})
			return
		}
	}

	scriptJSON, iacJSON, ansibleJSON, err := marshalTemplateConfigs(req.Script, req.IaC, req.Ansible)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encode workspace template config JSON")
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedProcessRequest"),
		})
		return
	}

	tmpl := &db.WorkspaceTemplate{
		WorkspaceID:         workspace.ID,
		RepositoryID:        db.ID(req.RepositoryID),
		Name:                req.Name,
		Description:         req.Description,
		Engine:              req.Engine,
		WorkingDirectory:    req.WorkingDirectory,
		GitBranchOverride:   req.BranchOverride,
		GitTagOverride:      req.TagOverride,
		RunnerTags:          templateRunnerTagsToJSON(templateNormalizeRunnerTags(req.RunnerTags)),
		AllowConcurrentRuns: req.AllowConcurrentRuns,
		MaxConcurrentRuns:   req.MaxConcurrentRuns,
		ApprovalRequired:    req.ApprovalRequired,
		ScriptConfig:        scriptJSON,
		IaCConfig:           iacJSON,
		AnsibleConfig:       ansibleJSON,
		ExtraArgs:           req.ExtraArgs,
	}
	if req.GroupVarsID != "" {
		gid := db.ID(req.GroupVarsID)
		tmpl.GroupVarsID = &gid
	}

	tmplRepo := db.NewWorkspaceTemplateRepository(db.GetDB())
	if err := tmplRepo.Create(tmpl); err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to create workspace template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCreateTemplate"),
		})
		return
	}

	log.Info().
		Str("templateID", tmpl.ID.String()).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Workspace template created successfully")

	resp, err := workspaceTemplateToResponse(tmpl)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", tmpl.ID.String()).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to build workspace template response")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedProcessRequest"),
		})
		return
	}
	util.WriteJSON(w, http.StatusCreated, resp)
}

// ListWorkspaceTemplatesAction lists workspace templates.
func ListWorkspaceTemplatesAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("List workspace templates endpoint called")

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
	allowed, err := permModule.CheckWorkspaceTemplatePermission(
		user,
		&db.WorkspaceTemplate{WorkspaceID: workspace.ID},
		module.CanGetWorkspaceTemplates,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace template permission")
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

	tmplRepo := db.NewWorkspaceTemplateRepository(db.GetDB())
	rows, err := tmplRepo.ListByWorkspaceID(workspace.ID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to list workspace templates")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedListTemplates"),
		})
		return
	}

	total, _ := tmplRepo.CountByWorkspaceID(workspace.ID)
	templates := make([]map[string]interface{}, 0, len(rows))
	for _, t := range rows {
		item, err := workspaceTemplateToResponse(t)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceID", workspaceIDStr).
				Str("userID", user.ID.String()).
				Msg("Failed to build workspace template response")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedProcessRequest"),
			})
			return
		}
		templates = append(templates, item)
	}

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetWorkspaceTemplateAction returns one workspace template.
func GetWorkspaceTemplateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Get workspace template endpoint called")

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
			"errorMessage": locale.TR(r, "invalidTemplateID"),
		})
		return
	}

	tmplRepo := db.NewWorkspaceTemplateRepository(db.GetDB())
	tmpl, err := tmplRepo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetTemplate"),
		})
		return
	}
	if tmpl == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "templateNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceTemplatePermission(
		user,
		tmpl,
		module.CanGetWorkspaceTemplate,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace template permission")
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

	resp, err := workspaceTemplateToResponse(tmpl)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to build workspace template response")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedProcessRequest"),
		})
		return
	}
	util.WriteJSON(w, http.StatusOK, resp)
}

// UpdateWorkspaceTemplateAction updates a workspace template.
func UpdateWorkspaceTemplateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Update workspace template endpoint called")

	var req UpdateWorkspaceTemplateRequest
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
			"errorMessage": locale.TR(r, "invalidTemplateID"),
		})
		return
	}

	tmplRepo := db.NewWorkspaceTemplateRepository(db.GetDB())
	tmpl, err := tmplRepo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetTemplate"),
		})
		return
	}
	if tmpl == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "templateNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceTemplatePermission(
		user,
		tmpl,
		module.CanUpdateWorkspaceTemplate,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace template permission")
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

	repoRepo := db.NewWorkspaceRepoRepository(db.GetDB())
	wsRepo, err := repoRepo.GetByIDAndWorkspace(db.ID(req.RepositoryID), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceID", workspaceIDStr).
			Str("repositoryID", req.RepositoryID).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace repository for template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetRepository"),
		})
		return
	}
	if wsRepo == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "repositoryNotFound"),
		})
		return
	}

	if req.GroupVarsID != "" {
		gvRepo := db.NewWorkspaceVariableGroupRepository(db.GetDB())
		gv, err := gvRepo.GetByIDAndWorkspace(db.ID(req.GroupVarsID), workspace.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceID", workspaceIDStr).
				Str("groupVarsID", req.GroupVarsID).
				Str("userID", user.ID.String()).
				Msg("Failed to get variable group for template")
			util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedGetVariableGroup"),
			})
			return
		}
		if gv == nil {
			util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": locale.TR(r, "variableGroupNotFound"),
			})
			return
		}
	}

	scriptJSON, iacJSON, ansibleJSON, err := marshalTemplateConfigs(req.Script, req.IaC, req.Ansible)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to encode workspace template config JSON")
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedProcessRequest"),
		})
		return
	}

	tmpl.RepositoryID = db.ID(req.RepositoryID)
	tmpl.Name = req.Name
	tmpl.Description = req.Description
	tmpl.Engine = req.Engine
	tmpl.WorkingDirectory = req.WorkingDirectory
	tmpl.GitBranchOverride = req.BranchOverride
	tmpl.GitTagOverride = req.TagOverride
	tmpl.RunnerTags = templateRunnerTagsToJSON(templateNormalizeRunnerTags(req.RunnerTags))
	tmpl.AllowConcurrentRuns = req.AllowConcurrentRuns
	tmpl.MaxConcurrentRuns = req.MaxConcurrentRuns
	tmpl.ApprovalRequired = req.ApprovalRequired
	tmpl.ScriptConfig = scriptJSON
	tmpl.IaCConfig = iacJSON
	tmpl.AnsibleConfig = ansibleJSON
	tmpl.ExtraArgs = req.ExtraArgs
	tmpl.GroupVarsID = nil
	if req.GroupVarsID != "" {
		gid := db.ID(req.GroupVarsID)
		tmpl.GroupVarsID = &gid
	}

	if err := tmplRepo.Update(tmpl); err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to update workspace template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateTemplate"),
		})
		return
	}

	tmpl, err = tmplRepo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil || tmpl == nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to reload workspace template after update")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetTemplate"),
		})
		return
	}

	resp, err := workspaceTemplateToResponse(tmpl)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to build workspace template response")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedProcessRequest"),
		})
		return
	}
	util.WriteJSON(w, http.StatusOK, resp)
}

// DeleteWorkspaceTemplateAction deletes a workspace template.
func DeleteWorkspaceTemplateAction(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Delete workspace template endpoint called")

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
			"errorMessage": locale.TR(r, "invalidTemplateID"),
		})
		return
	}

	tmplRepo := db.NewWorkspaceTemplateRepository(db.GetDB())
	tmpl, err := tmplRepo.GetByIDAndWorkspace(db.ID(idStr), workspace.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to get workspace template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetTemplate"),
		})
		return
	}
	if tmpl == nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": locale.TR(r, "templateNotFound"),
		})
		return
	}

	permModule := module.NewPermission(db.GetDB())
	allowed, err := permModule.CheckWorkspaceTemplatePermission(
		user,
		tmpl,
		module.CanDeleteWorkspaceTemplate,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to check workspace template permission")
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

	if err := tmplRepo.Delete(tmpl.ID); err != nil {
		log.Error().
			Err(err).
			Str("templateID", idStr).
			Str("workspaceID", workspaceIDStr).
			Str("userID", user.ID.String()).
			Msg("Failed to delete workspace template")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedDeleteTemplate"),
		})
		return
	}

	log.Info().
		Str("templateID", idStr).
		Str("workspaceID", workspaceIDStr).
		Str("userID", user.ID.String()).
		Msg("Workspace template deleted successfully")
	util.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}

func marshalTemplateConfigs(script Script, iac IaC, ansible Ansible) (string, string, string, error) {
	sb, err := json.Marshal(script)
	if err != nil {
		return "", "", "", err
	}

	ib, err := json.Marshal(iac)
	if err != nil {
		return "", "", "", err
	}

	ab, err := json.Marshal(ansible)
	if err != nil {
		return "", "", "", err
	}

	return string(sb), string(ib), string(ab), nil
}

func workspaceTemplateToResponse(t *db.WorkspaceTemplate) (map[string]interface{}, error) {
	var script Script
	var iac IaC
	var ansible Ansible

	if t.ScriptConfig != "" {
		if err := json.Unmarshal([]byte(t.ScriptConfig), &script); err != nil {
			return nil, err
		}
	}
	if t.IaCConfig != "" {
		if err := json.Unmarshal([]byte(t.IaCConfig), &iac); err != nil {
			return nil, err
		}
	}
	if t.AnsibleConfig != "" {
		if err := json.Unmarshal([]byte(t.AnsibleConfig), &ansible); err != nil {
			return nil, err
		}
	}

	var groupVarsID string
	if t.GroupVarsID != nil {
		groupVarsID = t.GroupVarsID.String()
	}

	return map[string]interface{}{
		"id":                  t.ID,
		"name":                t.Name,
		"description":         t.Description,
		"engine":              t.Engine,
		"workspaceId":         t.WorkspaceID,
		"repositoryId":        t.RepositoryID,
		"workingDirectory":    t.WorkingDirectory,
		"branchOverride":      t.GitBranchOverride,
		"tagOverride":         t.GitTagOverride,
		"runnerTags":          templateRunnerTagsFromStr(t.RunnerTags),
		"allowConcurrentRuns": t.AllowConcurrentRuns,
		"maxConcurrentRuns":   t.MaxConcurrentRuns,
		"approvalRequired":    t.ApprovalRequired,
		"script":              script,
		"iac":                 iac,
		"ansible":             ansible,
		"extraArgs":           t.ExtraArgs,
		"groupVarsId":         groupVarsID,
		"createdAt":           t.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":           t.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func templateNormalizeRunnerTags(in map[string]string) map[string]string {
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

func templateRunnerTagsToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func templateRunnerTagsFromStr(s string) map[string]string {
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
