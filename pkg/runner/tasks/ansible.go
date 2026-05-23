// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"

	"github.com/clivern/finn/pkg/runner"
)

// RunAnsible runs an Ansible template task.
func RunAnsible(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "Ansible engine placeholder",
		"msg":    "template",
	})
	return nil
}
