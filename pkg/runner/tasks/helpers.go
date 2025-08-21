// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import "strconv"

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
