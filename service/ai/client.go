// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ai

import (
	"context"
	"fmt"

	genkitai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

const (
	ProviderOpenAI    = "openai"
	ProviderGoogleAI  = "googleai"
	ProviderVertexAI  = "vertexai"
	ProviderAnthropic = "anthropic"
)

const (
	DefaultModelOpenAI    = "openai/gpt-4o"
	DefaultModelGoogleAI  = "googleai/gemini-2.0-flash"
	DefaultModelAnthropic = "anthropic/claude-3-5-sonnet-20241022"
	DefaultModel          = DefaultModelOpenAI
)

// ProviderConfig configures one AI provider (API key or Vertex project/location).
type ProviderConfig struct {
	Provider string
	APIKey   string
	BaseURL  string

	// For Vertex AI only:
	ProjectID string
	Location  string
}

// GenerateInput holds system prompt, instructions, context, and prompt.
type GenerateInput struct {
	SystemPrompt string
	Instructions string
	Context      string
	Prompt       string
}

// Client wraps a Genkit instance for AI text generation with one or more providers.
type Client struct {
	client *genkit.Genkit
}

// New creates an AI client with the given provider configs and model name.
func New(ctx context.Context, modelName string, providerConfig ProviderConfig) (*Client, error) {
	var plugin api.Plugin

	switch providerConfig.Provider {
	case ProviderOpenAI:
		plugin = &openai.OpenAI{APIKey: providerConfig.APIKey}
	case ProviderGoogleAI:
		plugin = &googlegenai.GoogleAI{APIKey: providerConfig.APIKey}
	case ProviderVertexAI:
		plugin = &googlegenai.VertexAI{ProjectID: providerConfig.ProjectID, Location: providerConfig.Location}
	case ProviderAnthropic:
		plugin = &anthropic.Anthropic{APIKey: providerConfig.APIKey, BaseURL: providerConfig.BaseURL}
	default:
		return nil, fmt.Errorf("ai client: unknown provider %q", providerConfig.Provider)
	}

	client := genkit.Init(ctx, genkit.WithPlugins(plugin), genkit.WithDefaultModel(modelName))
	return &Client{client: client}, nil
}

// Generate runs generation with the given model name and input.
func (c *Client) Generate(ctx context.Context, modelName string, input *GenerateInput) (string, error) {
	opts := []genkitai.GenerateOption{
		genkitai.WithModelName(modelName),
		genkitai.WithSystem(input.SystemPrompt + "\n\n" + input.Instructions),
		genkitai.WithMessages(genkitai.NewUserTextMessage(input.Context)),
		genkitai.WithPrompt(input.Prompt),
	}

	resp, err := genkit.Generate(ctx, c.client, opts...)
	if err != nil {
		return "", fmt.Errorf("ai generate: %w", err)
	}
	return resp.Text(), nil
}
