// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

type scriptRunConfig struct {
	engine        string
	engineVersion string
	scriptRel     string
	hostScript    string
	dockerScript  string
	baseClean     string
}

func runScriptEngine(ctx context.Context, deps *runner.Deps, task *runner.Task, workRoot string, p map[string]interface{}) error {
	sm := MapFromPayload(p, "script")
	cfg, err := scriptRunConfigFromPayload(workRoot, sm)
	if err != nil {
		return err
	}

	mode := strings.ToLower(StringFromMap(sm, "mode"))
	switch mode {
	case "local":
		return runScriptEngineLocal(ctx, deps, task, workRoot, p, cfg)
	default:
		return runScriptEngineDocker(ctx, deps, task, workRoot, p, cfg)
	}
}

func runScriptEngineLocal(
	ctx context.Context,
	deps *runner.Deps,
	task *runner.Task,
	workRoot string,
	p map[string]interface{},
	cfg scriptRunConfig,
) error {
	cmdArgs := scriptCommand(cfg.engine, cfg.hostScript)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = workRoot
	cmd.Env = append(os.Environ(), envList(EnvFromPayload(p))...)

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
	go scanScriptOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go scanScriptOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("script: local: %w", err)
	}
	return nil
}

func runScriptEngineDocker(
	ctx context.Context,
	deps *runner.Deps,
	task *runner.Task,
	workRoot string,
	p map[string]interface{},
	cfg scriptRunConfig,
) error {
	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     workRoot,
		Image:   scriptDockerImage(cfg.engine, cfg.engineVersion),
		Command: scriptCommand(cfg.engine, cfg.dockerScript),
		Env:     EnvFromPayload(p),
		Timeout: remaining(ctx, defaultTemplateTimeout),
		OnOutput: func(stream, line string) {
			appendTemplateLog(deps, task.ID, stream, line)
		},
	})
}

func scriptRunConfigFromPayload(workRoot string, sm map[string]interface{}) (scriptRunConfig, error) {
	if len(sm) == 0 {
		return scriptRunConfig{}, fmt.Errorf("template: missing script configuration")
	}
	engine := strings.ToLower(StringFromMap(sm, "engine"))
	if engine == "" {
		engine = "bash"
	}
	scriptRel := strings.TrimSpace(StringFromMap(sm, "scriptPath"))
	if scriptRel == "" {
		return scriptRunConfig{}, fmt.Errorf("template: script path is required")
	}

	hostScript := filepath.Join(workRoot, filepath.FromSlash(scriptRel))
	hostScript = filepath.Clean(hostScript)
	baseClean := filepath.Clean(workRoot)
	if rel, err := filepath.Rel(baseClean, hostScript); err != nil || strings.HasPrefix(rel, "..") {
		return scriptRunConfig{}, fmt.Errorf("template: invalid script path")
	}
	if _, err := os.Stat(hostScript); err != nil {
		return scriptRunConfig{}, fmt.Errorf("template: script file: %w", err)
	}

	dockerScript := "/workspace/" + filepath.ToSlash(strings.TrimPrefix(
		strings.TrimPrefix(hostScript, baseClean), string(filepath.Separator)))

	return scriptRunConfig{
		engine:        engine,
		engineVersion: StringFromMap(sm, "engineVersion"),
		scriptRel:     scriptRel,
		hostScript:    hostScript,
		dockerScript:  dockerScript,
		baseClean:     baseClean,
	}, nil
}

func scriptCommand(engine, scriptPath string) []string {
	switch engine {
	case "python3":
		return []string{"python3", scriptPath}
	case "node":
		return []string{"node", scriptPath}
	case "sh":
		return []string{"sh", scriptPath}
	default:
		return []string{"bash", scriptPath}
	}
}

func scriptDockerImage(engine, version string) string {
	v := strings.TrimSpace(version)
	switch engine {
	case "sh":
		return "docker.io/alpine:3.20"
	case "python3":
		if v != "" {
			return "docker.io/python:" + v + "-slim"
		}
		return "docker.io/python:3.12-slim"
	case "node":
		if v != "" {
			return "docker.io/node:" + v + "-bookworm-slim"
		}
		return "docker.io/node:20-bookworm-slim"
	default:
		return "docker.io/bash:5.2"
	}
}

func envList(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

func scanScriptOutput(
	r interface{ Read([]byte) (int, error) },
	stream string,
	deps *runner.Deps,
	taskID string,
	done chan<- struct{},
) {
	defer func() { done <- struct{}{} }()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			appendTemplateLog(deps, taskID, stream, line)
		}
	}
}
