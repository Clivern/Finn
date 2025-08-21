// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import "fmt"

// Deps holds dependencies for runner HTTP handlers
type Deps struct {
	Bolt   *Bolt
	Token  string
	Worker *Worker
}

// NewDeps loads runner identity from bolt and builds deps (token used for request signing).
func NewDeps(bolt *Bolt) (*Deps, error) {
	if bolt == nil {
		return nil, fmt.Errorf("runner deps: bolt is required")
	}

	mod := &Module{Bolt: bolt}
	identity, err := mod.GetIdentity()
	if err != nil || identity == nil {
		return &Deps{Bolt: bolt, Worker: NewWorker(bolt)}, nil
	}

	return &Deps{
		Bolt:   bolt,
		Token:  identity.Token,
		Worker: NewWorker(bolt),
	}, nil
}
