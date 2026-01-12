// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const GithubURL = "https://api.github.com"

func (c *Client) getEndpointPath(suffix string) string {
	return fmt.Sprintf(
		"%s/repos/%s/%s%s",
		GithubURL,
		c.Author,
		c.Repository,
		suffix,
	)
}

func (c *Client) CreateCheckRun(ctx context.Context, p CheckRunPayload) (CheckRun, error) {
	raw, err := json.Marshal(&p)
	if err != nil {
		return CheckRun{}, err
	}

	u := c.getEndpointPath("/check-runs")
	b, st, err := c.Request(
		ctx,
		http.MethodPost,
		u,
		bytes.NewReader(raw),
		"application/json",
	)
	if err != nil {
		return CheckRun{}, err
	}
	if st != http.StatusCreated {
		return CheckRun{}, fmt.Errorf(
			"github: HTTP %d: %s",
			st,
			b,
		)
	}

	var out CheckRun
	if err := json.Unmarshal(b, &out); err != nil {
		return CheckRun{}, err
	}
	return out, nil
}

func (c *Client) UpdateCheckRun(ctx context.Context, id int, p CheckRunPayload) (CheckRun, error) {
	raw, err := json.Marshal(&p)
	if err != nil {
		return CheckRun{}, err
	}

	u := c.getEndpointPath(fmt.Sprintf("/check-runs/%d", id))
	b, st, err := c.Request(
		ctx,
		http.MethodPatch,
		u,
		bytes.NewReader(raw),
		"application/json",
	)
	if err != nil {
		return CheckRun{}, err
	}
	if st != http.StatusOK {
		return CheckRun{}, fmt.Errorf(
			"github: HTTP %d: %s",
			st,
			b,
		)
	}

	var out CheckRun
	if err := json.Unmarshal(b, &out); err != nil {
		return CheckRun{}, err
	}
	return out, nil
}
