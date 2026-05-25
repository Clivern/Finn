// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

var (
	ansibleImages = map[string]string{
		"":       "quay.io/ansible/creator-ee:latest",
		"latest": "quay.io/ansible/creator-ee:latest",
	}
)

// AnsibleExec is the parsed Ansible command for a cloned workspace.
type AnsibleExec struct {
	EngineVersion string
	PlaybookPath  string
	InventoryPath string
	Env           []string
	Args          []string
}

// RunAnsible runs an Ansible template task.
func RunAnsible(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildAnsibleExec(dir, task.Payload)
	if err != nil {
		return err
	}

	if IsDockerAvailable() {
		return RunAnsibleDocker(ctx, deps, task, dir, execCmd)
	}

	return RunAnsibleLocal(ctx, deps, task, dir, execCmd)
}

func RunAnsibleLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, execCmd *AnsibleExec) error {
	args := append([]string{}, execCmd.Args...)
	args = append(args, execCmd.PlaybookPath)

	cmd := exec.CommandContext(ctx, "ansible-playbook", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), execCmd.Env...)

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
	go appendIaCOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go appendIaCOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ansible: local: %w", err)
	}
	return nil
}

func RunAnsibleDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string, execCmd *AnsibleExec) error {
	image, err := ansibleDockerImage(execCmd.EngineVersion)
	if err != nil {
		return err
	}

	command := append([]string{"ansible-playbook"}, execCmd.Args...)
	command = append(command, execCmd.PlaybookPath)

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     dir,
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
	playbookPath := StringFromPayload(payload, "ansible.playbookPath")
	if playbookPath == "" {
		playbookPath = StringFromPayload(payload, "ansible_config.playbook_path")
	}
	if playbookPath == "" {
		return nil, fmt.Errorf("ansible: playbook path is required")
	}

	env := append(
		keyValueArgs(MapFromPayload(payload, "groupVars.environmentVariables")),
		keyValueArgs(MapFromPayload(payload, "groupVars.secretEnvironmentVariables"))...,
	)
	args := append(
		append(
			ansibleExtraVarArgs(MapFromPayload(payload, "extraArgs")),
			ansibleExtraVarArgs(MapFromPayload(payload, "groupVars.extraVariables"))...,
		),
		ansibleExtraVarArgs(MapFromPayload(payload, "groupVars.secretExtraVariables"))...,
	)

	inventoryPath, err := ansibleInventoryPath(dir, payload)
	if err != nil {
		return nil, err
	}
	if inventoryPath != "" {
		args = append(args, "-i", inventoryPath)
	}

	if vaultPath := ansibleVaultPath(payload); vaultPath != "" {
		args = append(args, "--extra-vars", "@"+vaultPath)
	}
	vaultPasswordPath, err := ansibleVaultPasswordPath(dir, payload)
	if err != nil {
		return nil, err
	}
	if vaultPasswordPath != "" {
		args = append(args, "--vault-password-file", vaultPasswordPath)
	}
	args = append(args, ansibleVerbosityArgs(payload)...)

	return &AnsibleExec{
		EngineVersion: ansibleEngineVersion(payload),
		PlaybookPath:  workspaceRelPath(playbookPath),
		InventoryPath: inventoryPath,
		Env:           env,
		Args:          args,
	}, nil
}

func ansibleDockerImage(version string) (string, error) {
	version = strings.TrimSpace(version)
	image, ok := ansibleImages[version]
	if !ok {
		return "", fmt.Errorf("ansible: unsupported version %q", version)
	}
	return image, nil
}

func ansibleEngineVersion(payload map[string]interface{}) string {
	version := StringFromPayload(payload, "ansible.engineVersion")
	if version == "" {
		version = StringFromPayload(payload, "ansible_config.engine_version")
	}
	return version
}

func ansibleInventoryPath(dir string, payload map[string]interface{}) (string, error) {
	if inventory := StringFromPayload(payload, "ansible_config.inventory.inventory"); inventory != "" {
		relPath := ".finn/ansible/inventory.ini"
		if err := writeWorkspaceFile(dir, relPath, inventory); err != nil {
			return "", err
		}
		return relPath, nil
	}

	inventoryPath := StringFromPayload(payload, "ansible_config.inventory.inventory_path")
	if inventoryPath == "" {
		return "", nil
	}
	return workspaceRelPath(inventoryPath), nil
}

func ansibleVaultPath(payload map[string]interface{}) string {
	vaultPath := StringFromPayload(payload, "ansible.vaultPath")
	if vaultPath == "" {
		vaultPath = StringFromPayload(payload, "ansible_config.vault_path")
	}
	if vaultPath == "" {
		return ""
	}
	return workspaceRelPath(vaultPath)
}

func ansibleVaultPasswordPath(dir string, payload map[string]interface{}) (string, error) {
	password := StringFromPayload(payload, "ansible.vaultPassword")
	if password == "" {
		password = StringFromPayload(payload, "ansible_config.vault_password")
	}
	if password == "" {
		return "", nil
	}

	relPath := ".finn/ansible/vault-password"
	if err := writeWorkspaceFile(dir, relPath, password+"\n"); err != nil {
		return "", err
	}
	return relPath, nil
}

func ansibleVerbosityArgs(payload map[string]interface{}) []string {
	level := StringFromPayload(payload, "ansible.logLevel")
	if level == "" {
		level = StringFromPayload(payload, "ansible_config.log_level")
	}
	if level == "" || level == "0" {
		return nil
	}

	n, err := strconv.Atoi(level)
	if err != nil || n <= 0 {
		return nil
	}
	if n > 5 {
		n = 5
	}
	return []string{"-" + strings.Repeat("v", n)}
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

func writeWorkspaceFile(dir, relPath, contents string) error {
	relPath = workspaceRelPath(relPath)
	path := filepath.Join(dir, filepath.FromSlash(relPath))
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("ansible: resolve generated file: %w", err)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("ansible: resolve workspace directory: %w", err)
	}
	if absPath != absDir && !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		return fmt.Errorf("ansible: generated file escapes workspace")
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o700); err != nil {
		return fmt.Errorf("ansible: create generated file directory: %w", err)
	}
	if err := os.WriteFile(absPath, []byte(contents), 0o600); err != nil {
		return fmt.Errorf("ansible: write generated file: %w", err)
	}
	return nil
}

func workspaceRelPath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(path)), "/")
}
