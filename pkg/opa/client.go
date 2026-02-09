// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package opa

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/rego"
)

const (
	DataPathAnsible    = "ansible/policy"
	DataPathTerraform  = "terraform/policy"
	DataPathTerragrunt = "terragrunt/policy"
)

// PolicyResult is the document returned by a Rego policy package (allow + deny).
type PolicyResult struct {
	Allow bool
	Deny  []string
}

// Client holds Rego modules and evaluates policies in-process.
type Client struct {
	modules  map[string]string // module name → rego source
	mu       sync.RWMutex
	prepared map[string]rego.PreparedEvalQuery
}

// NewClient creates a client from Rego source strings
func NewClient(modules map[string]string) *Client {
	return &Client{
		modules:  modules,
		prepared: make(map[string]rego.PreparedEvalQuery),
	}
}

// Evaluate runs the Rego policy at dataPath (e.g. DataPathAnsible) with input
func (c *Client) Evaluate(ctx context.Context, dataPath string, input map[string]interface{}) (*PolicyResult, error) {
	pq, err := c.prepare(ctx, dataPath)
	if err != nil {
		return nil, err
	}

	rs, err := pq.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("opa: eval %q: %w", dataPath, err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return nil, fmt.Errorf("opa: no result for %q (policy package not defined?)", dataPath)
	}

	doc, ok := rs[0].Expressions[0].Value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("opa: result for %q is not a document", dataPath)
	}
	return parseResult(doc), nil
}

// EvaluateChange is a convenience wrapper: returns allowed=true only
func (c *Client) EvaluateChange(ctx context.Context, dataPath string, input map[string]interface{}) (allowed bool, violations []string, err error) {
	res, err := c.Evaluate(ctx, dataPath, input)
	if err != nil {
		return false, nil, err
	}
	if !res.Allow || len(res.Deny) > 0 {
		return false, res.Deny, nil
	}
	return true, nil, nil
}

func (c *Client) prepare(ctx context.Context, dataPath string) (rego.PreparedEvalQuery, error) {
	c.mu.RLock()
	pq, ok := c.prepared[dataPath]
	c.mu.RUnlock()
	if ok {
		return pq, nil
	}

	query := "data." + strings.ReplaceAll(strings.Trim(dataPath, "/"), "/", ".")
	opts := []func(*rego.Rego){rego.Query(query)}
	for name, src := range c.modules {
		opts = append(opts, rego.Module(name, src))
	}

	pq, err := rego.New(opts...).PrepareForEval(ctx)
	if err != nil {
		return rego.PreparedEvalQuery{}, fmt.Errorf("opa: prepare %q: %w", dataPath, err)
	}

	c.mu.Lock()
	c.prepared[dataPath] = pq
	c.mu.Unlock()
	return pq, nil
}

func parseResult(doc map[string]interface{}) *PolicyResult {
	r := &PolicyResult{}
	if allow, ok := doc["allow"].(bool); ok {
		r.Allow = allow
	}
	switch d := doc["deny"].(type) {
	case []interface{}:
		for _, v := range d {
			if s, ok := v.(string); ok {
				r.Deny = append(r.Deny, s)
			}
		}
	case map[string]interface{}:
		for _, v := range d {
			if s, ok := v.(string); ok {
				r.Deny = append(r.Deny, s)
			}
		}
	}
	return r
}
