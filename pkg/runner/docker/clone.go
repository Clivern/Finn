// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package docker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// CloneInDocker runs git clone inside a Docker container and streams stdout/stderr to the log.
func CloneInDocker(ctx context.Context, c *Cloner) error {
	if c == nil || c.RepoURL == "" {
		return fmt.Errorf("git: repo URL is required")
	}
	if c.Dir == "" {
		return fmt.Errorf("git: dir is required")
	}

	dir := c.Dir
	if err := ensureEmptyCloneDir(dir); err != nil {
		return err
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	cloneURL := c.RepoURL
	if c.Username != "" || c.Password != "" {
		var authErr error
		cloneURL, authErr = addAuthFromStrings(c.RepoURL, c.Username, c.Password)
		if authErr != nil {
			return fmt.Errorf("git: build HTTPS URL: %w", authErr)
		}
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = DefaultCloneTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("git: resolve clone dir: %w", err)
	}

	image := TestCloneDockerImage
	args := []string{"run", "--rm",
		"-v", absDir + ":/workspace",
		"-w", "/workspace",
	}

	if c.SSHPrivateKey != "" {
		keyPath, cleanup, keyErr := addSSHKey(c.SSHPrivateKey)
		if keyErr != nil {
			return fmt.Errorf("git: write SSH key: %w", keyErr)
		}
		defer cleanup()
		args = append(
			args, "-v", keyPath+":/tmp/git-key:ro",
		)
		args = append(
			args, "-e", "GIT_SSH_COMMAND=ssh -i /tmp/git-key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
		)
	}

	args = append(args, image, "clone", "--depth", "1")
	if c.Tag != "" {
		args = append(args, "--branch", c.Tag)
	} else if c.Branch != "" {
		args = append(args, "--branch", c.Branch)
	}
	args = append(args, cloneURL, ".")

	cmd := exec.CommandContext(runCtx, "docker", args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("git clone (docker): stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("git clone (docker): stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("git clone (docker): start: %w", err)
	}

	msg := "git clone"
	if c.StreamMessage != "" {
		msg = c.StreamMessage
	}

	go streamLogLines(stdoutPipe, "stdout", msg, c.StreamHandler)
	go streamLogLines(stderrPipe, "stderr", msg, c.StreamHandler)

	runErr := cmd.Wait()
	if runErr != nil {
		return fmt.Errorf("git clone (docker): %w", runErr)
	}

	revParseMsg := "git rev-parse HEAD"
	if c.StreamMessage != "" {
		revParseMsg = c.StreamMessage + " rev-parse"
	}

	revErr := runGitRevParseInDocker(
		runCtx,
		absDir,
		image,
		revParseMsg,
		c.StreamHandler,
	)
	if revErr != nil {
		log.Warn().
			Err(revErr).
			Msg("git clone: rev-parse after clone failed")
	}

	return nil
}

// ensureEmptyCloneDir removes existing contents of dir and recreates it empty so git clone . works.
func ensureEmptyCloneDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("git: clear clone dir: %w", err)
	}
	return os.MkdirAll(dir, 0o750)
}

// runGitRevParseInDocker runs git rev-parse HEAD in the cloned dir and streams output to the log.
func runGitRevParseInDocker(ctx context.Context, absDir, image, msg string, handler func(StreamEvent)) error {
	args := []string{
		"run", "--rm",
		"-v", absDir + ":/workspace",
		"-w", "/workspace",
		image,
		"rev-parse",
		"HEAD",
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go streamLogLines(stdoutPipe, "stdout", msg, handler)
	go streamLogLines(stderrPipe, "stderr", msg, handler)
	return cmd.Wait()
}

// streamLogLines streams log lines from a pipe to a handler.
func streamLogLines(pipe interface {
	Read(p []byte) (n int, err error)
}, stream string, msg string, handler func(StreamEvent)) {
	if msg == "" {
		msg = "git clone"
	}
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if handler != nil {
			handler(StreamEvent{Stream: stream, Line: line, Msg: msg})
		} else {
			if stream == "stderr" {
				log.Info().
					Str("stream", "stderr").
					Str("line", line).
					Msg(msg)
			} else {
				log.Info().
					Str("stream", "stdout").
					Str("line", line).
					Msg(msg)
			}
		}
	}
}
