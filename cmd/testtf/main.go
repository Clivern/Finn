// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// testtf runs a Terraform workspace template locally (no gRPC server).
//
// Usage:
//
//	go run ./cmd/testtf
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
	sshPrivateKey, _ := os.ReadFile("/Users/ahmedfathy/.ssh/pid_ed25519")

	payload := map[string]interface{}{
		"name":            "testtf",
		"description":     "testtf",
		"engine":          "terraform",
		"working_dir":     "etc/terraform",
		"branch_override": "dev",
		"timeout":         600,
		"tag_override":    "",
		"runner_tags": map[string]string{
			"key": "value",
		},
		"repository": map[string]interface{}{
			"name":       "finn",
			"git_url":    "git@github.com:Clivern/finn.git",
			"git_branch": "main",
			"ssh_key": map[string]interface{}{
				"name":            "my-deploy-key",
				"committer_name":  "John Doe",
				"committer_email": "john.doe@example.com",
				"private_key":     string(sshPrivateKey),
			},
		},
		"iac": map[string]interface{}{
			"version":   "1.11",
			"log_level": "INFO",
			"steps":     []string{"init", "plan", "apply"},
			"backend": map[string]interface{}{
				"id":       "backend-id",
				"name":     "dh",
				"slug":     "ail49cjb2g",
				"username": "tf-cq7lbxnb5c3qwr",
				"password": "&65zDv2lwmWqXZMiG4z5wDFx",
				"http": map[string]interface{}{
					"address":        "http://localhost:8000/client/ail49cjb2g/state",
					"lock_address":   "http://localhost:8000/client/ail49cjb2g/lock",
					"unlock_address": "http://localhost:8000/client/ail49cjb2g/unlock",
					"lock_method":    "POST",
					"unlock_method":  "POST",
				},
			},
		},
		"extra_args": map[string]interface{}{
			"example": "finn",
		},
		"group_vars": map[string]interface{}{
			"name": "variable-group-name",
			"extra_variables": map[string]interface{}{
				"evkey": "value",
			},
			"environment_variables": map[string]interface{}{
				"TF_IN_AUTOMATION": "true",
			},
			"secret_extra_variables":       map[string]interface{}{},
			"secret_environment_variables": map[string]interface{}{},
		},
	}

	id, err := uuid.NewRandom()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create task id: %s\n", err)
		os.Exit(1)
	}

	bolt := &runner.Bolt{
		Path: filepath.Join(os.TempDir(), fmt.Sprintf("finn-testtf-%s.db", id.String())),
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
	templateDone := make(chan error, 1)
	go func() {
		templateDone <- tasks.WorkspaceTemplate(ctx, deps, task)
	}()

	var templateErr error
	templateFinished := false
	printedLogs := 0
	for !templateFinished {
		select {
		case templateErr = <-templateDone:
			templateFinished = true
		case <-time.After(250 * time.Millisecond):
		}

		storedTask, err := worker.GetTask(task.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get task logs: %s\n", err)
			break
		}
		for printedLogs < len(storedTask.Log) {
			entry := storedTask.Log[printedLogs]
			fmt.Printf("[%s] %s\n", entry["stream"], entry["line"])
			printedLogs++
		}
	}

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
	for _, entry := range storedTask.Log[printedLogs:] {
		fmt.Printf("[%s] %s\n", entry["stream"], entry["line"])
	}

	if templateErr != nil {
		fmt.Fprintf(os.Stderr, "terraform template failed: %s\n", templateErr)
		os.Exit(1)
	}

	fmt.Println("terraform template done")
}
