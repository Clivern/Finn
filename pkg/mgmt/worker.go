// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mgmt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/pkg/runner"
	"github.com/rs/zerolog/log"
)

// PollInterval is how often we poll the runner for task status until success/failed.
const PollInterval = 3 * time.Second

// FollowUpTimeout is the max time to keep polling a task; after this we mark it failed and stop.
const FollowUpTimeout = 30 * time.Minute

// defaultWorker is set by the server at startup so API handlers can enqueue follow-up tasks.
var defaultWorker *Worker

// SetDefaultWorker sets the worker instance used by CloneWorkspaceRepositoryAction for follow-up.
func SetDefaultWorker(w *Worker) {
	defaultWorker = w
}

// DefaultWorker returns the server's mgmt worker, or nil if not set.
func DefaultWorker() *Worker {
	return defaultWorker
}

// FollowUpTask is a task to follow up on: mgmt task ID and runner task type (e.g. "clone").
type FollowUpTask struct {
	ID   db.ID
	Type string
}

// Worker runs on the mgmt server and listens for task IDs and types to follow up on
type Worker struct {
	CloneTasks db.WorkspaceRepositoryCloneTaskRepository
	Runners    db.RunnerRepository
	TaskChan   chan FollowUpTask
}

// NewWorker creates a worker with a buffered channel for task ID + type (buffer size cap).
func NewWorker(cloneTasksRepo db.WorkspaceRepositoryCloneTaskRepository, runnersRepo db.RunnerRepository) *Worker {
	return &Worker{
		CloneTasks: cloneTasksRepo,
		Runners:    runnersRepo,
		TaskChan:   make(chan FollowUpTask, 256),
	}
}

// Enqueue sends a task to the worker for follow-up. Non-blocking; drops if channel full.
func (w *Worker) Enqueue(task FollowUpTask) {
	select {
	case w.TaskChan <- task:
	default:
		log.Warn().
			Str("task_id", task.ID.String()).
			Str("task_type", task.Type).
			Msg("mgmt worker: enqueue dropped, channel full")
	}
}

// Run listens to TaskChan and follows up on each task (ID + type) until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-w.TaskChan:
			if !ok {
				return
			}
			if item.Type != "clone" {
				log.Warn().
					Str("task_id", item.ID.String()).
					Str("task_type", item.Type).
					Msg("mgmt worker: unsupported task type")
				continue
			}
			w.followClone(ctx, item)
		}
	}
}

// followClone follows up on the clone task.
func (w *Worker) followClone(ctx context.Context, task FollowUpTask) {
	t, err := w.CloneTasks.GetByID(task.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("task_id", task.ID.String()).
			Msg("mgmt worker: get clone task failed")
		return
	}

	run, err := w.Runners.GetByID(*t.RunnerID)
	if err != nil {
		log.Error().
			Err(err).
			Str("task_id", task.ID.String()).
			Str("runner_id", t.RunnerID.String()).
			Msg("mgmt worker: get runner failed")
		return
	}

	client, err := runner.NewGRPCClient(ctx, strings.TrimSpace(run.URL), run.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("task_id", task.ID.String()).
			Str("runner_name", run.Name).
			Str("runner_id", run.ID.String()).
			Msg("mgmt worker: runner grpc client failed")
		_ = w.CloneTasks.UpdateStatus(task.ID, "failed", errorAsTaskLog(err.Error(), "clone"))
		return
	}
	defer client.Close()

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()
	deadline := time.Now().Add(FollowUpTimeout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if time.Now().After(deadline) {
			log.Warn().
				Str("task_id", task.ID.String()).
				Str("runner_task_id", t.RunnerTaskID).
				Msg("mgmt worker: follow-up timeout")
			_ = w.CloneTasks.UpdateStatus(
				task.ID, "failed",
				errorAsTaskLog(
					"follow-up timeout: task did not reach success/failed within "+FollowUpTimeout.String(),
					"clone",
				),
			)
			return
		}

		resp, err := client.GetTask(ctx, t.RunnerTaskID)
		if err != nil {
			log.Error().
				Err(err).
				Str("task_id", task.ID.String()).
				Str("runner_task_id", t.RunnerTaskID).
				Msg("mgmt worker: get runner task failed")
			_ = w.CloneTasks.UpdateStatus(
				task.ID, "failed", errorAsTaskLog(err.Error(), "clone"),
			)
			return
		}

		if resp == nil {
			return
		}

		var result *string
		status := string(resp.Status)
		if len(resp.Log) > 0 {
			raw, _ := json.Marshal(resp.Log)
			s := string(raw)
			result = &s
		}

		if err := w.CloneTasks.UpdateStatus(task.ID, status, result); err != nil {
			log.Error().
				Err(err).
				Str("task_id", task.ID.String()).
				Msg("mgmt worker: update status failed")
			return
		}

		if status == "success" || status == "failed" {
			return
		}
	}
}

// errorAsTaskLog returns result as a task-log JSON array: [{"stream":"stderr","line":"...","msg":"..."}].
func errorAsTaskLog(line, msg string) *string {
	if msg == "" {
		msg = "clone"
	}
	entries := []map[string]string{{
		"stream": "stderr",
		"line":   line,
		"msg":    msg,
	}}
	raw, _ := json.Marshal(entries)
	s := string(raw)
	return &s
}
