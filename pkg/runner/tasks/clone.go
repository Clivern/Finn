// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

// Clone clones a repository using go-git and streams progress to the task log.
func Clone(ctx context.Context, deps *runner.Deps, task *runner.Task) error {
	dir, err := os.MkdirTemp("", task.ID)
	if err != nil {
		return fmt.Errorf("clone: create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	commit, err := workspace.Clone(ctx, &workspace.CloneOptions{
		URL:           StringFromPayload(task.Payload, "repository"),
		Branch:        StringFromPayload(task.Payload, "branch"),
		Tag:           StringFromPayload(task.Payload, "tag"),
		Dir:           dir,
		Timeout:       time.Duration(IntFromPayload(task.Payload, "timeout")) * time.Second,
		Username:      StringFromPayload(task.Payload, "username"),
		Password:      StringFromPayload(task.Payload, "password"),
		SSHPrivateKey: StringFromPayload(task.Payload, "ssh_private_key"),
		SSHPassphrase: StringFromPayload(task.Payload, "ssh_passphrase"),
		OnProgress: func(msg string) {
			_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
				"stream": "stdout",
				"line":   msg,
				"msg":    "clone",
			})
		},
	})
	if err != nil {
		return err
	}

	_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "HEAD is at " + commit,
		"msg":    "clone",
	})

	return nil
}
