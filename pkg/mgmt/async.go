// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/pkg/runner"
	"github.com/rs/zerolog/log"
)

// Async sends clone requests to runners for workspace repositories.
type Async struct {
	Runners  db.RunnerRepository
	Repos    db.WorkspaceRepoRepository
	Creds db.WorkspaceCredRepository
	SSHKeys  db.WorkspaceSSHKeyRepository
}

// NewAsync returns an Async that uses the given runner and repo repositories.
func NewAsync(
	runners db.RunnerRepository,
	repos db.WorkspaceRepoRepository,
	creds db.WorkspaceCredRepository,
	sshKeys db.WorkspaceSSHKeyRepository,
) *Async {
	return &Async{
		Runners:  runners,
		Repos:    repos,
		Creds: creds,
		SSHKeys:  sshKeys,
	}
}

// GetActiveRunner returns a random active runner.
func (a *Async) GetActiveRunner(ctx context.Context) (*db.Runner, error) {
	runners, err := a.Runners.ListActive()
	if err != nil {
		return nil, fmt.Errorf("get active runners: %w", err)
	}
	if len(runners) == 0 {
		return nil, fmt.Errorf("no active runners found")
	}
	return runners[rand.Intn(len(runners))], nil
}

// RequestClone sends a clone request to the runner via gRPC.
func (a *Async) Clone(ctx context.Context, repo *db.WorkspaceRepo, run *db.Runner) (string, error) {
	client, err := runner.NewGRPCClient(
		ctx,
		strings.TrimSpace(run.URL),
		run.Token,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("runner_name", run.Name).
			Str("runner_id", run.ID.String()).
			Msg("unable to connect to runner")
		return "", fmt.Errorf("unable to connect to runner %s", run.Name)
	}
	defer client.Close()

	payload := map[string]interface{}{
		"repository": repo.GitURL,
		"timeout":    300,
	}
	if repo.GitBranch != "" {
		payload["branch"] = repo.GitBranch
	}
	if repo.GitTag != "" {
		payload["tag"] = repo.GitTag
	}

	if repo.GitCredentialsID != nil {
		if cred, err := a.Creds.GetByIDAndWorkspace(*repo.GitCredentialsID, repo.WorkspaceID); err == nil {
			payload["username"] = cred.Username
			payload["password"] = cred.PasswordOrToken
		}
	}
	if repo.SSHKeyID != nil {
		if key, err := a.SSHKeys.GetByIDAndWorkspace(*repo.SSHKeyID, repo.WorkspaceID); err == nil {
			payload["ssh_private_key"] = key.PrivateKey
			if key.Passphrase != "" {
				payload["ssh_passphrase"] = key.Passphrase
			}
		}
	}

	taskID, err := client.Clone(ctx, payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("runner_name", run.Name).
			Str("runner_id", run.ID.String()).
			Msg("clone failed on runner")
		return "", fmt.Errorf("Clone failed on runner %s: %w", run.Name, err)
	}
	if taskID == "" {
		return "", fmt.Errorf("Clone response missing task id on runner %s", run.Name)
	}

	return taskID, nil
}
