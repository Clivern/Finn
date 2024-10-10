// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import "fmt"

// Deps holds dependencies for runner HTTP handlers
type Deps struct {
	Bolt      *Bolt
	Encrypter *Encrypter
	Token     string
}

// NewDeps loads runner identity from bolt, builds an encrypter
func NewDeps(bolt *Bolt) (*Deps, error) {
	if bolt == nil {
		return nil, fmt.Errorf("runner deps: bolt is required")
	}

	mod := &Module{Bolt: bolt, Key: &Key{}}
	identity, err := mod.GetIdentity()
	if err != nil || identity == nil {
		return &Deps{Bolt: bolt}, nil
	}

	enc, err := NewEncrypter(identity)
	if err != nil {
		return &Deps{Bolt: bolt, Token: identity.ID}, nil
	}

	return &Deps{
		Bolt:      bolt,
		Encrypter: enc,
		Token:     identity.ID,
	}, nil
}
