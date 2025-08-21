// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package docker

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const DefaultCloneTimeout = 5 * time.Minute

// StreamEvent is one line of output from a clone or rev-parse stream (for DB or channel consumers).
type StreamEvent struct {
	Stream string // "stdout" or "stderr"
	Line   string
	Msg    string // e.g. "git clone", "git rev-parse HEAD"
}

// Cloner holds options for cloning a git repository.
type Cloner struct {
	RepoURL       string
	Branch        string
	Tag           string
	Dir           string
	SSHPrivateKey string
	Username      string
	Password      string
	Timeout       time.Duration

	// StreamMessage is the label used for stream events and logs (e.g. "git clone", or a job ID).
	// If empty, "git clone" is used for clone output and "git rev-parse HEAD" for rev-parse output.
	StreamMessage string

	// StreamHandler is called for each line from clone/rev-parse output. Use it to send to a channel or persist to DB.
	// Called from the same goroutine that reads the stream; avoid blocking (e.g. send to buffered channel or enqueue).
	//
	// Example: send to channel for a DB writer goroutine
	//
	//   events := make(chan docker.StreamEvent, 256)
	//   go func() {
	//     for e := range events {
	//       db.InsertCloneLog(ctx, jobID, e.Stream, e.Line, e.Msg)
	//     }
	//   }()
	//   cloner.StreamHandler = func(e docker.StreamEvent) { events <- e }
	//   err := docker.CloneInDocker(ctx, cloner)
	//   close(events) // after clone finishes so the writer goroutine exits
	//
	StreamHandler func(event StreamEvent)
}

// Clone runs git clone into c.Dir using c.RepoURL and optional auth.
func (c *Cloner) Clone(ctx context.Context) (cloneDir string, err error) {
	if c.RepoURL == "" {
		return "", fmt.Errorf("git: repo URL is required")
	}
	if c.Dir == "" {
		return "", fmt.Errorf("git: dir is required")
	}

	cloneDir = c.Dir
	if err := os.MkdirAll(cloneDir, 0o750); err != nil {
		return "", fmt.Errorf("git: create clone dir: %w", err)
	}

	cleanupOnErr := true
	defer func() {
		if cleanupOnErr && err != nil {
			_ = os.RemoveAll(cloneDir)
		}
	}()

	cloneURL := c.RepoURL
	var env []string

	if c.Username != "" || c.Password != "" {
		cloneURL, err = addAuthFromStrings(c.RepoURL, c.Username, c.Password)
		if err != nil {
			return "", fmt.Errorf("git: build HTTPS URL: %w", err)
		}
	}

	if c.SSHPrivateKey != "" {
		keyPath, cleanup, err := addSSHKey(c.SSHPrivateKey)
		if err != nil {
			return "", fmt.Errorf("git: write SSH key: %w", err)
		}
		defer cleanup()
		env = append(
			env,
			fmt.Sprintf(
				"GIT_SSH_COMMAND=ssh -i %q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
				keyPath,
			),
		)
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = DefaultCloneTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{"clone", "--depth", "1"}
	if c.Tag != "" {
		args = append(args, "--branch", c.Tag)
	} else if c.Branch != "" {
		args = append(args, "--branch", c.Branch)
	}
	args = append(args, cloneURL, ".")

	cmd := exec.CommandContext(runCtx, "git", args...)
	cmd.Dir = cloneDir
	cmd.Env = append(os.Environ(), env...)
	out, runErr := cmd.CombinedOutput()

	if runErr != nil {
		return "", fmt.Errorf("git clone: %w: %s", runErr, strings.TrimSpace(string(out)))
	}

	cleanupOnErr = false
	return cloneDir, nil
}

// addAuthFromStrings adds authentication to a URL from username and password strings.
func addAuthFromStrings(rawURL, user, pass string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if user != "" || pass != "" {
		u.User = url.UserPassword(user, pass)
	}
	return u.String(), nil
}

// addSSHKey adds an SSH key to a temporary file and returns the path and cleanup function.
func addSSHKey(privateKey string) (path string, cleanup func(), err error) {
	dir := os.TempDir()
	f, err := os.CreateTemp(dir, "git-key-*")
	if err != nil {
		return "", nil, err
	}

	path = f.Name()
	cleanup = func() { _ = os.Remove(path) }
	keyContent := strings.TrimSpace(privateKey)

	if keyContent == "" {
		cleanup()
		return "", nil, fmt.Errorf("empty private key")
	}
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}
	if _, err := f.WriteString(keyContent); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := f.Chmod(0o600); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}
