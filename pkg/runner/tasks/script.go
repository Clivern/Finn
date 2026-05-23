// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

// RunScript runs a script template task.
func RunScript(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	if IsDockerAvailable() {
		return RunScriptDocker(ctx, deps, task, dir)
	}

	return RunScriptLocal(ctx, deps, task, dir)
}

func RunScriptLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	command, err := BuildScriptExec(dir, task.Payload)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, command.Engine, append([]string{command.ScriptPath}, command.Args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), command.Env...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("script: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("script: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("script: start local: %w", err)
	}

	done := make(chan struct{})
	go appendScriptOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go appendScriptOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("script: local: %w", err)
	}
	return nil
}

func RunScriptDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	command, err := BuildScriptExec(dir, task.Payload)
	if err != nil {
		return err
	}

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     dir,
		Image:   scriptDockerImage(command.Engine, StringFromPayload(task.Payload, "script.engineVersion")),
		Command: append([]string{command.Engine, command.ScriptPath}, command.Args...),
		Env:     envListToMap(command.Env),
		Timeout: remaining(ctx, defaultTemplateTimeout),
		OnOutput: func(stream, line string) {
			_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
				"stream": stream,
				"line":   line,
				"msg":    "template",
			})
		},
	})
}

func appendScriptOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		_ = deps.Worker.AppendTaskLog(taskID, map[string]string{
			"stream": stream,
			"line":   line,
			"msg":    "template",
		})
	}
}

func envListToMap(env []string) map[string]string {
	if len(env) == 0 {
		return nil
	}

	out := make(map[string]string)
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func scriptDockerImage(engine, version string) string {
	version = strings.TrimSpace(version)
	switch engine {
	case "sh":
		return "docker.io/alpine:3.20"
	case "python", "python3":
		if version != "" {
			return "docker.io/python:" + version + "-slim"
		}
		return "docker.io/python:3.12-slim"
	case "node":
		if version != "" {
			return "docker.io/node:" + version + "-bookworm-slim"
		}
		return "docker.io/node:20-bookworm-slim"
	default:
		return "docker.io/bash:5.2"
	}
}
