// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/met/db"
	"github.com/clivern/met/module"
	"github.com/clivern/met/service"

	"github.com/rs/zerolog/log"
)

// UpdateSettingsRequest represents the payload for updating application settings.
type UpdateSettingsRequest struct {
	PlatformName    string `json:"platformName" validate:"required,max=100" label:"Platform Name"`
	PlatformURL     string `json:"platformUrl" validate:"required,url,max=255" label:"Platform URL"`
	PlatformEmail   string `json:"platformEmail" validate:"required,email,max=255" label:"Platform Email"`
	SMTPServer      string `json:"smtpServer" validate:"omitempty,max=255" label:"SMTP Server"`
	SMTPPort        int    `json:"smtpPort" validate:"omitempty,gte=1,lte=65535" label:"SMTP Port"`
	SMTPFromEmail   string `json:"smtpFromEmail" validate:"omitempty,email,max=255" label:"SMTP From Email"`
	SMTPUsername    string `json:"smtpUsername" validate:"omitempty,max=255" label:"SMTP Username"`
	SMTPPassword    string `json:"smtpPassword" validate:"omitempty,max=255" label:"SMTP Password"`
	SMTPUseTLS      bool   `json:"smtpUseTLS" label:"SMTP Use TLS"`
	MaintenanceMode bool   `json:"maintenanceMode" label:"Maintenance Mode"`
}

// UpdateSettingsAction handles application settings update requests
func UpdateSettingsAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update settings endpoint called")

	var req UpdateSettingsRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	settingsModule := module.NewSettings(db.NewOptionRepository(db.GetDB()))

	err := settingsModule.Update(&module.UpdateSettingsOptions{
		PlatformName:    req.PlatformName,
		PlatformURL:     req.PlatformURL,
		PlatformEmail:   req.PlatformEmail,
		SMTPServer:      req.SMTPServer,
		SMTPPort:        req.SMTPPort,
		SMTPFromEmail:   req.SMTPFromEmail,
		SMTPUsername:    req.SMTPUsername,
		SMTPPassword:    req.SMTPPassword,
		SMTPUseTLS:      req.SMTPUseTLS,
		MaintenanceMode: req.MaintenanceMode,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to update settings")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update settings",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "Settings updated successfully",
	})
}

// GetSettingsAction handles user settings get requests
func GetSettingsAction(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msg("Get settings endpoint called")

	settingsModule := module.NewSettings(db.NewOptionRepository(db.GetDB()))
	settings, err := settingsModule.GetSettings()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get settings")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get settings",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"settings": settings,
	})
}
