// Package testutil provides shared test helpers for use across test packages.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"backup-rsync/backup/internal"
)

// WriteConfigFile writes YAML content to a temp file and returns its path.
func WriteConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	err := os.WriteFile(path, []byte(content), internal.LogFilePermission)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	return path
}
