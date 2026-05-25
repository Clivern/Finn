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
		"":     "docker.io/hashicorp/terraform:1.11",
		"1.9":  "docker.io/hashicorp/terraform:1.9",
		"1.10": "docker.io/hashicorp/terraform:1.10",
		"1.11": "docker.io/hashicorp/terraform:1.11",
	}
)

// TerraformExec is the parsed Terraform command for a cloned workspace.
type TerraformExec struct {
	Version  string
	WorkDir  string
	RootDir  string
	ChdirArg []string
	Env      []string
	Args     []string
}

// RunTerraform runs Terraform for a cloned workspace.
func RunTerraform(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	if IsDockerAvailable() {
		return RunTerraformDocker(ctx, deps, task, dir)
	}

	return RunTerraformLocal(ctx, deps, task, dir)
}

func terraformDockerImage(version string) (string, error) {
	version = strings.TrimSpace(version)
	image, ok := terraformImages[version]
	if !ok {
		return "", fmt.Errorf("terraform: unsupported version %q", version)
	}
	return image, nil
}

func BuildTerraformExec(dir string, payload map[string]interface{}) (*TerraformExec, error) {
	workDir, err := terraformWorkDir(dir, payload)
	if err != nil {
		return nil, err
	}

	env := append(
		keyValueArgs(MapFromPayload(payload, "group_vars.environment_variables")),
		keyValueArgs(MapFromPayload(payload, "group_vars.secret_environment_variables"))...,
	)
	if logLevel := StringFromPayload(payload, "iac.logLevel"); logLevel != "" {
		env = append(env, "TF_LOG="+strings.ToUpper(logLevel))
	}
	args := append(
		append(
			terraformVarFlags(MapFromPayload(payload, "extra_args")),
			terraformVarFlags(MapFromPayload(payload, "group_vars.extra_variables"))...,
		),
		terraformVarFlags(MapFromPayload(payload, "group_vars.secret_extra_variables"))...,
	)

	return &TerraformExec{
		Version:  StringFromPayload(payload, "iac.engineVersion"),
		WorkDir:  workDir,
		RootDir:  dir,
		ChdirArg: terraformChdirArg(dir, workDir),
		Env:      env,
		Args:     args,
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

func terraformWorkDir(root string, payload map[string]interface{}) (string, error) {
	rel := StringFromPayload(payload, "working_dir")
	rel = strings.TrimSpace(rel)
	if rel == "" || rel == "/" {
		return root, nil
	}

	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	workDir := filepath.Join(root, filepath.FromSlash(rel))
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("terraform: resolve working directory: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("terraform: resolve workspace directory: %w", err)
	}
	if absWorkDir != absRoot && !strings.HasPrefix(absWorkDir, absRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("terraform: working directory escapes workspace")
	}
	return absWorkDir, nil
}

func terraformChdirArg(root, workDir string) []string {
	if workDir == root {
		return nil
	}

	rel, err := filepath.Rel(root, workDir)
	if err != nil || rel == "." {
		return nil
	}
	return []string{"-chdir=" + filepath.ToSlash(rel)}
}

func terraformSteps() [][]string {
	return [][]string{
		{"init", "-input=false"},
		{"plan", "-input=false"},
		{"apply", "-input=false", "-auto-approve"},
	}
}

func RunTerraformLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildTerraformExec(dir, task.Payload)
	if err != nil {
		return err
	}

	for _, step := range terraformSteps() {
		args := append(append([]string{}, execCmd.ChdirArg...), step...)
		args = append(args, execCmd.Args...)
		if err := runTerraformStepLocal(ctx, deps, task.ID, execCmd.RootDir, args, execCmd.Env); err != nil {
			return err
		}
	}
	return nil
}

func RunTerraformDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildTerraformExec(dir, task.Payload)
	if err != nil {
		return err
	}

	image, err := terraformDockerImage(execCmd.Version)
	if err != nil {
		return err
	}

	for _, step := range terraformSteps() {
		args := append(append([]string{"terraform"}, execCmd.ChdirArg...), step...)
		args = append(args, execCmd.Args...)
		if err := runTerraformStepDocker(ctx, deps, task, execCmd.RootDir, image, args, execCmd.Env); err != nil {
			return err
		}
	}
	return nil
}

func runTerraformStepLocal(ctx context.Context, deps *runner.Deps, taskID, dir string, args, env []string) error {
	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)

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
	go appendTerraformOutput(stdoutPipe, "stdout", deps, taskID, done)
	go appendTerraformOutput(stderrPipe, "stderr", deps, taskID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("terraform: local %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func runTerraformStepDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir, image string, command, env []string) error {
	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     dir,
		Image:   image,
		Command: command,
		Env:     envListToMap(env),
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
