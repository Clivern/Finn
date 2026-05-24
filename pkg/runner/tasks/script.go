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

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

var (
	scriptImages = map[string]map[string]string{
		"bash": {
			"":    "docker.io/bash:5.2",
			"5.2": "docker.io/bash:5.2",
			"5.3": "docker.io/bash:5.3",
		},
		"sh": {
			"":     "docker.io/alpine:3.20",
			"3.20": "docker.io/alpine:3.20",
			"3.21": "docker.io/alpine:3.21",
			"3.22": "docker.io/alpine:3.22",
			"3.23": "docker.io/alpine:3.23",
		},
		"python": {
			"":     "docker.io/python:3.12-slim",
			"3.10": "docker.io/python:3.10-slim",
			"3.11": "docker.io/python:3.11-slim",
			"3.12": "docker.io/python:3.12-slim",
			"3.13": "docker.io/python:3.13-slim",
			"3.14": "docker.io/python:3.14-slim",
		},
		"node": {
			"":   "docker.io/node:20-bookworm-slim",
			"18": "docker.io/node:18-bookworm-slim",
			"20": "docker.io/node:20-bookworm-slim",
			"22": "docker.io/node:22-bookworm-slim",
			"24": "docker.io/node:24-bookworm-slim",
		},
	}
)

type ScriptExec struct {
	Engine        string
	EngineVersion string
	ScriptRelPath string
	ScriptPath    string
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

	image, err := scriptDockerImage(command.Engine, command.EngineVersion)
	if err != nil {
		return err
	}

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     dir,
		Image:   image,
		Command: append([]string{command.Engine, DockerScriptPath(command.ScriptRelPath)}, command.Args...),
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

func scriptDockerImage(engine, version string) (string, error) {
	versions, ok := scriptImages[engine]
	if !ok {
		return "", fmt.Errorf("script: unsupported engine %q", engine)
	}

	image, ok := versions[version]
	if !ok {
		return "", fmt.Errorf("script: unsupported %s version %q", engine, version)
	}
	return image, nil
}

func BuildScriptExec(dir string, payload map[string]interface{}) (*ScriptExec, error) {
	env := append(
		keyValueArgs(MapFromPayload(payload, "groupVars.environmentVariables")),
		keyValueArgs(MapFromPayload(payload, "groupVars.secretEnvironmentVariables"))...,
	)
	args := append(
		append(
			keyValueArgs(MapFromPayload(payload, "extraArgs")),
			keyValueArgs(MapFromPayload(payload, "groupVars.extraVariables"))...,
		),
		keyValueArgs(MapFromPayload(payload, "groupVars.secretExtraVariables"))...,
	)

	return &ScriptExec{
		Engine:        StringFromPayload(payload, "script.engine"),
		EngineVersion: StringFromPayload(payload, "script.engineVersion"),
		ScriptRelPath: StringFromPayload(payload, "script.scriptPath"),
		ScriptPath: filepath.Join(
			dir,
			filepath.FromSlash(StringFromPayload(payload, "script.scriptPath")),
		),
		Env:  env,
		Args: args,
	}, nil
}

func DockerScriptPath(relPath string) string {
	relPath = strings.TrimPrefix(filepath.ToSlash(relPath), "/")
	return "/workspace/" + relPath
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
