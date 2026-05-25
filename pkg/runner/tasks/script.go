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
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

var (
	// @TODO: Allow customers to override the script images
	scriptImages = map[string]map[string]string{
		"bash": {
			"5.2": "docker.io/bash:5.2",
			"5.3": "docker.io/bash:5.3",
		},
		"sh": {
			"3.20": "docker.io/alpine:3.20",
			"3.21": "docker.io/alpine:3.21",
			"3.22": "docker.io/alpine:3.22",
			"3.23": "docker.io/alpine:3.23",
		},
		"python": {
			"3.10": "docker.io/python:3.10-slim",
			"3.11": "docker.io/python:3.11-slim",
			"3.12": "docker.io/python:3.12-slim",
			"3.13": "docker.io/python:3.13-slim",
			"3.14": "docker.io/python:3.14-slim",
		},
		"node": {
			"18": "docker.io/node:18-bookworm-slim",
			"20": "docker.io/node:20-bookworm-slim",
			"22": "docker.io/node:22-bookworm-slim",
			"24": "docker.io/node:24-bookworm-slim",
		},
	}
)

type ScriptExec struct {
	Bin           string
	Version       string
	WorkDir       string
	WorkRelPath   string
	ScriptRelPath string
	Env           []string
	Args          []string
}

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

	cmd := exec.CommandContext(
		ctx, command.Bin, append(
			[]string{command.ScriptRelPath}, command.Args...,
		)...,
	)
	cmd.Dir = command.WorkDir
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

	versions, ok := scriptImages[command.Bin]
	if !ok {
		return fmt.Errorf("script: unsupported engine %q", command.Bin)
	}

	image, ok := versions[command.Version]
	if !ok {
		return fmt.Errorf("script: unsupported %s version %q", command.Bin, command.Version)
	}

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     dir,
		RelPath: command.WorkRelPath,
		Image:   image,
		Command: append([]string{command.Bin, command.ScriptRelPath}, command.Args...),
		Env:     envListToMap(command.Env),
		Timeout: time.Duration(IntFromPayload(task.Payload, "timeout")) * time.Second,
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

func BuildScriptExec(dir string, payload map[string]interface{}) (*ScriptExec, error) {
	env := append(
		keyValueArgs(MapFromPayload(payload, "group_vars.environment_variables")),
		keyValueArgs(MapFromPayload(payload, "group_vars.secret_environment_variables"))...,
	)
	args := append(
		append(
			keyValueArgs(MapFromPayload(payload, "extra_args")),
			keyValueArgs(MapFromPayload(payload, "group_vars.extra_variables"))...,
		),
		keyValueArgs(MapFromPayload(payload, "group_vars.secret_extra_variables"))...,
	)

	return &ScriptExec{
		Bin:     StringFromPayload(payload, "script.bin"),
		Version: StringFromPayload(payload, "script.version"),
		WorkDir: filepath.Join(
			dir, filepath.FromSlash(StringFromPayload(payload, "working_dir")),
		),
		WorkRelPath: strings.TrimPrefix(
			filepath.ToSlash(StringFromPayload(payload, "working_dir")),
			"/",
		),
		ScriptRelPath: StringFromPayload(payload, "script.path"),
		Env:           env,
		Args:          args,
	}, nil
}

func keyValueArgs(values map[string]interface{}) []string {
	if len(values) == 0 {
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	args := make([]string, 0, len(keys))
	for _, key := range keys {
		if values[key] == nil {
			continue
		}
		args = append(args, fmt.Sprintf("%s=%s", key, strings.TrimSpace(fmt.Sprint(values[key]))))
	}
	return args
}
