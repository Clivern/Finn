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
		"":     "ghcr.io/opentofu/opentofu:1.9",
		"1.8":  "ghcr.io/opentofu/opentofu:1.8",
		"1.9":  "ghcr.io/opentofu/opentofu:1.9",
		"1.10": "ghcr.io/opentofu/opentofu:1.10",
	}
)

// OpenTofuExec is the parsed OpenTofu command for a cloned workspace.
type OpenTofuExec struct {
	Version  string
	WorkDir  string
	RootDir  string
	ChdirArg []string
	Env      []string
	Args     []string
}

// RunOpenTofu runs OpenTofu for a cloned workspace.
func RunOpenTofu(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildOpenTofuExec(dir, task.Payload)
	if err != nil {
		return err
	}

	if IsDockerAvailable() {
		return runOpenTofuDocker(ctx, deps, task, execCmd)
	}

	return runOpenTofuLocal(ctx, deps, task, execCmd)
}

func opentofuDockerImage(version string) (string, error) {
	version = strings.TrimSpace(version)
	image, ok := opentofuImages[version]
	if !ok {
		return "", fmt.Errorf("opentofu: unsupported version %q", version)
	}
	return image, nil
}

// BuildOpenTofuExec builds an OpenTofu command from the task payload.
func BuildOpenTofuExec(dir string, payload map[string]interface{}) (*OpenTofuExec, error) {
	workDir, err := opentofuWorkDir(dir, payload)
	if err != nil {
		return nil, err
	}

	env := append(
		keyValueArgs(MapFromPayload(payload, "groupVars.environmentVariables")),
		keyValueArgs(MapFromPayload(payload, "groupVars.secretEnvironmentVariables"))...,
	)
	if logLevel := StringFromPayload(payload, "iac.logLevel"); logLevel != "" {
		env = append(env, "TF_LOG="+strings.ToUpper(logLevel))
	}
	args := append(
		append(
			opentofuVarFlags(MapFromPayload(payload, "extraArgs")),
			opentofuVarFlags(MapFromPayload(payload, "groupVars.extraVariables"))...,
		),
		opentofuVarFlags(MapFromPayload(payload, "groupVars.secretExtraVariables"))...,
	)

	return &OpenTofuExec{
		Version:  StringFromPayload(payload, "iac.engineVersion"),
		WorkDir:  workDir,
		RootDir:  dir,
		ChdirArg: opentofuChdirArg(dir, workDir),
		Env:      env,
		Args:     args,
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

func opentofuWorkDir(root string, payload map[string]interface{}) (string, error) {
	rel := StringFromPayload(payload, "workingDirectory")
	if rel == "" {
		rel = StringFromPayload(payload, "working_dir")
	}
	rel = strings.TrimSpace(rel)
	if rel == "" || rel == "/" {
		return root, nil
	}

	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	workDir := filepath.Join(root, filepath.FromSlash(rel))
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("opentofu: resolve working directory: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("opentofu: resolve workspace directory: %w", err)
	}
	if absWorkDir != absRoot && !strings.HasPrefix(absWorkDir, absRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("opentofu: working directory escapes workspace")
	}
	return absWorkDir, nil
}

func opentofuChdirArg(root, workDir string) []string {
	if workDir == root {
		return nil
	}

	rel, err := filepath.Rel(root, workDir)
	if err != nil || rel == "." {
		return nil
	}
	return []string{"-chdir=" + filepath.ToSlash(rel)}
}

func opentofuSteps() [][]string {
	return [][]string{
		{"init", "-input=false"},
		{"plan", "-input=false"},
		{"apply", "-input=false", "-auto-approve"},
	}
}

func runOpenTofuLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, execCmd *OpenTofuExec) error {
	for _, step := range opentofuSteps() {
		args := append(append([]string{}, execCmd.ChdirArg...), step...)
		args = append(args, execCmd.Args...)
		if err := runOpenTofuStepLocal(ctx, deps, task.ID, execCmd.RootDir, args, execCmd.Env); err != nil {
			return err
		}
	}
	return nil
}

func runOpenTofuDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, execCmd *OpenTofuExec) error {
	image, err := opentofuDockerImage(execCmd.Version)
	if err != nil {
		return err
	}

	for _, step := range opentofuSteps() {
		args := append(append([]string{"tofu"}, execCmd.ChdirArg...), step...)
		args = append(args, execCmd.Args...)
		if err := runOpenTofuStepDocker(ctx, deps, task, execCmd.RootDir, image, args, execCmd.Env); err != nil {
			return err
		}
	}
	return nil
}

func runOpenTofuStepLocal(ctx context.Context, deps *runner.Deps, taskID, dir string, args, env []string) error {
	cmd := exec.CommandContext(ctx, "tofu", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)

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
	go appendOpenTofuOutput(stdoutPipe, "stdout", deps, taskID, done)
	go appendOpenTofuOutput(stderrPipe, "stderr", deps, taskID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("opentofu: local %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func runOpenTofuStepDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir, image string, command, env []string) error {
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

func appendOpenTofuOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
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
