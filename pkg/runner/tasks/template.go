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

type templateCloneFields struct {
	repoURL       string
	branch        string
	tag           string
	username      string
	password      string
	sshPrivateKey string
	sshPassphrase string
	workingSubdir string
}

func templateFirstNonEmpty(vals ...string) string {
	for _, s := range vals {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// cloneFieldsFromTemplatePayload supports the flat clone payload (repository string, ssh_private_key, …)
// and the module API payload where repository is an object (git_url, ssh_key.private_key, ssh_key.passphrase, …).
func cloneFieldsFromTemplatePayload(p map[string]interface{}) templateCloneFields {
	var out templateCloneFields
	repo := MapFromPayload(p, "repository")
	if u := strings.TrimSpace(StringFromPayload(p, "repository")); u != "" {
		out.repoURL = u
	} else {
		out.repoURL = StringFromMap(repo, "git_url")
	}
	out.branch = templateFirstNonEmpty(
		StringFromPayload(p, "branch"),
		StringFromPayload(p, "git_branch_override"),
		StringFromMap(repo, "git_branch"),
	)
	out.tag = templateFirstNonEmpty(
		StringFromPayload(p, "tag"),
		StringFromPayload(p, "git_tag_override"),
		StringFromMap(repo, "git_tag"),
	)
	out.username = StringFromPayload(p, "username")
	out.password = StringFromPayload(p, "password")
	if c := MapFromPayload(repo, "git_credential"); c != nil {
		if out.username == "" {
			out.username = StringFromMap(c, "username")
		}
		if out.password == "" {
			out.password = StringFromMap(c, "password_or_token")
		}
	}
	out.sshPrivateKey = StringFromPayload(p, "ssh_private_key")
	out.sshPassphrase = StringFromPayload(p, "ssh_passphrase")
	sshKey := MapFromPayload(p, "ssh_key")
	if len(sshKey) == 0 {
		sshKey = MapFromPayload(repo, "ssh_key")
	}
	if len(sshKey) > 0 {
		if out.sshPrivateKey == "" {
			out.sshPrivateKey = StringFromMap(sshKey, "private_key")
		}
		if out.sshPassphrase == "" {
			out.sshPassphrase = StringFromMap(sshKey, "passphrase")
		}
	}
	out.workingSubdir = templateFirstNonEmpty(
		StringFromPayload(p, "working_directory"),
		StringFromPayload(p, "working_dir"),
	)
	return out
}

// WorkspaceTemplate clones the repository then runs script, IaC, or Ansible per payload.
func WorkspaceTemplate(ctx context.Context, deps *runner.Deps, task *runner.Task) error {
	p := task.Payload

	dir, err := os.MkdirTemp("", task.ID)
	if err != nil {
		return fmt.Errorf("template: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	clone := cloneFieldsFromTemplatePayload(p)
	commit, err := workspace.Clone(ctx, &workspace.CloneOptions{
		URL:           clone.repoURL,
		Branch:        clone.branch,
		Tag:           clone.tag,
		Dir:           dir,
		Timeout:       remaining(ctx, defaultTemplateTimeout),
		Username:      clone.username,
		Password:      clone.password,
		SSHPrivateKey: clone.sshPrivateKey,
		SSHPassphrase: clone.sshPassphrase,
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

	workRoot := filepath.Join(dir, strings.TrimSpace(clone.workingSubdir))
	st, err := os.Stat(workRoot)
	if err != nil || !st.IsDir() {
		return fmt.Errorf("template: working directory not found or not a directory")
	}

	engine := strings.TrimSpace(StringFromPayload(p, "engine"))
	switch engine {
	case "script":
		return runScriptEngine(ctx, deps, task, workRoot, p)
	case "iac":
		return runIacEngine(ctx, deps, task, workRoot, p)
	case "ansible":
		return runAnsibleEngine(ctx, deps, task, workRoot, p)
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
