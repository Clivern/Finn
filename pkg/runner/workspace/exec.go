// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package workspace

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"
)

const DefaultExecTimeout = 10 * time.Minute

// ExecOptions configures a command execution.
type ExecOptions struct {
	Dir      string
	RelPath  string
	Image    string
	Command  []string
	Env      map[string]string
	Timeout  time.Duration
	OnOutput func(stream, line string)
}

// Exec runs a command.
func Exec(ctx context.Context, opts *ExecOptions) error {
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	args := []string{
		"run", "--rm",
		"-v", opts.Dir + ":/workspace",
		"-w", path.Join("/workspace", opts.RelPath),
	}
	for k, v := range opts.Env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, opts.Image)
	args = append(args, opts.Command...)

	cmd := exec.CommandContext(ctx, "docker", args...)

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
