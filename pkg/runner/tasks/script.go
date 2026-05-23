// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/clivern/finn/pkg/runner"
)

// RunScript runs a script template task.
func RunScript(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	if IsDockerAvailable() {
		return RunScriptDocker(ctx, deps, task, dir)
	}

	return RunScriptLocal(ctx, deps, task, dir)
}

func RunScriptLocal(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	scriptCmd, err := BuildScriptExec(dir, task.Payload)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, scriptCmd.Engine, append([]string{scriptCmd.ScriptPath}, scriptCmd.Args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), scriptCmd.Env...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("script: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("script: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("script: start local: %w", err)
	}

	done := make(chan struct{})
	go appendScriptOutput(stdoutPipe, "stdout", deps, task.ID, done)
	go appendScriptOutput(stderrPipe, "stderr", deps, task.ID, done)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("script: local: %w", err)
	}
	return nil
}

func RunScriptDocker(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	command, err := BuildScriptCommand(dir, task.Payload)
	if err != nil {
		return err
	}

	_ = deps.Worker.AppendTaskLog(task.ID, map[string]string{
		"stream": "stdout",
		"line":   "script docker engine placeholder: " + strings.Join(command, " "),
		"msg":    "template",
	})
	return nil
}

func appendScriptOutput(r io.Reader, stream string, deps *runner.Deps, taskID string, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		_ = deps.Worker.AppendTaskLog(taskID, map[string]string{
			"stream": stream,
			"line":   line,
			"msg":    "template",
		})
	}
}
