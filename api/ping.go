// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/clivern/finn/locale"
	"github.com/clivern/finn/service/runner"
	"github.com/clivern/finn/service/util"

	"github.com/rs/zerolog/log"
)

// NewPingHandler returns a handler that verifies the ping payload using deps.Encrypter.
func NewPingHandler(deps *runner.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"errorMessage": locale.TR(r, "methodNotAllowed"),
			})
			return
		}

		if deps == nil || deps.Encrypter == nil {
			util.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"errorMessage": locale.TR(r, "runnerIdentityUnavailable"),
			})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "failedToReadBody"),
			})
			return
		}

		ciphertext, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "bodyMustBeBase64Ciphertext"),
			})
			return
		}

		plaintext, err := deps.Encrypter.DecryptString(ciphertext)
		if err != nil {
			log.Info().Err(err).Msg("Ping: decrypt failed")
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "decryptionFailed"),
			})
			return
		}

		if plaintext != "PONG" {
			log.Info().Str("plaintext", plaintext).Msg("Ping: payload must decrypt to PONG")
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": locale.TR(r, "payloadMustDecryptToPONG"),
			})
			return
		}

		log.Info().Msg("Ping: PONG verified successfully")
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	}
}
