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
	// ToolUpdateUserName is the Genkit tool name.
	ToolUpdateUserName = "updateUserName"
)

// UpdateUserNameInput is the input for updateUserName.
type UpdateUserNameInput struct {
	Name string `json:"name" jsonschema_description:"New display name (max 255 chars)"`
}

var updateUserNameRef genkitai.ToolRef

// RegisterUpdateUserName registers the updateUserName tool and returns its ref.
func RegisterUpdateUserName(g *genkit.Genkit) genkitai.ToolRef {
	updateUserNameRef = genkit.DefineTool(g, ToolUpdateUserName,
		"Updates the current user's display name in their Finn profile. Use when the user asks to change their name, set their display name, or update their profile name. Requires the new name as input.",
		func(tc *genkitai.ToolContext, input UpdateUserNameInput) (string, error) {
			name := strings.TrimSpace(input.Name)

			log.Info().
				Str("tool", ToolUpdateUserName).
				Str("name", name).
				Msg("updateUserName tool invoked")

			if name == "" {
				log.Warn().
					Str("tool", ToolUpdateUserName).
					Msg("updateUserName: name is empty")
				return "", fmt.Errorf("name is required and cannot be empty")
			}
			if len(name) > 255 {
				log.Warn().
					Str("tool", ToolUpdateUserName).
					Str("name", name).
					Int("len", len(name)).
					Msg("updateUserName: name exceeds 255 characters")
				return "", fmt.Errorf("name must be at most 255 characters")
			}

			var user *db.User
			dbConn := db.GetDB()
			if dbConn == nil {
				log.Error().
					Str("tool", ToolUpdateUserName).
					Msg("updateUserName: database not available")
				return "", fmt.Errorf("database not available")
			}

			userModule := module.NewUser(db.NewUserRepository(dbConn))
			id := UserIDFromContext(tc)
			if id == "" {
				log.Warn().
					Str("tool", ToolUpdateUserName).
					Msg("updateUserName: user id not found in context")
				return "", fmt.Errorf("user id not found")
			}

			var err error
			user, err = userModule.GetUser(id)
			if err != nil || user == nil {
				log.Warn().
					Str("tool", ToolUpdateUserName).
					Str("userID", id.String()).
					Err(err).
					Msg("updateUserName: user not found")
				return "", fmt.Errorf("user not found")
			}

			log.Info().
				Str("tool", ToolUpdateUserName).
				Str("userID", user.ID.String()).
				Str("name", name).
				Str("previousName", user.Name).
				Msg("Updating user display name")

			updated, err := userModule.UpdateUser(&module.UpdateUserOptions{
				UserID:   user.ID,
				Name:     name,
				Email:    user.Email,
				Role:     user.Role,
				IsActive: user.IsActive,
			})
			if err != nil {
				log.Error().
					Str("tool", ToolUpdateUserName).
					Str("userID", user.ID.String()).
					Err(err).
					Msg("updateUserName: failed to update name")
				return "", fmt.Errorf("failed to update name: %w", err)
			}

			log.Info().
				Str("tool", ToolUpdateUserName).
				Str("userID", updated.ID.String()).
				Str("name", updated.Name).
				Msg("User display name updated successfully")

			return fmt.Sprintf("Display name updated successfully to %q.", updated.Name), nil
		})
	return updateUserNameRef
}

// UpdateUserNameToolRef returns the updateUserName tool ref.
func UpdateUserNameToolRef() genkitai.ToolRef {
	return updateUserNameRef
}
