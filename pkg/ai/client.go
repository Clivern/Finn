// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ai

import (
	"context"
	"fmt"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/pkg/ai/tool"
	genkitai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/spf13/viper"
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

// ProviderConfig configures an AI provider.
type ProviderConfig struct {
	Provider string
	APIKey   string
	BaseURL  string

	// Vertex AI only
	ProjectID string
	Location  string
}

// GenerateInput is the input for Generate (prompt, context, tools, user).
type GenerateInput struct {
	SystemPrompt string
	Instructions string
	Context      string
	Prompt       string
	Tools        []genkitai.ToolRef
	UserID       db.ID
}

// Client is the Genkit-backed AI client.
type Client struct {
	client *genkit.Genkit
	tools  []genkitai.ToolRef
}

// New creates an AI client with the given provider and model.
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

	g := genkit.Init(ctx, genkit.WithPlugins(plugin), genkit.WithDefaultModel(modelName))
	tools := tool.RegisterTools(g)
	return &Client{client: g, tools: tools}, nil
}

// Genkit returns the underlying Genkit instance.
func (c *Client) Genkit() *genkit.Genkit {
	return c.client
}

// Tools returns the registered Finn tool refs for GenerateInput.Tools.
func (c *Client) Tools() []genkitai.ToolRef {
	return c.tools
}

// Generate runs generation with the given model and input.
func (c *Client) Generate(ctx context.Context, modelName string, input *GenerateInput) (string, error) {
	opts := []genkitai.GenerateOption{
		genkitai.WithModelName(modelName),
		genkitai.WithSystem(input.SystemPrompt + "\n\n" + input.Instructions),
		genkitai.WithMessages(genkitai.NewUserTextMessage(input.Context)),
		genkitai.WithPrompt(input.Prompt),
	}

	if len(input.Tools) > 0 {
		opts = append(opts, genkitai.WithTools(input.Tools...))
	}

	ctx = context.WithValue(ctx, tool.UserIDContextKey, input.UserID)

	resp, err := genkit.Generate(ctx, c.client, opts...)
	if err != nil {
		return "", fmt.Errorf("ai generate: %w", err)
	}
	return resp.Text(), nil
}

// IsEnabled returns whether AI is enabled in config.
func IsEnabled() bool {
	return viper.GetBool("app.ai.enabled")
}

// ModelName returns the configured model name or default.
func ModelName() string {
	if m := viper.GetString("app.ai.model"); m != "" {
		return m
	}
	return DefaultModel
}

// NewFromConfig creates a client from viper config.
func NewFromConfig(ctx context.Context) (*Client, error) {
	provider := viper.GetString("app.ai.provider")
	if provider == "" {
		provider = ProviderOpenAI
	}
	model := viper.GetString("app.ai.model")
	if model == "" {
		model = DefaultModel
	}
	cfg := ProviderConfig{
		Provider:  provider,
		APIKey:    viper.GetString("app.ai.api_key"),
		BaseURL:   viper.GetString("app.ai.base_url"),
		ProjectID: viper.GetString("app.ai.project_id"),
		Location:  viper.GetString("app.ai.location"),
	}
	return New(ctx, model, cfg)
}
