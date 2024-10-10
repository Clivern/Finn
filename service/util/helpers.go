// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// WriteJSON writes a JSON response with the given status code and data
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON response: %w", err)
	}
	return nil
}

// CalculateDataChecksum calculates the SHA256 checksum of the given data
func CalculateDataChecksum(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash), nil
}

// GenerateRandomSlug generates a random slug of length 10–14 characters.
func GenerateRandomSlug() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	n := make([]byte, 1)
	_, _ = rand.Read(n)
	length := 10 + int(n[0])%5 // 10 to 14 inclusive
	b := make([]byte, length)
	_, _ = rand.Read(b)
	var s strings.Builder
	s.Grow(length)
	for i := 0; i < length; i++ {
		s.WriteByte(chars[int(b[i])%len(chars)])
	}
	return s.String()
}
