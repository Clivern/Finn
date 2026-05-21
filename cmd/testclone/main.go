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
	"fmt"
	"os"
	"time"

	"github.com/clivern/finn/pkg/runner"
	"github.com/clivern/finn/pkg/runner/tasks"
)

func main() {
	sshPrivateKey, _ := os.ReadFile("/Users/ahmedfathy/.ssh/pid_ed25519")

	payload := map[string]interface{}{
		"repository":      "git@github.com:clivern/finn.git",
		"branch":          "main",
		"timeout":         300,
		"ssh_private_key": string(sshPrivateKey),
	}

	task := &runner.Task{
		ID:      fmt.Sprintf("testclone-%d", time.Now().UnixNano()),
		Type:    runner.TaskTypeClone,
		Status:  runner.TaskStatusPending,
		Payload: payload,
	}

	deps := &runner.Deps{
		Worker: runner.NewWorker(nil),
	}

	if err := tasks.Clone(deps, task); err != nil {
		fmt.Fprintf(os.Stderr, "clone failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("clone done")
}
