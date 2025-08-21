// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"fmt"

	"github.com/clivern/finn/pkg/util"
)

var (
	runnerBucket   = []byte("runner")
	runnerIDKey    = []byte("id")
	runnerTokenKey = []byte("token")
)

// Runner holds the local runner identity. ID identifies the runner; Token is used for signing requests with mgmt.
type Runner struct {
	ID    string
	Token string
}

// Module initializes and loads runner identity from Bolt.
type Module struct {
	Bolt *Bolt
}

// Init creates and stores the runner identity (id and token).
func (m *Module) Init() (*Runner, error) {
	if m.Bolt == nil {
		return nil, fmt.Errorf("runner: bolt is required")
	}

	if m.Bolt.DB() == nil {
		return nil, fmt.Errorf("runner: bolt database is not open")
	}

	existing, _ := m.GetIdentity()
	if existing != nil {
		return nil, fmt.Errorf("runner: already initialized (id: %s)", existing.ID)
	}

	id, err := util.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}

	if err := m.Bolt.CreateBucket(runnerBucket); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}
	if err := m.Bolt.Put(runnerBucket, runnerIDKey, []byte(id)); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}
	if err := m.Bolt.Put(runnerBucket, runnerTokenKey, []byte(id)); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}

	return &Runner{ID: id, Token: id}, nil
}

// GetIdentity loads the runner identity from the DB if it exists. Token is used for signing requests.
func (m *Module) GetIdentity() (*Runner, error) {
	if m.Bolt == nil || m.Bolt.DB() == nil {
		return nil, fmt.Errorf("runner: bolt database is not open")
	}

	id, err := m.Bolt.Get(runnerBucket, runnerIDKey)
	if err != nil || id == nil {
		return nil, err
	}
	token, err := m.Bolt.Get(runnerBucket, runnerTokenKey)
	if err != nil || token == nil {
		return nil, err
	}

	return &Runner{ID: string(id), Token: string(token)}, nil
}
