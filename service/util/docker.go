// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// CloneAndRun clones a repo then runs a command in a Docker container.
func CloneAndRun(ctx context.Context, cloneURL, image string, command []string, timeout time.Duration) (*RunResult, error) {
	dir, err := os.MkdirTemp("", "clone-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	if err := gitClone(ctx, cloneURL, dir, timeout); err != nil {
		return nil, fmt.Errorf("git clone: %w", err)
	}

	return RunInContainer(ctx, dir, image, command, timeout)
}

// RunInContainer runs a command in a Docker container with the given image and work dir.
func RunInContainer(ctx context.Context, workDir, image string, command []string, timeout time.Duration) (*RunResult, error) {
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve work dir: %w", err)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// docker run --rm -v workDir:/workspace -w /workspace image cmd...
	args := []string{
		"run", "--rm",
		"-v", workDir + ":/workspace",
		"-w", "/workspace",
		image,
	}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	var runErr error
	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		runErr = ctx.Err()
	case runErr = <-done:
	}

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &RunResult{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: -1,
				Err:      runErr,
			}, nil
		}
	}

	return &RunResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Err:      runErr,
	}, nil
}

// RunInContainerStream runs a command in a container, streaming stdout and stderr to the given writers.
func RunInContainerStream(ctx context.Context, workDir, image string, command []string, stdout, stderr io.Writer) (int, error) {
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return -1, fmt.Errorf("resolve work dir: %w", err)
	}

	args := []string{"run", "--rm", "-v", workDir + ":/workspace", "-w", "/workspace", image}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, err
		}
	}
	return exitCode, nil
}

// gitClone clones a Git repo into the given directory.
func gitClone(ctx context.Context, cloneURL, dir string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", cloneURL, ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RunWithLiveOutput runs a command in a container and streams output to stdout/stderr.
func RunWithLiveOutput(ctx context.Context, workDir, image string, command []string) (int, error) {
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return -1, err
	}

	args := []string{"run", "--rm", "-v", workDir + ":/workspace", "-w", "/workspace", image}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return -1, err
	}

	scanOut := bufio.NewScanner(stdout)
	scanErr := bufio.NewScanner(stderr)
	go func() {
		for scanOut.Scan() {
			fmt.Println(scanOut.Text())
		}
	}()
	go func() {
		for scanErr.Scan() {
			fmt.Fprintln(os.Stderr, scanErr.Text())
		}
	}()

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, err
		}
	}
	return exitCode, nil
}
