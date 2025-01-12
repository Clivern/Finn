// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/clivern/finn/runner"
	"github.com/clivern/finn/service"

	"github.com/rs/zerolog/log"
)

// NewPingHandler returns a handler that verifies the ping payload using deps.Encrypter.
func NewPingHandler(deps *runner.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			service.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"errorMessage": "method not allowed",
			})
			return
		}

		if deps == nil || deps.Encrypter == nil {
			service.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"errorMessage": "runner identity unavailable",
			})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "failed to read body",
			})
			return
		}

		ciphertext, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "body must be base64-encoded ciphertext",
			})
			return
		}

		plaintext, err := deps.Encrypter.DecryptString(ciphertext)
		if err != nil {
			log.Info().Err(err).Msg("Ping: decrypt failed")
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "decryption failed",
			})
			return
		}

		if plaintext != "PONG" {
			log.Info().Str("plaintext", plaintext).Msg("Ping: payload must decrypt to PONG")
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "payload must decrypt to PONG",
			})
			return
		}

		log.Info().Msg("Ping: PONG verified successfully")
		service.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	}
}
