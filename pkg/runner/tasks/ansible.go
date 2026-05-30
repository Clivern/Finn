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
	ansibleImages = map[string]string{
		"":        "quay.io/ansible/creator-ee:latest",
		"latest":  "quay.io/ansible/creator-ee:latest",
		"24.2":    "quay.io/ansible/creator-ee:v24.2.0",
		"24.2.0":  "quay.io/ansible/creator-ee:v24.2.0",
		"v24.2.0": "quay.io/ansible/creator-ee:v24.2.0",
	}
)

// AnsibleExec is the parsed Ansible command for a cloned workspace.
type AnsibleExec struct {
	Version       string
	PlaybookPath  string
	WorkDir       string
	WorkRelPath   string
	RootDir       string
	InventoryPath string
	Env           []string
	Args          []string
}

// RunAnsible runs an Ansible template task.
func RunAnsible(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	if IsDockerAvailable() {
		return RunAnsibleDocker(ctx, deps, task, dir)
	}

	return RunAnsibleLocal(ctx, deps, task, dir)
}

func RunAnsibleLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	command, err := BuildAnsibleExec(dir, task.Payload)
	if err != nil {
		return err
	}

	args := append([]string{}, command.Args...)
	args = append(args, command.PlaybookPath)

	cmd := exec.CommandContext(ctx, "ansible-playbook", args...)
	cmd.Dir = command.WorkDir
	cmd.Env = append(os.Environ(), command.Env...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ansible: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ansible: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ansible: start local: %w", err)
	}

	done := make(chan struct{})
	go appendAnsibleOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go appendAnsibleOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ansible: local: %w", err)
	}
	return nil
}

func RunAnsibleDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildAnsibleExec(dir, task.Payload)
	if err != nil {
		return err
	}

	version := strings.TrimSpace(execCmd.Version)
	image, ok := ansibleImages[version]
	if !ok {
		return fmt.Errorf("ansible: unsupported version %q", version)
	}

	command := append([]string{"ansible-playbook"}, execCmd.Args...)
	command = append(command, execCmd.PlaybookPath)

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     execCmd.RootDir,
		RelPath: execCmd.WorkRelPath,
		Image:   image,
		Command: command,
		Env:     envListToMap(execCmd.Env),
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

func BuildAnsibleExec(dir string, payload map[string]interface{}) (*AnsibleExec, error) {
	env := append(
		keyValueArgs(MapFromPayload(payload, "group_vars.environment_variables")),
		keyValueArgs(MapFromPayload(payload, "group_vars.secret_environment_variables"))...,
	)
	args := append(
		append(
			ansibleExtraVarArgs(MapFromPayload(payload, "extra_args")),
			ansibleExtraVarArgs(MapFromPayload(payload, "group_vars.extra_variables"))...,
		),
		ansibleExtraVarArgs(MapFromPayload(payload, "group_vars.secret_extra_variables"))...,
	)

	args = append(args, "-i", workspaceRelPath(StringFromPayload(payload, "ansible.inventory.inventory_path")))
	args = append(args, "-"+strings.Repeat("v", IntFromPayload(payload, "ansible.log_level")))

	return &AnsibleExec{
		Version:       StringFromPayload(payload, "ansible.version"),
		PlaybookPath:  workspaceRelPath(StringFromPayload(payload, "ansible.playbook_path")),
		WorkDir:       filepath.Join(dir, filepath.FromSlash(StringFromPayload(payload, "working_dir"))),
		WorkRelPath:   strings.TrimPrefix(filepath.ToSlash(StringFromPayload(payload, "working_dir")), "/"),
		RootDir:       dir,
		InventoryPath: workspaceRelPath(StringFromPayload(payload, "ansible.inventory.inventory_path")),
		Env:           env,
		Args:          args,
	}, nil
}

func ansibleExtraVarArgs(values map[string]interface{}) []string {
	pairs := keyValueArgs(values)
	if len(pairs) == 0 {
		return nil
	}

	args := make([]string, 0, len(pairs)*2)
	for _, pair := range pairs {
		args = append(args, "--extra-vars", pair)
	}
	return args
}

func appendAnsibleOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
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

func workspaceRelPath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(path)), "/")
}
