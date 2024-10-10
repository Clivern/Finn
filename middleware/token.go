// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/clivern/finn/service/runner"
	"github.com/clivern/finn/service/util"
)

const runnerTokenHeader = "X-Runner-Token"

// RunnerToken verifies the request has header X-Runner-Token matching the token from runner.Deps
func RunnerToken(deps *runner.Deps) func(http.Handler) http.Handler {
	token := ""
	if deps != nil {
		token = deps.Token
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": "invalid or missing X-Runner-Token",
				})
				return
			}
			got := r.Header.Get(runnerTokenHeader)
			if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": "invalid or missing X-Runner-Token",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
