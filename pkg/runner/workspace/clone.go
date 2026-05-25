// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package workspace

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/rs/zerolog/log"
	gossh "golang.org/x/crypto/ssh"
)

const DefaultCloneTimeout = 5 * time.Minute

// CloneOptions configures a git clone operation.
type CloneOptions struct {
	URL           string
	Branch        string
	Tag           string
	Dir           string
	Timeout       time.Duration
	Username      string
	Password      string
	SSHPrivateKey string
	SSHPassphrase string

	OnProgress func(msg string)
}

// Clone performs a shallow git clone into opts.Dir
func Clone(ctx context.Context, opts *CloneOptions) (string, error) {
	if err := os.MkdirAll(opts.Dir, 0o750); err != nil {
		return "", fmt.Errorf("clone: create dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	auth, err := buildAuth(opts)
	if err != nil {
		return "", fmt.Errorf("clone: auth: %w", err)
	}

	copts := &git.CloneOptions{
		URL:   opts.URL,
		Auth:  auth,
		Depth: 1,
	}

	if opts.Tag != "" {
		copts.ReferenceName = plumbing.NewTagReferenceName(opts.Tag)
		copts.SingleBranch = true
	} else if opts.Branch != "" {
		copts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		copts.SingleBranch = true
	}

	if opts.OnProgress != nil {
		copts.Progress = &lineWriter{callback: opts.OnProgress}
	}

	log.Info().
		Str("url", opts.URL).
		Str("dir", opts.Dir).
		Msg("Starting git clone")

	repo, err := git.PlainCloneContext(ctx, opts.Dir, false, copts)
	if err != nil {
		return "", fmt.Errorf("clone: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("clone: resolve HEAD: %w", err)
	}

	return head.Hash().String(), nil
}

func buildAuth(opts *CloneOptions) (transport.AuthMethod, error) {
	if opts.SSHPrivateKey != "" {
		publicKeys, err := ssh.NewPublicKeys(
			"git",
			[]byte(opts.SSHPrivateKey),
			opts.SSHPassphrase,
		)
		if err != nil {
			return nil, fmt.Errorf("SSH key: %w", err)
		}
		publicKeys.HostKeyCallback = gossh.InsecureIgnoreHostKey()
		return publicKeys, nil
	}

	if opts.Username != "" || opts.Password != "" {
		return &http.BasicAuth{
			Username: opts.Username,
			Password: opts.Password,
		}, nil
	}

	return nil, nil
}

type lineWriter struct {
	callback func(msg string)
}

func (w *lineWriter) Write(p []byte) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(p)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			w.callback(line)
		}
	}
	return len(p), nil
}

var _ io.Writer = (*lineWriter)(nil)
