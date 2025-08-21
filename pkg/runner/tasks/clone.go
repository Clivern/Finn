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
	"github.com/clivern/finn/pkg/runner/docker"
)

// Clone clones a repository using docker.CloneInDocker and streams output to the task log.
func Clone(deps *runner.Deps, task *runner.Task) error {
	dir, err := os.MkdirTemp("", task.ID)
	if err != nil {
		return fmt.Errorf("clone: create temp dir: %w", err)
	}

	timeoutSec := IntFromPayload(task.Payload, "timeout")
	if timeoutSec <= 0 {
		timeoutSec = 300
	}

	events := make(chan docker.StreamEvent, 256)
	go func() {
		for e := range events {
			_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
				"stream": e.Stream,
				"line":   e.Line,
				"msg":    e.Msg,
			})
		}
	}()

	err = docker.CloneInDocker(context.Background(), &docker.Cloner{
		RepoURL:       StringFromPayload(task.Payload, "repository"),
		Branch:        StringFromPayload(task.Payload, "branch"),
		Tag:           StringFromPayload(task.Payload, "tag"),
		Dir:           dir,
		Timeout:       time.Duration(timeoutSec) * time.Second,
		StreamMessage: "clone",
		Username:      StringFromPayload(task.Payload, "username"),
		Password:      StringFromPayload(task.Payload, "password"),
		SSHPrivateKey: StringFromPayload(task.Payload, "ssh_private_key"),
		StreamHandler: func(e docker.StreamEvent) {
			events <- e
		},
	})
	close(events)

	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	return nil
}
