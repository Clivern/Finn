// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/pkg/util"

	"github.com/rs/zerolog/log"
)

// UpdateSettingsRequest is the body for updating app settings.
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

// UpdateSettingsAction updates app settings (admin only).
func UpdateSettingsAction(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUserFromContext(r.Context())

	log.Info().
		Str("userID", user.ID.String()).
		Msg("Update settings endpoint called")

	var req UpdateSettingsRequest
	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
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
		log.Error().
			Err(err).
			Str("userID", user.ID.String()).
			Msg("Failed to update settings")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedUpdateSettings"),
		})
		return
	}

	log.Info().
		Str("userID", user.ID.String()).
		Msg("Settings updated successfully")

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": locale.TR(r, "settingsUpdatedSuccessfully"),
	})
}

// GetSettingsAction returns current app settings.
func GetSettingsAction(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUserFromContext(r.Context())

	log.Info().
		Str("userID", user.ID.String()).
		Msg("Get settings endpoint called")

	settingsModule := module.NewSettings(db.NewOptionRepository(db.GetDB()))
	settings, err := settingsModule.GetSettings()

	if err != nil {
		log.Error().
			Err(err).
			Str("userID", user.ID.String()).
			Msg("Failed to get settings")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedGetSettings"),
		})
		return
	}

	log.Info().
		Str("userID", user.ID.String()).
		Msg("Settings retrieved successfully")

	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"settings": settings,
	})
}
