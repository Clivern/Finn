// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tool

import (
	"fmt"
	"strings"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/module"
	genkitai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/rs/zerolog/log"
)

const (
	ToolCreateWorkspace = "createWorkspace"
)

// CreateWorkspaceInput is the input for createWorkspace.
type CreateWorkspaceInput struct {
	Name string `json:"name" jsonschema_description:"Workspace name (1–255 chars)"`
}

var createWorkspaceRef genkitai.ToolRef

// RegisterCreateWorkspace registers the createWorkspace tool and returns its ref.
func RegisterCreateWorkspace(g *genkit.Genkit) genkitai.ToolRef {
	createWorkspaceRef = genkit.DefineTool(g, ToolCreateWorkspace,
		"Creates a new workspace for the current user and adds them as owner. Use when the user asks to create a workspace. Input: name.",
		func(tc *genkitai.ToolContext, input CreateWorkspaceInput) (string, error) {
			name := strings.TrimSpace(input.Name)

			log.Info().
				Str("tool", ToolCreateWorkspace).
				Str("name", name).
				Msg("createWorkspace tool invoked")

			if name == "" {
				log.Warn().
					Str("tool", ToolCreateWorkspace).
					Msg("createWorkspace: name is empty")
				return "", fmt.Errorf("name is required and cannot be empty")
			}
			if len(name) > 255 {
				log.Warn().
					Str("tool", ToolCreateWorkspace).
					Int("len", len(name)).
					Msg("createWorkspace: name exceeds 255 characters")
				return "", fmt.Errorf("name must be at most 255 characters")
			}

			dbConn := db.GetDB()
			if dbConn == nil {
				log.Error().
					Str("tool", ToolCreateWorkspace).
					Msg("createWorkspace: database not available")
				return "", fmt.Errorf("database not available")
			}

			userID := UserIDFromContext(tc)
			if userID == "" {
				log.Warn().
					Str("tool", ToolCreateWorkspace).
					Msg("createWorkspace: user id not found in context")
				return "", fmt.Errorf("user id not found")
			}

			wsModule := module.NewWorkspace(dbConn)
			workspace, err := wsModule.CreateWorkspace(&module.CreateWorkspaceOptions{Name: name})
			if err != nil {
				log.Error().
					Str("tool", ToolCreateWorkspace).
					Err(err).
					Msg("createWorkspace: failed to create workspace")
				return "", fmt.Errorf("failed to create workspace: %w", err)
			}

			_, err = wsModule.AddWorkspaceUser(workspace.ID, userID, db.UserRoleOwner)
			if err != nil {
				log.Error().
					Str("tool", ToolCreateWorkspace).
					Str("workspaceID", workspace.ID.String()).
					Err(err).
					Msg("createWorkspace: failed to add user as owner")
				return "", fmt.Errorf("failed to add you as workspace owner: %w", err)
			}

			log.Info().
				Str("tool", ToolCreateWorkspace).
				Str("workspaceID", workspace.ID.String()).
				Str("userID", userID.String()).
				Msg("Workspace created and user added as owner")

			return fmt.Sprintf("Workspace %q created successfully. You are the owner.", workspace.Name), nil
		})
	return createWorkspaceRef
}

// CreateWorkspaceToolRef returns the createWorkspace tool ref.
func CreateWorkspaceToolRef() genkitai.ToolRef {
	return createWorkspaceRef
}
