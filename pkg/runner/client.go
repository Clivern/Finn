// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	runv1 "github.com/clivern/finn/proto/runner/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// TaskResponse is what the runner API returns for a task.
type TaskResponse struct {
	ID        string              `json:"id"`
	Status    TaskStatus          `json:"status"`
	Type      TaskType            `json:"type"`
	Log       []map[string]string `json:"log"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// GRPCClient wraps the generated gRPC client with auth and convenience helpers.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client runv1.RunnerServiceClient
	token  string
	target string
}

// NewGRPCClient dials the runner gRPC endpoint using the provided token.
func NewGRPCClient(ctx context.Context, target, token string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("runner grpc dial: %w", err)
	}
	return &GRPCClient{
		conn:   conn,
		client: runv1.NewRunnerServiceClient(conn),
		token:  token,
		target: target,
	}, nil
}

// withAuth attaches the runner token to the outgoing request context.
func (c *GRPCClient) withAuth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "X-Request-Token", c.token)
}

// protoLogToTask converts protobuf log entries into a simple JSON-friendly slice.
func ProtoLogToTask(entries []*runv1.TaskLogEntry) []map[string]string {
	if len(entries) == 0 {
		return nil
	}
	out := make([]map[string]string, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		m := make(map[string]string)
		if e.Stream != "" {
			m["stream"] = e.Stream
		}
		if e.Line != "" {
			m["line"] = e.Line
		}
		if e.Msg != "" {
			m["msg"] = e.Msg
		}
		out = append(out, m)
	}
	return out
}

// Ping checks that the runner is reachable and the token is accepted.
func (c *GRPCClient) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(c.withAuth(ctx), 10*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx, &runv1.PingRequest{})

	return err
}

// Clone enqueues a clone task on the runner and returns its task ID.
func (c *GRPCClient) Clone(ctx context.Context, payload map[string]interface{}) (taskID string, err error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("clone marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(c.withAuth(ctx), 30*time.Second)
	defer cancel()

	resp, err := c.client.Clone(ctx, &runv1.CloneRequest{Payload: data})
	if err != nil {
		return "", err
	}

	return resp.GetId(), nil
}

// RunTemplate enqueues a workspace template run on the runner and returns its task ID.
func (c *GRPCClient) RunTemplate(ctx context.Context, payload map[string]interface{}) (taskID string, err error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("run template marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(c.withAuth(ctx), 30*time.Second)
	defer cancel()

	resp, err := c.client.RunTemplate(ctx, &runv1.RunTemplateRequest{Payload: data})
	if err != nil {
		return "", err
	}

	return resp.GetId(), nil
}

// GetTask fetches the latest state of a task by ID.
func (c *GRPCClient) GetTask(ctx context.Context, taskID string) (*TaskResponse, error) {
	ctx, cancel := context.WithTimeout(c.withAuth(ctx), 10*time.Second)
	defer cancel()

	resp, err := c.client.GetTask(ctx, &runv1.GetTaskRequest{TaskId: taskID})
	if err != nil {
		return nil, err
	}

	task := &TaskResponse{
		ID:        resp.GetId(),
		Status:    TaskStatus(resp.GetStatus()),
		Type:      TaskType(resp.GetType()),
		Log:       ProtoLogToTask(resp.GetLog()),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	if t, err := time.Parse(time.RFC3339, resp.GetCreatedAt()); err == nil {
		task.CreatedAt = t
	}

	if t, err := time.Parse(time.RFC3339, resp.GetUpdatedAt()); err == nil {
		task.UpdatedAt = t
	}

	return task, nil
}

// Close tears down the underlying gRPC connection.
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
