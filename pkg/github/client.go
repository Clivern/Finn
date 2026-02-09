// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the GitHub client.
type Client struct {
	Token      string
	Author     string
	Repository string
	httpClient *http.Client
}

// NewClient returns a client for owner/repo with the given token.
func NewClient(token, owner, repo string) *Client {
	return &Client{
		Token:      token,
		Author:     owner,
		Repository: repo,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Request performs an authenticated HTTP call to github API.
func (c *Client) Request(ctx context.Context, method, url string, body io.Reader, contentType string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", "Bearer", c.Token))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "finn-pkg-github")
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return b, resp.StatusCode, nil
}
