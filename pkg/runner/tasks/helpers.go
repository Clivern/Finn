// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// StringFromPayload returns a string from the payload map. Dot paths read nested maps.
func StringFromPayload(p map[string]interface{}, key string) string {
	v, ok := valueFromPayload(p, key)
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	case map[string]interface{}, map[string]string:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

// IntFromPayload returns an integer from the payload map. Dot paths read nested maps.
func IntFromPayload(p map[string]interface{}, key string) int {
	v, ok := valueFromPayload(p, key)
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(n))
		return i
	}
	return 0
}

// BoolFromPayload parses a boolean from the payload map. Dot paths read nested maps.
func BoolFromPayload(p map[string]interface{}, key string) bool {
	v, ok := valueFromPayload(p, key)
	if !ok || v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(strings.TrimSpace(b), "true") || strings.TrimSpace(b) == "1"
	default:
		return false
	}
}

// ArrayStringFromPayload returns a string slice from the payload map. Dot paths read nested maps.
func ArrayStringFromPayload(p map[string]interface{}, key string) []string {
	v, ok := valueFromPayload(p, key)
	if !ok || v == nil {
		return nil
	}

	items, ok := v.([]string)
	if !ok {
		return nil
	}
	return items
}

// MapFromPayload returns a nested JSON object as a map. Dot paths read nested maps.
func MapFromPayload(p map[string]interface{}, key string) map[string]interface{} {
	v, ok := valueFromPayload(p, key)
	if !ok || v == nil {
		return nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}

// StringFromMap returns a trimmed string from a nested map (template script/iac/ansible JSON).
func StringFromMap(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	default:
		return strings.TrimSpace(fmt.Sprint(s))
	}
}

// BoolFromMap parses a bool from a nested map.
func BoolFromMap(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(strings.TrimSpace(b), "true") || strings.TrimSpace(b) == "1"
	default:
		return false
	}
}

// EnvFromPayload returns string env vars from payload["env"] (map[string]interface{} or map[string]string).
func EnvFromPayload(p map[string]interface{}) map[string]string {
	v, ok := p["env"]
	if !ok || v == nil {
		return nil
	}
	out := make(map[string]string)
	switch m := v.(type) {
	case map[string]interface{}:
		for k, val := range m {
			if val == nil {
				continue
			}
			out[k] = strings.TrimSpace(fmt.Sprint(val))
		}
	case map[string]string:
		for k, val := range m {
			out[k] = strings.TrimSpace(val)
		}
	default:
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// IsDockerAvailable checks if Docker is available on the system.
func IsDockerAvailable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return exec.CommandContext(ctx, "docker", "info").Run() == nil
}

// IacStepArgs returns a string slice of step arguments for the IAC engine.
func IacStepArgs(steps []string) ([][]string, error) {
	out := make([][]string, 0, len(steps))
	for _, step := range steps {
		switch strings.TrimSpace(step) {
		case "init":
			out = append(out, []string{"init", "-input=false"})
		case "plan":
			out = append(out, []string{"plan", "-input=false"})
		case "apply":
			out = append(out, []string{"apply", "-input=false", "-auto-approve"})
		default:
			return nil, fmt.Errorf("iac: unsupported step %q", step)
		}
	}
	return out, nil
}

func valueFromPayload(p map[string]interface{}, key string) (interface{}, bool) {
	if p == nil || strings.TrimSpace(key) == "" {
		return nil, false
	}

	parts := strings.Split(key, ".")
	var current interface{} = p
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, false
		}

		switch m := current.(type) {
		case map[string]interface{}:
			v, ok := m[part]
			if !ok {
				return nil, false
			}
			current = v
		case map[string]string:
			v, ok := m[part]
			if !ok {
				return nil, false
			}
			current = v
		default:
			return nil, false
		}
	}

	if current == nil {
		return nil, false
	}
	return current, true
}
