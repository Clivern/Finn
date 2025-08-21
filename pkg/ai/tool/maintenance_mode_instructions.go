// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tool

import (
	genkitai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/rs/zerolog/log"
)

const (
	// ToolGetMaintenanceModeEnableInstructions is the Genkit tool name.
	ToolGetMaintenanceModeEnableInstructions = "getMaintenanceModeEnableInstructions"
)

// GetMaintenanceModeEnableInstructionsInput has no required fields.
type GetMaintenanceModeEnableInstructionsInput struct{}

var getMaintenanceModeEnableInstructionsRef genkitai.ToolRef

// RegisterGetMaintenanceModeEnableInstructions registers the tool and returns its ref.
func RegisterGetMaintenanceModeEnableInstructions(g *genkit.Genkit) genkitai.ToolRef {
	getMaintenanceModeEnableInstructionsRef = genkit.DefineTool(g, ToolGetMaintenanceModeEnableInstructions,
		`Returns instructions for enabling maintenance mode in Finn.
When the user asks whether maintenance mode is enabled or disabled: call summarizePlatformSettings first. If the result shows maintenance_mode_enabled is false, call this tool and include the returned instructions in your response.`,
		func(tc *genkitai.ToolContext, input GetMaintenanceModeEnableInstructionsInput) (string, error) {
			log.Info().
				Str("tool", ToolGetMaintenanceModeEnableInstructions).
				Msg("getMaintenanceModeEnableInstructions tool invoked")

			return `To enable maintenance mode: open the Finn dashboard, go to Settings (sidebar or /settings),
					turn on the Maintenance mode option, and save.
					While enabled, only admins can sign in; other users see a maintenance message.`, nil
		})
	return getMaintenanceModeEnableInstructionsRef
}
