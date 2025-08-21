// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tool

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/module"
	genkitai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/rs/zerolog/log"
)

const (
	// ToolSummarizePlatformSettings is the Genkit tool name.
	ToolSummarizePlatformSettings = "summarizePlatformSettings"
)

// SummarizePlatformSettingsInput has no required fields.
type SummarizePlatformSettingsInput struct{}

var summarizePlatformSettingsRef genkitai.ToolRef

// RegisterSummarizePlatformSettings registers the tool and returns its ref.
func RegisterSummarizePlatformSettings(g *genkit.Genkit) genkitai.ToolRef {
	summarizePlatformSettingsRef = genkit.DefineTool(g, ToolSummarizePlatformSettings,
		`Returns the Finn system settings as JSON (platform name, URL, email, maintenance mode, maintenance_mode_enabled, SMTP configuration, smtp_enabled).
		Use when the user asks about system settings, configuration, maintenance mode, SMTP status, or how Finn is configured.
		Summarize the returned JSON for the user. Sensitive values (e.g. SMTP password) are masked.`,
		func(tc *genkitai.ToolContext, input SummarizePlatformSettingsInput) (string, error) {
			log.Info().
				Str("tool", ToolSummarizePlatformSettings).
				Msg("summarizePlatformSettings tool invoked")

			dbConn := db.GetDB()
			if dbConn == nil {
				log.Error().
					Str("tool", ToolSummarizePlatformSettings).
					Msg("summarizePlatformSettings: database not available")
				return "", fmt.Errorf("database not available")
			}

			id := UserIDFromContext(tc)
			if id == "" {
				log.Warn().
					Str("tool", ToolSummarizePlatformSettings).
					Msg("summarizePlatformSettings: user id not found in context")
				return "", fmt.Errorf("user id not found")
			}

			userModule := module.NewUser(db.NewUserRepository(dbConn))
			user, err := userModule.GetUser(id)
			if err != nil || user == nil {
				log.Warn().
					Str("tool", ToolSummarizePlatformSettings).
					Str("userID", id.String()).
					Err(err).
					Msg("summarizePlatformSettings: user not found")
				return "", fmt.Errorf("user not found")
			}

			if user.Role != db.UserRoleAdmin {
				log.Warn().
					Str("tool", ToolSummarizePlatformSettings).
					Str("userID", id.String()).
					Str("role", user.Role).
					Msg("summarizePlatformSettings: only admins can view system settings")
				return "", fmt.Errorf("only admins can view system settings")
			}

			settingsModule := module.NewSettings(db.NewOptionRepository(dbConn))
			settings, err := settingsModule.GetSettings()
			if err != nil {
				log.Error().
					Str("tool", ToolSummarizePlatformSettings).
					Err(err).
					Msg("summarizePlatformSettings: failed to get settings")
				return "", fmt.Errorf("failed to get settings: %w", err)
			}

			settingsJSON := settingsToJSON(settings)
			return settingsJSON, nil
		})
	return summarizePlatformSettingsRef
}

// settingsToJSON returns settings as JSON with sensitive values masked.
func settingsToJSON(settings map[string]string) string {
	out := make(map[string]string, len(settings)+2)
	for k, v := range settings {
		out[k] = strings.TrimSpace(v)
	}
	if strings.TrimSpace(settings["smtp_password"]) != "" {
		out["smtp_password"] = "***"
	}
	if strings.TrimSpace(settings["smtp_server"]) != "" {
		out["smtp_enabled"] = "true"
	} else {
		out["smtp_enabled"] = "false"
	}
	if strings.TrimSpace(settings["maintenance_mode"]) == "1" {
		out["maintenance_mode_enabled"] = "true"
	} else {
		out["maintenance_mode_enabled"] = "false"
	}
	raw, _ := json.Marshal(out)
	return string(raw)
}
