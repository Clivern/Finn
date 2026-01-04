// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import "strings"

// IsAssetPath returns true if path is a static asset
func IsAssetPath(path string) bool {
	return strings.HasPrefix(path, "/assets/")
}
