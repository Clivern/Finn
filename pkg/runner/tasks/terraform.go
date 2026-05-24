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

// IaCExec is the parsed IaC command for a cloned workspace.
type IaCExec struct {
	EngineVersion string
	WorkDir       string
	RootDir       string
	ChdirArg      []string
	Env           []string
}

// RunIaC runs an IaC template task.
func RunIaC(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	engine := iacEngine(task.Payload)
	switch engine {
	case "terraform", "":
		return RunTerraform(ctx, deps, task, dir)
	case "opentofu":
		return RunOpenTofu(ctx, deps, task, dir)
	default:
		return fmt.Errorf("iac: unsupported engine %q", engine)
	}
}

func RunTerraform(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildIaCExec(dir, task.Payload)
	if err != nil {
		return err
	}

	if IsDockerAvailable() {
		return runIaCDocker(ctx, deps, task, execCmd, "terraform", terraformDockerImage)
	}

	return runIaCLocal(ctx, deps, task, execCmd, "terraform")
}

func terraformDockerImage(version string) (string, error) {
	version = strings.TrimSpace(version)
	image, ok := terraformImages[version]
	if !ok {
		return "", fmt.Errorf("iac: unsupported terraform version %q", version)
	}
	return image, nil
}

func BuildIaCExec(dir string, payload map[string]interface{}) (*IaCExec, error) {
	workDir, err := iacWorkDir(dir, payload)
	if err != nil {
		return nil, err
	}

	env := append(
		keyValueArgs(MapFromPayload(payload, "groupVars.environmentVariables")),
		keyValueArgs(MapFromPayload(payload, "groupVars.secretEnvironmentVariables"))...,
	)
	if env == nil {
		env = append(
			keyValueArgs(MapFromPayload(payload, "group_vars.environment_variables")),
			keyValueArgs(MapFromPayload(payload, "group_vars.secret_environment_variables"))...,
		)
	}
	if lvl := iacLogLevel(payload); lvl != "" {
		env = append(env, "TF_LOG="+strings.ToUpper(lvl))
	}

	return &IaCExec{
		EngineVersion: iacEngineVersion(payload),
		WorkDir:       workDir,
		RootDir:       dir,
		ChdirArg:      iacChdirArg(dir, workDir),
		Env:           env,
	}, nil
}

func iacEngine(payload map[string]interface{}) string {
	engine := StringFromPayload(payload, "iac.engine")
	if engine == "" {
		engine = StringFromPayload(payload, "iac_config.engine")
	}
	if engine == "" {
		return "terraform"
	}
	return engine
}

func iacEngineVersion(payload map[string]interface{}) string {
	version := StringFromPayload(payload, "iac.engineVersion")
	if version == "" {
		version = StringFromPayload(payload, "iac_config.engineVersion")
	}
	if version == "" {
		version = StringFromPayload(payload, "iac_config.engine_version")
	}
	return version
}

func iacLogLevel(payload map[string]interface{}) string {
	level := StringFromPayload(payload, "iac.logLevel")
	if level == "" {
		level = StringFromPayload(payload, "iac_config.logLevel")
	}
	if level == "" {
		level = StringFromPayload(payload, "iac_config.log_level")
	}
	return level
}

func iacWorkDir(root string, payload map[string]interface{}) (string, error) {
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
		return "", fmt.Errorf("iac: resolve working directory: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("iac: resolve workspace directory: %w", err)
	}
	if absWorkDir != absRoot && !strings.HasPrefix(absWorkDir, absRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("iac: working directory escapes workspace")
	}
	return absWorkDir, nil
}

func iacChdirArg(root, workDir string) []string {
	if workDir == root {
		return nil
	}

	rel, err := filepath.Rel(root, workDir)
	if err != nil || rel == "." {
		return nil
	}
	return []string{"-chdir=" + filepath.ToSlash(rel)}
}

func iacSteps(execCmd *IaCExec) [][]string {
	return [][]string{
		{"init", "-input=false"},
		{"plan", "-input=false"},
		{"apply", "-input=false", "-auto-approve"},
	}
}

func runIaCLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, execCmd *IaCExec, binary string) error {
	for _, step := range iacSteps(execCmd) {
		args := append(append([]string{}, execCmd.ChdirArg...), step...)
		if err := runIaCStepLocal(ctx, deps, task.ID, execCmd.RootDir, binary, args, execCmd.Env); err != nil {
			return err
		}
	}
	return nil
}

func runIaCDocker(
	ctx context.Context,
	deps *runner.Deps,
	task *runner.Task,
	execCmd *IaCExec,
	binary string,
	imageFn func(string) (string, error),
) error {
	image, err := imageFn(execCmd.EngineVersion)
	if err != nil {
		return err
	}

	for _, step := range iacSteps(execCmd) {
		args := append(append([]string{binary}, execCmd.ChdirArg...), step...)
		if err := runIaCStepDocker(ctx, deps, task.ID, execCmd.RootDir, image, args, execCmd.Env); err != nil {
			return err
		}
	}
	return nil
}

func runIaCStepLocal(ctx context.Context, deps *runner.Deps, taskID, dir, binary string, args, env []string) error {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("iac: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("iac: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("iac: start local: %w", err)
	}

	done := make(chan struct{})
	go appendIaCOutput(stdoutPipe, "stdout", deps, taskID, done)
	go appendIaCOutput(stderrPipe, "stderr", deps, taskID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("iac: local %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func runIaCStepDocker(ctx context.Context, deps *runner.Deps, taskID, dir, image string, command, env []string) error {
	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     dir,
		Image:   image,
		Command: command,
		Env:     envListToMap(env),
		Timeout: remaining(ctx, defaultTemplateTimeout),
		OnOutput: func(stream, line string) {
			_ = deps.Worker.AppendTaskLog(taskID, map[string]string{
				"stream": stream,
				"line":   line,
				"msg":    "template",
			})
		},
	})
}

func appendIaCOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
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
