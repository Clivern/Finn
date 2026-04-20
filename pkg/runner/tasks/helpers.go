// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"strconv"
	"strings"
)

// StringFromPayload returns the string value from the payload map.
func StringFromPayload(p map[string]interface{}, key string) string {
	if v, ok := p[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// IntFromPayload returns the integer value from the payload map.
func IntFromPayload(p map[string]interface{}, key string) int {
	if v, ok := p[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case string:
			i, _ := strconv.Atoi(n)
			return i
		}
	}
	return 0
}

// BoolFromPayload parses a boolean from the payload map (JSON often uses bool or string).
func BoolFromPayload(p map[string]interface{}, key string) bool {
	v, ok := p[key]
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

// MapFromPayload returns a nested JSON object as a map.
func MapFromPayload(p map[string]interface{}, key string) map[string]interface{} {
	if p == nil {
		return nil
	}
	v, ok := p[key]
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
