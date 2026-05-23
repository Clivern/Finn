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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/tasks"
	"github.com/google/uuid"
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

	id, err := uuid.NewRandom()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create task id: %s\n", err)
		os.Exit(1)
	}

	bolt := &runner.Bolt{
		Path: filepath.Join(os.TempDir(), fmt.Sprintf("finn-testscript-%s.db", id.String())),
	}
	if err := bolt.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "open bolt: %s\n", err)
		os.Exit(1)
	}
	defer bolt.Close()

	worker := runner.NewWorker(bolt)
	now := time.Now().UTC()
	task := &runner.Task{
		ID:        id.String(),
		Type:      runner.TaskTypeTemplate,
		Status:    runner.TaskStatusPending,
		Payload:   payload,
		Log:       []map[string]string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	deps := &runner.Deps{
		Bolt:   bolt,
		Worker: worker,
	}

	if err := worker.SaveTask(task); err != nil {
		fmt.Fprintf(os.Stderr, "save task: %s\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	_ = worker.UpdateTaskStatus(task.ID, runner.TaskStatusRunning)
	templateErr := tasks.WorkspaceTemplate(ctx, deps, task)
	if templateErr != nil {
		_ = worker.UpdateTaskStatus(task.ID, runner.TaskStatusFailed)
	} else {
		_ = worker.UpdateTaskStatus(task.ID, runner.TaskStatusSuccess)
	}

	storedTask, err := worker.GetTask(task.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get task: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("task id: %s\n", task.ID)
	fmt.Printf("bolt db: %s\n", bolt.Path)
	fmt.Println("task output:")
	for _, entry := range storedTask.Log {
		fmt.Printf("[%s] %s\n", entry["stream"], entry["line"])
	}

	if templateErr != nil {
		fmt.Fprintf(os.Stderr, "template failed: %s\n", templateErr)
		os.Exit(1)
	}

	fmt.Println("script template done")
}
