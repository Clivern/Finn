// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/pkg/util"
)

// Auth does login and password checks.
type Auth struct {
	UserRepository db.UserRepository
}

// NewAuth creates a new auth.
func NewAuth(repo db.UserRepository) *Auth {
	return &Auth{UserRepository: repo}
}

// Login authenticates a user.
func (a *Auth) Login(email, password string) (*db.User, error) {
	user, err := a.UserRepository.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("email not found")
	}
	if !util.ComparePassword(user.Password, password) {
		return nil, errors.New("invalid password")
	}

	a.UserRepository.UpdateLastLogin(user.ID)

	return user, nil
}
