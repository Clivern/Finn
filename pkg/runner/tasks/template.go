// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/workspace"
)

// WorkspaceTemplate clones the repository then runs the selected engine.
func WorkspaceTemplate(ctx context.Context, deps *runner.Deps, task *runner.Task) error {
	p := task.Payload

	dir, err := os.MkdirTemp("", task.ID)
	if err != nil {
		return fmt.Errorf("template: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	branch := StringFromPayload(p, "repository.git_branch")
	tag := StringFromPayload(p, "repository.git_tag")
	branchOverride := StringFromPayload(p, "branch_override")
	tagOverride := StringFromPayload(p, "tag_override")

	if branchOverride != "" {
		branch = branchOverride
	}
	if tagOverride != "" {
		tag = tagOverride
	}

	commit, err := workspace.Clone(ctx, &workspace.CloneOptions{
		URL:           StringFromPayload(p, "repository.git_url"),
		Branch:        branch,
		Tag:           tag,
		Dir:           dir,
		Timeout:       time.Duration(IntFromPayload(task.Payload, "timeout")) * time.Second,
		Username:      StringFromPayload(p, "repository.git_credential.username"),
		Password:      StringFromPayload(p, "repository.git_credential.password_or_token"),
		SSHPrivateKey: StringFromPayload(p, "repository.ssh_key.private_key"),
		SSHPassphrase: StringFromPayload(p, "repository.ssh_key.passphrase"),
		OnProgress: func(msg string) {
			deps.Worker.AppendTaskLog(task.ID, map[string]string{
				"stream": "stdout",
				"line":   msg,
				"msg":    "clone",
			})
		},
	})
	if err != nil {
		return err
	}

	deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "HEAD is at " + commit,
		"msg":    "clone",
	})

	engine := strings.TrimSpace(StringFromPayload(p, "engine"))
	switch engine {
	case "script":
		return RunScript(ctx, deps, task, dir)
	case "iac":
		switch StringFromPayload(p, "iac.bin") {
		case "terraform":
			return RunTerraform(ctx, deps, task, dir, ArrayStringFromPayload(p, "iac.steps"))
		case "opentofu":
			return RunOpentofu(ctx, deps, task, dir, ArrayStringFromPayload(p, "iac.steps"))
		default:
			return fmt.Errorf("iac: unsupported bin %q", StringFromPayload(p, "iac.bin"))
		}
	case "ansible":
		return RunAnsible(ctx, deps, task, dir)
	default:
		deps.Worker.AppendTaskLog(task.ID, map[string]string{
			"stream": "stdout",
			"line":   fmt.Sprintf("unknown engine %s", engine),
			"msg":    "template",
		})
		return fmt.Errorf("unknown engine %s", engine)
	}
}
