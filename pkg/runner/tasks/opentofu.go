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
	"strings"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

var (
	opentofuImages = map[string]string{
		"1.8":  "ghcr.io/opentofu/opentofu:1.8",
		"1.9":  "ghcr.io/opentofu/opentofu:1.9",
		"1.10": "ghcr.io/opentofu/opentofu:1.10",
	}

	opentofuSteps = [][]string{
		{"init", "-input=false"},
		{"plan", "-input=false"},
		{"apply", "-input=false", "-auto-approve"},
	}
)

type OpentofuExec struct {
	Version     string
	WorkDir     string
	WorkRelPath string
	RootDir     string
	Env         []string
	Args        []string
}

func RunOpentofu(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	if IsDockerAvailable() {
		return RunOpentofuDocker(ctx, deps, task, dir)
	}

	return RunOpentofuLocal(ctx, deps, task, dir)
}

func BuildOpentofuExec(dir string, payload map[string]interface{}) (*OpentofuExec, error) {
	env := append(
		keyValueArgs(MapFromPayload(payload, "group_vars.environment_variables")),
		keyValueArgs(MapFromPayload(payload, "group_vars.secret_environment_variables"))...,
	)
	env = append(env, fmt.Sprintf("TF_LOG=%s", strings.ToUpper(StringFromPayload(payload, "iac.log_level"))))

	args := append(
		append(
			opentofuVarFlags(MapFromPayload(payload, "extra_args")),
			opentofuVarFlags(MapFromPayload(payload, "group_vars.extra_variables"))...,
		),
		opentofuVarFlags(MapFromPayload(payload, "group_vars.secret_extra_variables"))...,
	)

	if StringFromPayload(payload, "iac.backend.id") != "" {
		env = append(env, "TF_HTTP_USERNAME="+StringFromPayload(payload, "iac.backend.username"))
		env = append(env, "TF_HTTP_PASSWORD="+StringFromPayload(payload, "iac.backend.password"))
		env = append(env, "TF_HTTP_ADDRESS="+StringFromPayload(payload, "iac.backend.http.address"))
		env = append(env, "TF_HTTP_LOCK_ADDRESS="+StringFromPayload(payload, "iac.backend.http.lock_address"))
		env = append(env, "TF_HTTP_UNLOCK_ADDRESS="+StringFromPayload(payload, "iac.backend.http.unlock_address"))
		env = append(env, "TF_HTTP_LOCK_METHOD="+StringFromPayload(payload, "iac.backend.http.lock_method"))
		env = append(env, "TF_HTTP_UNLOCK_METHOD="+StringFromPayload(payload, "iac.backend.http.unlock_method"))
	}

	return &OpentofuExec{
		Version: StringFromPayload(payload, "iac.version"),
		WorkDir: filepath.Join(
			dir, filepath.FromSlash(StringFromPayload(payload, "working_dir")),
		),
		WorkRelPath: strings.TrimPrefix(
			filepath.ToSlash(StringFromPayload(payload, "working_dir")),
			"/",
		),
		RootDir: dir,
		Env:     env,
		Args:    args,
	}, nil
}

func opentofuVarFlags(values map[string]interface{}) []string {
	pairs := keyValueArgs(values)
	if len(pairs) == 0 {
		return nil
	}

	flags := make([]string, 0, len(pairs)*2)
	for _, pair := range pairs {
		flags = append(flags, "-var", pair)
	}
	return flags
}

func RunOpentofuLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	for _, step := range opentofuSteps {
		if err := RunOpentofuLocalStep(ctx, deps, task, dir, step); err != nil {
			return err
		}
	}
	return nil
}

func RunOpentofuDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	for _, step := range opentofuSteps {
		if err := RunOpentofuDockerStep(ctx, deps, task, dir, step); err != nil {
			return err
		}
	}
	return nil
}

func RunOpentofuLocalStep(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, step []string) error {
	command, err := BuildOpentofuExec(dir, task.Payload)
	if err != nil {
		return err
	}

	args := append([]string{}, step...)
	args = append(args, command.Args...)

	cmd := exec.CommandContext(ctx, "tofu", args...)
	cmd.Dir = command.WorkDir
	cmd.Env = append(os.Environ(), command.Env...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("opentofu: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("opentofu: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("opentofu: start local: %w", err)
	}

	done := make(chan struct{})
	go appendOpentofuOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go appendOpentofuOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("opentofu: local %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func RunOpentofuDockerStep(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, step []string) error {
	command, err := BuildOpentofuExec(dir, task.Payload)
	if err != nil {
		return err
	}

	image, ok := opentofuImages[command.Version]
	if !ok {
		return fmt.Errorf("opentofu: unsupported version %q", command.Version)
	}

	args := append([]string{}, step...)
	args = append(args, command.Args...)

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     command.RootDir,
		RelPath: command.WorkRelPath,
		Image:   image,
		Command: args,
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

func appendOpentofuOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
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
