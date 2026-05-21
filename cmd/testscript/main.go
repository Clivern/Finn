// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// testscript runs a script workspace template locally (no gRPC server).
//
// Usage:
//
//	go run ./cmd/testscript
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/tasks"
)

func main() {
	payload := map[string]interface{}{
		"name":             "testscript",
		"description":      "testscript",
		"engine":           "script",
		"workingDirectory": "/",
		"branchOverride":   "main",
		"tagOverride":      "",
		"runnerTags": map[string]string{
			"key": "value",
		},
		"repository": map[string]interface{}{
			"name":       "repository-name",
			"git_url":    "https://github.com/clivern/finn.git",
			"git_branch": "main",
			"git_tag":    "v1.0.0",
			"ssh_key": map[string]interface{}{
				"name":            "my-deploy-key",
				"committer_name":  "John Doe",
				"committer_email": "john.doe@example.com",
				"private_key":     "ssh-private-key",
				"public_key":      "ssh-public-key",
				"passphrase":      "ssh-passphrase",
			},
			"git_credential": map[string]interface{}{
				"name":              "github-token",
				"committer_name":    "John Doe",
				"committer_email":   "john.doe@example.com",
				"username":          "username",
				"password_or_token": "password-or-token",
			},
		},
		"allowConcurrentRuns": true,
		"maxConcurrentRuns":   1,
		"approvalRequired":    false,
		"script": map[string]interface{}{
			"engine":        "bash",
			"engineVersion": "",
			"scriptPath":    "etc/runner-ping.sh",
		},
		"extraArgs": map[string]interface{}{
			"key":  "value",
			"key2": "value2",
		},
		"groupVars": map[string]interface{}{
			"name": "variable-group-name",
			"extraVariables": map[string]interface{}{
				"key":  "value",
				"key2": "value2",
			},
			"environmentVariables": map[string]interface{}{
				"key":  "value",
				"key2": "value2",
			},
			"secretExtraVariables": map[string]interface{}{
				"key":  "value",
				"key2": "value2",
			},
			"secretEnvironmentVariables": map[string]interface{}{
				"key":  "value",
				"key2": "value2",
			},
		},
	}

	task := &runner.Task{
		ID:      fmt.Sprintf("testscript-%d", time.Now().UnixNano()),
		Type:    runner.TaskTypeTemplate,
		Status:  runner.TaskStatusPending,
		Payload: payload,
	}

	deps := &runner.Deps{
		Worker: runner.NewWorker(nil),
	}

	if err := tasks.WorkspaceTemplate(deps, task); err != nil {
		fmt.Fprintf(os.Stderr, "template failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("script template done")
}
