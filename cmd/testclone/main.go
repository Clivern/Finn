// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// testclone runs the runner clone task locally (no gRPC server).
//
// Usage:
//
//	go run ./cmd/testclone
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
		"repository":      "git@github.com:clivern/finn.git",
		"branch":          "main",
		"timeout":         300,
		"ssh_private_key": string(sshPrivateKey),
	}

	id, err := uuid.NewRandom()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create task id: %s\n", err)
		os.Exit(1)
	}

	bolt := &runner.Bolt{
		Path: filepath.Join(os.TempDir(), fmt.Sprintf("finn-testclone-%s.db", id.String())),
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
		Type:      runner.TaskTypeClone,
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_ = worker.UpdateTaskStatus(task.ID, runner.TaskStatusRunning)
	cloneErr := tasks.Clone(ctx, deps, task)
	if cloneErr != nil {
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

	if cloneErr != nil {
		fmt.Fprintf(os.Stderr, "clone failed: %s\n", cloneErr)
		os.Exit(1)
	}

	fmt.Println("clone done")
}
