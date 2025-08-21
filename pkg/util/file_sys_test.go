// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testBaseDir(t *testing.T) string {
	dir := t.TempDir()
	_ = os.WriteFile(fmt.Sprintf("%s/.gitignore", dir), nil, 0644)
	return dir
}

func TestUnitFileExists(t *testing.T) {
	base := testBaseDir(t)
	t.Run("should return true for existing files", func(t *testing.T) {
		assert.True(t, FileExists(fmt.Sprintf("%s/.gitignore", base)))
	})
	t.Run("should return false for non-existing files", func(t *testing.T) {
		assert.False(t, FileExists(fmt.Sprintf("%s/not_found.md", base)))
	})
	t.Run("should return false for directories", func(t *testing.T) {
		assert.False(t, FileExists(base))
	})
}

func TestUnitDirExists(t *testing.T) {
	base := testBaseDir(t)
	t.Run("should return true for existing directories", func(t *testing.T) {
		assert.True(t, DirExists(base))
	})
	t.Run("should return false for non-existing directories", func(t *testing.T) {
		assert.False(t, DirExists(fmt.Sprintf("%s/not_found", base)))
	})
	t.Run("should return false for files", func(t *testing.T) {
		assert.False(t, DirExists(fmt.Sprintf("%s/.gitignore", base)))
	})
}

func TestUnitEnsureDir(t *testing.T) {
	base := testBaseDir(t)
	t.Run("should create new directory", func(t *testing.T) {
		newDir := fmt.Sprintf("%s/test_new_dir", base)
		_ = DeleteDir(newDir)
		assert.NoError(t, EnsureDir(newDir, 0755))
		assert.True(t, DirExists(newDir))
		assert.NoError(t, DeleteDir(newDir))
	})

	t.Run("should not error if directory already exists", func(t *testing.T) {
		assert.NoError(t, EnsureDir(base, 0755))
		assert.True(t, DirExists(base))
	})
}

func TestUnitDeleteDir(t *testing.T) {
	base := testBaseDir(t)
	t.Run("should delete existing directory", func(t *testing.T) {
		testDir := fmt.Sprintf("%s/test_delete_dir", base)
		assert.NoError(t, EnsureDir(testDir, 0755))
		assert.True(t, DirExists(testDir))
		assert.NoError(t, DeleteDir(testDir))
		assert.False(t, DirExists(testDir))
	})

	t.Run("should not error when deleting non-existing directory", func(t *testing.T) {
		nonExistingDir := fmt.Sprintf("%s/non_existing_dir", base)
		assert.NoError(t, DeleteDir(nonExistingDir))
	})
}
