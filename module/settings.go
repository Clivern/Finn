// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/clivern/finn/db"
)

// Settings manages app settings (platform URL, SMTP, etc.).
type Settings struct {
	OptionRepository db.OptionRepository
}

// NewSettings creates a settings module with the given option repo.
func NewSettings(optionRepository db.OptionRepository) *Settings {
	return &Settings{OptionRepository: optionRepository}
}

// UpdateSettingsOptions is what you pass when updating settings.
type UpdateSettingsOptions struct {
	PlatformName    string
	PlatformURL     string
	PlatformEmail   string
	SMTPServer      string
	SMTPPort        int
	SMTPFromEmail   string
	SMTPUsername    string
	SMTPPassword    string
	SMTPUseTLS      bool
	MaintenanceMode bool
}

// Update persists the provided settings to the options repository.
func (s *Settings) Update(options *UpdateSettingsOptions) error {
	updates := map[string]string{
		"platform_name":    strings.TrimSpace(options.PlatformName),
		"platform_url":     strings.TrimSpace(options.PlatformURL),
		"platform_email":   strings.TrimSpace(options.PlatformEmail),
		"smtp_server":      strings.TrimSpace(options.SMTPServer),
		"smtp_port":        strconv.Itoa(options.SMTPPort),
		"smtp_from_email":  strings.TrimSpace(options.SMTPFromEmail),
		"smtp_username":    strings.TrimSpace(options.SMTPUsername),
		"smtp_password":    options.SMTPPassword,
		"smtp_use_tls":     boolToOption(options.SMTPUseTLS),
		"maintenance_mode": boolToOption(options.MaintenanceMode),
	}

	for key, value := range updates {
		if err := s.OptionRepository.Update(key, value); err != nil {
			return fmt.Errorf("failed to update option %s: %w", key, err)
		}
	}

	return nil
}

func boolToOption(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

// GetSettings retrieves the current settings from the options repository.
func (s *Settings) GetSettings() (map[string]string, error) {
	keys := []string{
		"platform_name",
		"platform_url",
		"platform_email",
		"maintenance_mode",
		"smtp_server",
		"smtp_port",
		"smtp_from_email",
		"smtp_username",
		"smtp_password",
		"smtp_use_tls",
	}

	settings := make(map[string]string, len(keys))

	for _, key := range keys {
		option, err := s.OptionRepository.Get(key)
		if err != nil {
			return nil, err
		}
		if option != nil {
			settings[key] = option.Value
		} else {
			settings[key] = ""
		}
	}

	return settings, nil
}
