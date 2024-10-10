// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"time"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/service/util"
)

// Setup handles the initial installation and configuration of the gateway.
type Setup struct {
	OptionRepository      db.OptionRepository
	UserRepository        db.UserRepository
	SystemEventRepository db.SystemEventRepository
}

// SetupOptions contains the configuration options for gateway setup.
type SetupOptions struct {
	PlatformURL   string
	PlatformEmail string
	PlatformName  string
	AdminEmail    string
	AdminPassword string
}

// NewSetup creates a new Setup instance with the provided repositories.
func NewSetup(optionRepository db.OptionRepository, userRepository db.UserRepository, systemEventRepository db.SystemEventRepository) *Setup {
	return &Setup{
		OptionRepository:      optionRepository,
		UserRepository:        userRepository,
		SystemEventRepository: systemEventRepository,
	}
}

// IsInstalled checks whether the platform has been installed.
func (s *Setup) IsInstalled() bool {
	option, err := s.OptionRepository.Get("is_installed")
	if err != nil {
		return false
	}
	return option != nil
}

// Install performs the initial platform installation with the provided options.
func (s *Setup) Install(options *SetupOptions) error {
	if s.IsInstalled() {
		return errors.New("platform is already installed")
	}

	hashedPassword, err := util.HashPassword(options.AdminPassword)
	if err != nil {
		return err
	}
	user := &db.User{
		Name:        "",
		Email:       options.AdminEmail,
		Password:    hashedPassword,
		Role:        db.UserRoleAdmin,
		IsActive:    true,
		LastLoginAt: time.Now().UTC(),
	}
	err = s.UserRepository.Create(user)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("is_installed", "1")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("platform_url", options.PlatformURL)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("platform_email", options.PlatformEmail)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("platform_name", options.PlatformName)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create(
		"platform_description",
		"A Platform to Run, Schedule, and Automate Ansible and Terraform Workflows.",
	)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("maintenance_mode", "0")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("smtp_server", "smtp.gmail.com")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("smtp_port", "587")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("smtp_from_email", "noreply@finn.io")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("smtp_username", "")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("smtp_password", "")
	if err != nil {
		return err
	}

	err = s.OptionRepository.Create("smtp_use_tls", "1")
	if err != nil {
		return err
	}

	s.SystemEventRepository.Create("platform_installed", &db.SystemEventPlatformInstalled{
		UserID:    user.ID.String(),
		UserEmail: user.Email,
	})

	return nil
}
