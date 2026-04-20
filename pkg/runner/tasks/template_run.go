// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

const defaultTemplateTimeout = time.Hour

// WorkspaceTemplate clones the repository then runs script, IaC, or Ansible per payload.
func WorkspaceTemplate(deps *runner.Deps, task *runner.Task) error {
	p := task.Payload
	timeoutSec := IntFromPayload(p, "timeout")
	if timeoutSec <= 0 {
		timeoutSec = int(defaultTemplateTimeout / time.Second)
	}
	runCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	dir, err := os.MkdirTemp("", task.ID)
	if err != nil {
		return fmt.Errorf("template: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	commit, err := workspace.Clone(runCtx, &workspace.CloneOptions{
		URL:           StringFromPayload(p, "repository"),
		Branch:        StringFromPayload(p, "branch"),
		Tag:           StringFromPayload(p, "tag"),
		Dir:           dir,
		Timeout:       time.Duration(timeoutSec) * time.Second,
		Username:      StringFromPayload(p, "username"),
		Password:      StringFromPayload(p, "password"),
		SSHPrivateKey: StringFromPayload(p, "ssh_private_key"),
		SSHPassphrase: StringFromPayload(p, "ssh_passphrase"),
		OnProgress: func(msg string) {
			_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
				"stream": "stdout",
				"line":   msg,
				"msg":    "clone",
			})
		},
	})
	if err != nil {
		return err
	}

	_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "HEAD is at " + commit,
		"msg":    "clone",
	})

	workRoot := filepath.Join(dir, strings.TrimSpace(StringFromPayload(p, "working_directory")))
	st, err := os.Stat(workRoot)
	if err != nil || !st.IsDir() {
		return fmt.Errorf("template: working directory not found or not a directory")
	}

	engine := strings.TrimSpace(StringFromPayload(p, "engine"))
	switch engine {
	case "script":
		return runScriptEngine(runCtx, deps, task, workRoot, p)
	case "iac":
		return runIacEngine(runCtx, deps, task, workRoot, p)
	case "ansible":
		return runAnsibleEngine(runCtx, deps, task, workRoot, p)
	default:
		return fmt.Errorf("template: unknown engine %q", engine)
	}
}

func appendTemplateLog(deps *runner.Deps, taskID, stream, line string) {
	_ = deps.Worker.AppendTaskLog(taskID, map[string]string{
		"stream": stream,
		"line":   line,
		"msg":    "template",
	})
}

func runScriptEngine(ctx context.Context, deps *runner.Deps, task *runner.Task, workRoot string, p map[string]interface{}) error {
	sm := MapFromPayload(p, "script")
	if len(sm) == 0 {
		return fmt.Errorf("template: missing script configuration")
	}
	engine := strings.ToLower(StringFromMap(sm, "engine"))
	if engine == "" {
		engine = "bash"
	}
	scriptRel := strings.TrimSpace(StringFromMap(sm, "scriptPath"))
	if scriptRel == "" {
		return fmt.Errorf("template: script path is required")
	}

	hostScript := filepath.Join(workRoot, filepath.FromSlash(scriptRel))
	hostScript = filepath.Clean(hostScript)
	baseClean := filepath.Clean(workRoot)
	if rel, err := filepath.Rel(baseClean, hostScript); err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("template: invalid script path")
	}
	if _, err := os.Stat(hostScript); err != nil {
		return fmt.Errorf("template: script file: %w", err)
	}

	image := scriptDockerImage(engine, StringFromMap(sm, "engineVersion"))
	dockerScript := "/workspace/" + filepath.ToSlash(strings.TrimPrefix(
		strings.TrimPrefix(hostScript, baseClean), string(filepath.Separator)))

	var cmd []string
	switch engine {
	case "python3":
		cmd = []string{"python3", dockerScript}
	case "node":
		cmd = []string{"node", dockerScript}
	case "sh":
		cmd = []string{"sh", dockerScript}
	default:
		cmd = []string{"bash", dockerScript}
	}

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     workRoot,
		Image:   image,
		Command: cmd,
		Env:     EnvFromPayload(p),
		Timeout: remaining(ctx, defaultTemplateTimeout),
		OnOutput: func(stream, line string) {
			appendTemplateLog(deps, task.ID, stream, line)
		},
	})
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

func runIacEngine(ctx context.Context, deps *runner.Deps, task *runner.Task, workRoot string, p map[string]interface{}) error {
	im := MapFromPayload(p, "iac")
	if len(im) == 0 {
		return fmt.Errorf("template: missing IaC configuration")
	}
	engine := strings.ToLower(StringFromMap(im, "engine"))
	if engine == "" {
		engine = "terraform"
	}
	binary := "terraform"
	image := "docker.io/hashicorp/terraform:1.9"
	if engine == "opentofu" {
		binary = "tofu"
		image = "ghcr.io/opentofu/opentofu:1.8"
	}
	autoApprove := BoolFromMap(im, "autoApprove")
	extra := strings.TrimSpace(StringFromPayload(p, "extra_args"))
	if extra == "" {
		extra = strings.TrimSpace(StringFromPayload(p, "extraArgs"))
	}

	var shellCmd string
	if autoApprove {
		shellCmd = fmt.Sprintf("%s init -input=false && %s apply -auto-approve", binary, binary)
	} else {
		shellCmd = fmt.Sprintf("%s init -input=false && %s plan", binary, binary)
	}
	if extra != "" {
		shellCmd += " " + extra
	}

	env := EnvFromPayload(p)
	if lvl := strings.TrimSpace(StringFromMap(im, "logLevel")); lvl != "" {
		env["TF_LOG"] = lvl
	}

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     workRoot,
		Image:   image,
		Command: []string{"sh", "-c", shellCmd},
		Env:     env,
		Timeout: remaining(ctx, defaultTemplateTimeout),
		OnOutput: func(stream, line string) {
			appendTemplateLog(deps, task.ID, stream, line)
		},
	})
}

func runAnsibleEngine(ctx context.Context, deps *runner.Deps, task *runner.Task, workRoot string, p map[string]interface{}) error {
	am := MapFromPayload(p, "ansible")
	if len(am) == 0 {
		return fmt.Errorf("template: missing Ansible configuration")
	}
	playRel := strings.TrimSpace(StringFromMap(am, "playbookPath"))
	if playRel == "" {
		return fmt.Errorf("template: ansible playbook path is required")
	}

	inlineInv := strings.TrimSpace(StringFromPayload(p, "ansible_inventory_body"))
	repoInv := strings.TrimSpace(StringFromPayload(p, "ansible_inventory_repo_path"))

	var invDockerPath string
	switch {
	case inlineInv != "":
		hostInv := filepath.Join(workRoot, ".finn_inventory")
		if err := os.WriteFile(hostInv, []byte(inlineInv), 0o600); err != nil {
			return fmt.Errorf("template: write inventory: %w", err)
		}
		invDockerPath = "/workspace/.finn_inventory"
	case repoInv != "":
		hostInv := filepath.Join(workRoot, filepath.FromSlash(repoInv))
		hostInv = filepath.Clean(hostInv)
		baseClean := filepath.Clean(workRoot)
		if rel, err := filepath.Rel(baseClean, hostInv); err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("template: invalid ansible inventory path")
		}
		if _, err := os.Stat(hostInv); err != nil {
			return fmt.Errorf("template: inventory file: %w", err)
		}
		invDockerPath = "/workspace/" + filepath.ToSlash(strings.TrimPrefix(
			strings.TrimPrefix(hostInv, baseClean), string(filepath.Separator)))
	default:
		return fmt.Errorf("template: ansible inventory is required (configure inventory in Finn or add a file in the repo)")
	}

	hostPB := filepath.Join(workRoot, filepath.FromSlash(playRel))
	hostPB = filepath.Clean(hostPB)
	baseClean := filepath.Clean(workRoot)
	if rel, err := filepath.Rel(baseClean, hostPB); err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("template: invalid playbook path")
	}
	if _, err := os.Stat(hostPB); err != nil {
		return fmt.Errorf("template: playbook file: %w", err)
	}
	playDocker := "/workspace/" + filepath.ToSlash(strings.TrimPrefix(
		strings.TrimPrefix(hostPB, baseClean), string(filepath.Separator)))

	image := "docker.io/cytopia/ansible:latest"
	args := []string{"ansible-playbook", "-i", invDockerPath}
	if vv := ansibleVerbosity(StringFromMap(am, "logLevel")); vv != "" {
		args = append(args, vv)
	}
	args = append(args, playDocker)

	env := EnvFromPayload(p)
	if vault := strings.TrimSpace(StringFromPayload(p, "ansible_vault_password")); vault != "" {
		if env == nil {
			env = make(map[string]string)
		}
		env["ANSIBLE_VAULT_PASSWORD"] = vault
	}

	return workspace.Exec(ctx, &workspace.ExecOptions{
		Dir:     workRoot,
		Image:   image,
		Command: args,
		Env:     env,
		Timeout: remaining(ctx, defaultTemplateTimeout),
		OnOutput: func(stream, line string) {
			appendTemplateLog(deps, task.ID, stream, line)
		},
	})
}

func ansibleVerbosity(logLevel string) string {
	logLevel = strings.TrimSpace(logLevel)
	if logLevel == "" {
		return ""
	}
	n, err := strconv.Atoi(logLevel)
	if err != nil || n < 0 || n > 4 {
		return ""
	}
	return "-" + strings.Repeat("v", n)
}

func remaining(ctx context.Context, fallback time.Duration) time.Duration {
	d, ok := ctx.Deadline()
	if !ok {
		return fallback
	}
	left := time.Until(d)
	if left <= 0 {
		return time.Second
	}
	return left
}
