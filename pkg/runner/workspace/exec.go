// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package workspace

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const DefaultExecTimeout = 10 * time.Minute

// ExecOptions configures a command execution inside a Docker container
// with a workspace directory mounted.
type ExecOptions struct {
	Dir     string            // host directory to mount as /workspace
	Image   string            // Docker image to run
	Command []string          // command and arguments
	Env     map[string]string // extra environment variables
	Timeout time.Duration

	// OnOutput is called for each line of stdout/stderr.
	OnOutput func(stream, line string)
}

// Exec runs a command inside a Docker container with opts.Dir mounted at /workspace.
func Exec(ctx context.Context, opts *ExecOptions) error {
	if opts.Dir == "" {
		return fmt.Errorf("exec: workspace directory is required")
	}
	if opts.Image == "" {
		return fmt.Errorf("exec: Docker image is required")
	}
	if len(opts.Command) == 0 {
		return fmt.Errorf("exec: command is required")
	}

	absDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return fmt.Errorf("exec: resolve dir: %w", err)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultExecTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{
		"run", "--rm",
		"-v", absDir + ":/workspace",
		"-w", "/workspace",
	}
	for k, v := range opts.Env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, opts.Image)
	args = append(args, opts.Command...)

	cmd := exec.CommandContext(runCtx, "docker", args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("exec: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("exec: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("exec: start: %w", err)
	}

	done := make(chan struct{})
	go func() {
		scanLines(stdoutPipe, "stdout", opts.OnOutput)
		done <- struct{}{}
	}()
	go func() {
		scanLines(stderrPipe, "stderr", opts.OnOutput)
		done <- struct{}{}
	}()

	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func scanLines(r interface{ Read([]byte) (int, error) }, stream string, handler func(string, string)) {
	if handler == nil {
		return
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			handler(stream, line)
		}
	}
}
