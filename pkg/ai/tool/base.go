// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tool

import (
	"context"

	"github.com/clivern/finn/db"
)

// UserIDContextKey is the context key for the current user ID.
const UserIDContextKey = "user_id"

// UserIDFromContext returns the user ID from context (set by Generate).
func UserIDFromContext(ctx context.Context) db.ID {
	if id, ok := ctx.Value(UserIDContextKey).(db.ID); ok {
		return id
	}
	return ""
}
