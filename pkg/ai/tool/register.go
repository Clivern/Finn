// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tool

import (
	genkitai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// RegisterTools registers all Finn AI tools and returns their refs.
func RegisterTools(g *genkit.Genkit) []genkitai.ToolRef {
	return []genkitai.ToolRef{
		RegisterUpdateUserName(g),
		RegisterSummarizePlatformSettings(g),
		RegisterGetMaintenanceModeEnableInstructions(g),
		RegisterCreateWorkspace(g),
	}
}
