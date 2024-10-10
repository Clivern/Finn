// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/module"
	"github.com/clivern/finn/service/util"

	"github.com/rs/zerolog/log"
)

// SetupRequest represents the setup request payload
type SetupRequest struct {
	PlatformURL   string `json:"platformURL" validate:"required,url,min=4,max=60" label:"Platform URL"`
	PlatformEmail string `json:"platformEmail" validate:"required,email,min=4,max=60" label:"Platform Email"`
	PlatformName  string `json:"platformName" validate:"required,min=2,max=50" label:"Platform Name"`
	AdminEmail    string `json:"adminEmail" validate:"required,email,min=4,max=60" label:"Admin Email"`
	AdminPassword string `json:"adminPassword" validate:"required,strong_password,min=8,max=60" label:"Admin Password"`
}

// SetupAction handles the setup installation
func SetupAction(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest

	if err := util.DecodeAndValidate(r, &req); err != nil {
		util.WriteValidationError(w, err)
		return
	}

	setupModule := module.NewSetup(
		db.NewOptionRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
		db.NewSystemEventRepository(db.GetDB()),
	)

	if setupModule.IsInstalled() {
		util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": locale.TR(r, "platformAlreadyInstalled"),
		})
		return
	}

	err := setupModule.Install(&module.SetupOptions{
		PlatformURL:   req.PlatformURL,
		PlatformEmail: req.PlatformEmail,
		PlatformName:  req.PlatformName,
		AdminEmail:    req.AdminEmail,
		AdminPassword: req.AdminPassword,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to complete setup")
		util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": locale.TR(r, "failedCompleteSetup"),
		})
		return
	}

	log.Info().Msg("Platform setup completed successfully")
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": locale.TR(r, "setupCompletedSuccessfully"),
	})
}

// SetupStatusAction checks if the platform is already installed
func SetupStatusAction(w http.ResponseWriter, _ *http.Request) {
	setupModule := module.NewSetup(
		db.NewOptionRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
		db.NewSystemEventRepository(db.GetDB()),
	)
	util.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"installed": setupModule.IsInstalled(),
	})
}
