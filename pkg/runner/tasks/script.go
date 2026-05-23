// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"

	"github.com/clivern/finn/pkg/runner"
)

// RunScript runs a script template task.
func RunScript(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	if IsDockerAvailable() {
		return RunScriptDocker(ctx, deps, task, dir)
	}

	return RunScriptLocal(ctx, deps, task, dir)
}

func RunScriptLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "script local engine placeholder",
		"msg":    "template",
	})
	return nil
}

func RunScriptDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "script docker engine placeholder",
		"msg":    "template",
	})
	return nil
}
