// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"github.com/clivern/finn/pkg/runner"
)

// init registers the clone task handler.
func init() {
	runner.RegisterTaskHandler(runner.TaskTypeClone, Clone)
}
