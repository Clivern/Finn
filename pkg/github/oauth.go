// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	GithubOAuthURL   = "https://github.com/login/oauth/authorize"
	OAuthAccessToken = "https://github.com/login/oauth/access_token"
)

// OAuthApp is the GitHub OAuth app (authorize URL + token exchange).
type OAuthApp struct {
	ClientID     string
	RedirectURI  string
	Scope        string
	Scopes       []string
	State        string
	AllowSignup  string
	ClientSecret string
	AccessToken  string
	TokenType    string
	httpClient   *http.Client `json:"-"`
}

// NewOAuthApp returns an app with client ID, secret, and redirect URI; HTTP client has a 60s timeout.
func NewOAuthApp(clientID, clientSecret, redirectURI string) *OAuthApp {
	return &OAuthApp{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

// GenerateState generates a random state for the OAuth app.
func (e *OAuthApp) GenerateState() {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return
	}
	e.State = hex.EncodeToString(b)
}

// AddScope adds a scope to the OAuth app.
func (e *OAuthApp) AddScope(scope string) {
	e.Scopes = append(e.Scopes, scope)
	e.Scope = strings.Join(e.Scopes, ",")
}

// AddScopes adds multiple scopes to the OAuth app.
func (e *OAuthApp) AddScopes(scopes []string) {
	e.Scopes = scopes
	e.Scope = strings.Join(e.Scopes, ",")
}

// BuildAuthorizeURL builds the authorize URL for the OAuth app.
func (e *OAuthApp) BuildAuthorizeURL() string {
	e.Scope = strings.Join(e.Scopes, ",")
	u, _ := url.Parse(GithubOAuthURL)
	q := u.Query()
	q.Set("client_id", e.ClientID)
	q.Set("redirect_uri", e.RedirectURI)
	q.Set("scope", e.Scope)
	q.Set("state", e.State)
	q.Set("allow_signup", e.AllowSignup)
	u.RawQuery = q.Encode()
	return u.String()
}

// FetchAccessToken fetches the access token for the OAuth app.
func (e *OAuthApp) FetchAccessToken(ctx context.Context, code, state string) error {
	if state != e.State {
		return fmt.Errorf("oauth state mismatch")
	}
	payload, err := json.Marshal(struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		State        string `json:"state"`
	}{
		ClientID:     e.ClientID,
		ClientSecret: e.ClientSecret,
		Code:         code,
		State:        e.State,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		OAuthAccessToken,
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("oauth HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	e.AccessToken = out.AccessToken
	e.TokenType = out.TokenType
	return nil
}

// GetAccessToken returns the access token for the OAuth app.
func (e *OAuthApp) GetAccessToken() string {
	return e.AccessToken
}

// GetTokenType returns the token type for the OAuth app.
func (e *OAuthApp) GetTokenType() string {
	return e.TokenType
}
