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
	terraformImages = map[string]string{
		"1.9":  "docker.io/hashicorp/terraform:1.9",
		"1.10": "docker.io/hashicorp/terraform:1.10",
		"1.11": "docker.io/hashicorp/terraform:1.11",
	}
)

type TerraformExec struct {
	Version     string
	WorkDir     string
	WorkRelPath string
	RootDir     string
	Env         []string
	Args        []string
}

func RunTerraform(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, steps []string) error {
	stepArgs, err := IacStepArgs(steps)
	if err != nil {
		return err
	}

	if IsDockerAvailable() {
		return RunTerraformDocker(ctx, deps, task, dir, stepArgs)
	}

	return RunTerraformLocal(ctx, deps, task, dir, stepArgs)
}

func BuildTerraformExec(dir string, payload map[string]interface{}) (*TerraformExec, error) {
	env := append(
		keyValueArgs(MapFromPayload(payload, "group_vars.environment_variables")),
		keyValueArgs(MapFromPayload(payload, "group_vars.secret_environment_variables"))...,
	)
	env = append(env, fmt.Sprintf("TF_LOG=%s", strings.ToUpper(StringFromPayload(payload, "iac.log_level"))))

	args := append(
		append(
			terraformVarFlags(MapFromPayload(payload, "extra_args")),
			terraformVarFlags(MapFromPayload(payload, "group_vars.extra_variables"))...,
		),
		terraformVarFlags(MapFromPayload(payload, "group_vars.secret_extra_variables"))...,
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

	return &TerraformExec{
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

func terraformVarFlags(values map[string]interface{}) []string {
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

func RunTerraformLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, steps [][]string) error {
	for _, step := range steps {
		if err := RunTerraformLocalStep(ctx, deps, task, dir, step); err != nil {
			return err
		}
	}
	return nil
}

func RunTerraformDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, steps [][]string) error {
	for _, step := range steps {
		if err := RunTerraformDockerStep(ctx, deps, task, dir, step); err != nil {
			return err
		}
	}
	return nil
}

func RunTerraformLocalStep(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, step []string) error {
	command, err := BuildTerraformExec(dir, task.Payload)
	if err != nil {
		return err
	}

	args := append([]string{}, step...)
	args = append(args, command.Args...)

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = command.WorkDir
	cmd.Env = append(os.Environ(), command.Env...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("terraform: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("terraform: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("terraform: start local: %w", err)
	}

	done := make(chan struct{})
	go appendTerraformOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go appendTerraformOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("terraform: local %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func RunTerraformDockerStep(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, step []string) error {
	command, err := BuildTerraformExec(dir, task.Payload)
	if err != nil {
		return err
	}

	image, ok := terraformImages[command.Version]
	if !ok {
		return fmt.Errorf("terraform: unsupported version %q", command.Version)
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

func appendTerraformOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
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
