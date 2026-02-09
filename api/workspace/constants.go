// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

const (
	// backendLockHistoryRetentionDays is how many days of inactive
	// Terraform backend lock rows we keep before cleanup.
	backendLockHistoryRetentionDays = 30
)
