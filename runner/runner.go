// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"fmt"
)

var (
	runnerBucket   = []byte("runner")
	runnerIDKey    = []byte("id")
	runnerTokenKey = []byte("token")
	runnerPubKey   = []byte("public_key")
	runnerPrivKey  = []byte("private_key")
)

// Runner holds the local runner identity (RSA key pair for signing and encryption).
type Runner struct {
	ID         string
	PublicKey  []byte
	PrivateKey []byte
}

// Module initializes and loads runner
type Module struct {
	Bolt *Bolt
	Key  *Key
}

// Init creates a new runner and stores it in the database
func (m *Module) Init() (*Runner, error) {
	if m.Bolt == nil || m.Key == nil {
		return nil, fmt.Errorf("runner: bolt and key are required")
	}

	if m.Bolt.DB() == nil {
		return nil, fmt.Errorf("runner: bolt database is not open")
	}

	existing, _ := m.GetIdentity()
	if existing != nil {
		return nil, fmt.Errorf("runner: already initialized (id: %s)", existing.ID)
	}

	id, err := m.Key.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}

	pair, err := m.Key.GenerateKeyPair()
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
	if err := m.Bolt.Put(runnerBucket, runnerPubKey, pair.PublicKey); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}
	if err := m.Bolt.Put(runnerBucket, runnerPrivKey, pair.PrivateKey); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}

	return &Runner{
		ID:         id,
		PublicKey:  pair.PublicKey,
		PrivateKey: pair.PrivateKey,
	}, nil
}

// GetIdentity loads the runner identity from the DB if it exists
func (m *Module) GetIdentity() (*Runner, error) {
	if m.Bolt == nil || m.Bolt.DB() == nil {
		return nil, fmt.Errorf("runner: bolt database is not open")
	}

	id, err := m.Bolt.Get(runnerBucket, runnerIDKey)
	if err != nil || id == nil {
		return nil, err
	}

	pub, err := m.Bolt.Get(runnerBucket, runnerPubKey)
	if err != nil || pub == nil {
		return nil, err
	}

	priv, err := m.Bolt.Get(runnerBucket, runnerPrivKey)
	if err != nil || priv == nil {
		return nil, err
	}

	return &Runner{
		ID:         string(id),
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}
