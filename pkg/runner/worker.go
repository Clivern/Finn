// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	tasksBucket = []byte("tasks")
)

// TaskHandler runs a task. It is registered per TaskType.
type TaskHandler func(deps *Deps, task *Task) error

// taskHandlers maps task types to their handlers.
var taskHandlers = make(map[TaskType]TaskHandler)

// TaskStatus is the lifecycle of a runner task (pending, running, success, failed).
type TaskStatus string

// TaskType is the kind of task (e.g. clone).
type TaskType string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusFailed  TaskStatus = "failed"
	TaskTypeClone     TaskType   = "clone"
)

// Task is a stored unit of work
type Task struct {
	ID        string                 `json:"id"`
	Status    TaskStatus             `json:"status"`
	Type      TaskType               `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Log       []map[string]string    `json:"log"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// RegisterTaskHandler registers a handler for the given task type.
func RegisterTaskHandler(taskType TaskType, h TaskHandler) {
	taskHandlers[taskType] = h
}

// Worker runs runner workloads in separate goroutines.
type Worker struct {
	Bolt     *Bolt
	taskChan chan string
	mu       sync.Mutex
}

// NewWorker creates a Worker that uses the given Bolt DB for task storage.
func NewWorker(bolt *Bolt) *Worker {
	return &Worker{
		Bolt:     bolt,
		taskChan: make(chan string, 256),
	}
}

// Enqueue stores a new task in Bolt and returns the task ID.
func (w *Worker) Enqueue(deps *Deps, taskType TaskType, payload map[string]interface{}) (*Task, error) {
	if w.Bolt == nil || w.Bolt.DB() == nil {
		return nil, fmt.Errorf("worker: bolt database is not open")
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("worker: generate task id: %w", err)
	}

	taskID := id.String()
	now := time.Now().UTC()
	task := &Task{
		ID:        taskID,
		Status:    TaskStatusPending,
		Type:      taskType,
		Payload:   payload,
		Log:       []map[string]string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := w.SaveTask(task); err != nil {
		return nil, err
	}

	w.taskChan <- taskID

	return task, nil
}

// GetTask returns the task by ID from Bolt, or nil if not found.
func (w *Worker) GetTask(taskID string) (*Task, error) {
	if w.Bolt == nil || w.Bolt.DB() == nil {
		return nil, fmt.Errorf("worker: bolt database is not open")
	}

	raw, err := w.Bolt.Get(tasksBucket, []byte(taskID))
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	var task Task
	if err := json.Unmarshal(raw, &task); err != nil {
		return nil, fmt.Errorf("worker: decode task: %w", err)
	}
	return &task, nil
}

// SaveTask stores the task in Bolt.
func (w *Worker) SaveTask(task *Task) error {
	task.UpdatedAt = time.Now().UTC()
	raw, err := json.Marshal(task)

	if err != nil {
		return fmt.Errorf("worker: encode task: %w", err)
	}

	if err := w.Bolt.CreateBucket(tasksBucket); err != nil {
		return err
	}

	return w.Bolt.Put(tasksBucket, []byte(task.ID), raw)
}

// UpdateTaskStatus updates the task status in Bolt.
func (w *Worker) UpdateTaskStatus(taskID string, status TaskStatus) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	task, err := w.GetTask(taskID)
	if err != nil || task == nil {
		return fmt.Errorf("worker: get task: %w", err)
	}

	task.Status = status
	err = w.SaveTask(task)

	if err != nil {
		return fmt.Errorf("worker: update task status: %w", err)
	}

	return nil
}

// AppendTaskLog appends a log line to the task in Bolt.
func (w *Worker) AppendTaskLog(taskID string, log map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	task, err := w.GetTask(taskID)
	if err != nil || task == nil {
		return fmt.Errorf("worker: get task: %w", err)
	}

	task.Log = append(task.Log, log)

	err = w.SaveTask(task)
	if err != nil {
		return fmt.Errorf("worker: append task log: %w", err)
	}

	return nil
}

// Run listens to taskChan and launches a goroutine per task to handle work by task type.
func (w *Worker) Run(ctx context.Context, deps *Deps) {
	for {
		select {
		case <-ctx.Done():
			return
		case taskID, ok := <-w.taskChan:
			if !ok {
				return
			}
			go w.handleTask(ctx, deps, taskID)
		}
	}
}

// handleTask loads the task, looks up the handler for its type, and runs it.
func (w *Worker) handleTask(ctx context.Context, deps *Deps, taskID string) {
	task, err := w.GetTask(taskID)
	if err != nil || task == nil {
		return
	}

	handler := taskHandlers[task.Type]
	if handler == nil {
		w.UpdateTaskStatus(taskID, TaskStatusFailed)
		w.AppendTaskLog(taskID, map[string]string{
			"stream": "stderr",
			"line":   "no handler for task type " + string(task.Type),
			"msg":    string(task.Type),
		})
		return
	}

	if err := w.UpdateTaskStatus(taskID, TaskStatusRunning); err != nil {
		return
	}

	if err := handler(deps, task); err != nil {
		w.UpdateTaskStatus(taskID, TaskStatusFailed)
		w.AppendTaskLog(taskID, map[string]string{
			"stream": "stderr",
			"line":   err.Error(),
			"msg":    string(task.Type),
		})
		return
	}

	w.UpdateTaskStatus(taskID, TaskStatusSuccess)
}
