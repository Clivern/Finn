// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"context"
	"encoding/json"
	"strings"

	runv1 "github.com/clivern/finn/proto/runner/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RunnerGRPCService is the server implementation of the RunnerService.
type RunnerGRPCService struct {
	runv1.UnimplementedRunnerServiceServer
	deps *Deps
}

// NewRunnerGRPCService creates a new RunnerGRPCService.
func NewRunnerGRPCService(deps *Deps) *RunnerGRPCService {
	return &RunnerGRPCService{deps: deps}
}

// TokenFromContext retrieves the token from the context.
func TokenFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, k := range []string{"X-Request-Token"} {
		if v := md.Get(k); len(v) > 0 && v[0] != "" {
			return strings.TrimSpace(v[0])
		}
	}
	return ""
}

// LoggingUnaryInterceptor logs each incoming gRPC request.
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Info().
			Str("method", info.FullMethod).
			Msg("runner: incoming request")
		return handler(ctx, req)
	}
}

// AuthUnaryInterceptor intercepts a unary RPC and checks if the request is authenticated.
func AuthUnaryInterceptor(deps *Deps) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if deps == nil || deps.Token == "" {
			return nil, status.Error(codes.Internal, "runner identity unavailable")
		}
		token := TokenFromContext(ctx)
		if token == "" || token != deps.Token {
			return nil, status.Error(codes.Unauthenticated, "invalid or missing runner token")
		}
		return handler(ctx, req)
	}
}

// Ping checks if the runner is alive.
func (s *RunnerGRPCService) Ping(ctx context.Context, _ *runv1.PingRequest) (*runv1.PingResponse, error) {
	return &runv1.PingResponse{Status: "ok"}, nil
}

// Clone creates a new task and returns the task ID.
func (s *RunnerGRPCService) Clone(ctx context.Context, in *runv1.CloneRequest) (*runv1.CloneResponse, error) {
	payload := make(map[string]interface{})
	if len(in.GetPayload()) > 0 {
		if err := json.Unmarshal(in.Payload, &payload); err != nil {
			return nil, status.Error(codes.InvalidArgument, "payload must be valid JSON")
		}
	}

	task, err := s.deps.Worker.Enqueue(s.deps, TaskTypeClone, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &runv1.CloneResponse{
		Id:        task.ID,
		Status:    string(task.Status),
		Type:      string(task.Type),
		Log:       TaskLogToProto(task.Log),
		CreatedAt: task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// RunTemplate enqueues a workspace template run (clone + automation).
func (s *RunnerGRPCService) RunTemplate(ctx context.Context, in *runv1.RunTemplateRequest) (*runv1.RunTemplateResponse, error) {
	payload := make(map[string]interface{})
	if len(in.GetPayload()) > 0 {
		if err := json.Unmarshal(in.Payload, &payload); err != nil {
			return nil, status.Error(codes.InvalidArgument, "payload must be valid JSON")
		}
	}

	task, err := s.deps.Worker.Enqueue(s.deps, TaskTypeTemplate, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &runv1.RunTemplateResponse{
		Id:        task.ID,
		Status:    string(task.Status),
		Type:      string(task.Type),
		Log:       TaskLogToProto(task.Log),
		CreatedAt: task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetTask retrieves a task by its ID.
func (s *RunnerGRPCService) GetTask(ctx context.Context, in *runv1.GetTaskRequest) (*runv1.GetTaskResponse, error) {
	taskID := strings.TrimSpace(in.GetTaskId())
	if taskID == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	task, err := s.deps.Worker.GetTask(taskID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if task == nil {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	return &runv1.GetTaskResponse{
		Id:        task.ID,
		Status:    string(task.Status),
		Type:      string(task.Type),
		Log:       TaskLogToProto(task.Log),
		CreatedAt: task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// TaskLogToProto converts a slice of map[string]string to a slice of *runv1.TaskLogEntry.
func TaskLogToProto(log []map[string]string) []*runv1.TaskLogEntry {
	if len(log) == 0 {
		return nil
	}

	out := make([]*runv1.TaskLogEntry, 0, len(log))
	for _, m := range log {
		out = append(out, &runv1.TaskLogEntry{
			Stream: GetStr(m, "stream"),
			Line:   GetStr(m, "line"),
			Msg:    GetStr(m, "msg"),
		})
	}

	return out
}

// getStr returns the string value from the map.
func GetStr(m map[string]string, key string) string {
	if m == nil {
		return ""
	}

	return m[key]
}
