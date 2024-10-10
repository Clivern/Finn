// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGravatarURL(t *testing.T) {
	t.Run("returns URL for valid email", func(t *testing.T) {
		url := GravatarURL("user@example.com", 0)
		assert.True(t, strings.HasPrefix(url, "https://www.gravatar.com/avatar/"))
		assert.Contains(t, url, "?d=identicon")
		assert.NotEmpty(t, url)
	})

	t.Run("same email produces same hash", func(t *testing.T) {
		url1 := GravatarURL("test@example.com", 80)
		url2 := GravatarURL("test@example.com", 80)
		assert.Equal(t, url1, url2)
	})

	t.Run("email is normalized to lowercase", func(t *testing.T) {
		url1 := GravatarURL("User@Example.COM", 0)
		url2 := GravatarURL("user@example.com", 0)
		assert.Equal(t, url1, url2)
	})

	t.Run("includes size when in range", func(t *testing.T) {
		url := GravatarURL("u@e.com", 128)
		assert.Contains(t, url, "s=128")
	})

	t.Run("empty email returns empty string", func(t *testing.T) {
		assert.Empty(t, GravatarURL("", 0))
		assert.Empty(t, GravatarURL("   ", 0))
	})
}
