package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"backup-rsync/backup/cmd"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)

	return path
}

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	return executeCommandWithFs(t, afero.NewMemMapFs(), args...)
}

func executeCommandWithFs(t *testing.T, fs afero.Fs, args ...string) (string, error) {
	t.Helper()

	rootCmd := cmd.BuildRootCommandWithFs(fs)

	var stdout bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	return stdout.String(), err
}

// --- config show ---

func TestConfigShow_ValidConfig(t *testing.T) {
	cfgPath := writeConfigFile(t, `
sources:
  - path: "/home"
targets:
  - path: "/backup"
jobs:
  - name: "docs"
    source: "/home/docs/"
    target: "/backup/docs/"
`)

	stdout, err := executeCommand(t, "config", "show", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "docs")
	assert.Contains(t, stdout, "/home/docs")
	assert.Contains(t, stdout, "/backup/docs")
}

func TestConfigShow_MissingFile(t *testing.T) {
	_, err := executeCommand(t, "config", "show", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

func TestConfigShow_InvalidYAML(t *testing.T) {
	cfgPath := writeConfigFile(t, `{{{invalid yaml`)

	_, err := executeCommand(t, "config", "show", "--config", cfgPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

// --- config validate ---

func TestConfigValidate_ValidConfig(t *testing.T) {
	cfgPath := writeConfigFile(t, `
sources:
  - path: "/home"
targets:
  - path: "/backup"
jobs:
  - name: "docs"
    source: "/home/docs/"
    target: "/backup/docs/"
`)

	stdout, err := executeCommand(t, "config", "validate", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Configuration is valid.")
}

func TestConfigValidate_MissingFile(t *testing.T) {
	_, err := executeCommand(t, "config", "validate", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validating config")
}

func TestConfigValidate_DuplicateJobNames(t *testing.T) {
	cfgPath := writeConfigFile(t, `
sources:
  - path: "/home"
targets:
  - path: "/backup"
jobs:
  - name: "same"
    source: "/home/a/"
    target: "/backup/a/"
  - name: "same"
    source: "/home/b/"
    target: "/backup/b/"
`)

	_, err := executeCommand(t, "config", "validate", "--config", cfgPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validating config")
}

// --- list ---

func TestList_MissingConfig(t *testing.T) {
	_, err := executeCommand(t, "list", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

// --- run ---

func TestRun_MissingConfig(t *testing.T) {
	_, err := executeCommand(t, "run", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

// --- simulate ---

func TestSimulate_MissingConfig(t *testing.T) {
	_, err := executeCommand(t, "simulate", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

// --- version ---

func TestVersion_InvalidRsyncPath(t *testing.T) {
	_, err := executeCommand(t, "version", "--rsync-path", "not-absolute")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting version info")
}

func TestVersion_NonExistentRsyncPath(t *testing.T) {
	_, err := executeCommand(t, "version", "--rsync-path", "/nonexistent/rsync")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting version info")
}

// --- check-coverage ---

func TestCheckCoverage_MissingConfig(t *testing.T) {
	_, err := executeCommand(t, "check-coverage", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

func TestCheckCoverage_ValidConfig(t *testing.T) {
	cfgPath := writeConfigFile(t, `
sources:
  - path: "/src"
targets:
  - path: "/dst"
jobs:
  - name: "docs"
    source: "/src/docs/"
    target: "/dst/docs/"
`)

	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/src/docs", 0755)

	stdout, err := executeCommandWithFs(t, fs, "check-coverage", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Uncovered paths:")
}

// --- version (positive path) ---

func TestVersion_ValidRsync(t *testing.T) {
	// Only run if rsync is actually installed
	_, err := os.Stat("/usr/bin/rsync")
	if os.IsNotExist(err) {
		t.Skip("rsync not installed")
	}

	stdout, err := executeCommand(t, "version", "--rsync-path", "/usr/bin/rsync")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Rsync Binary Path: /usr/bin/rsync")
	assert.Contains(t, stdout, "Version Info:")
}

// --- list (positive path) ---

func TestList_ValidConfig(t *testing.T) {
	cfgPath := writeConfigFile(t, `
sources:
  - path: "/home"
targets:
  - path: "/backup"
jobs:
  - name: "docs"
    source: "/home/docs/"
    target: "/backup/docs/"
    enabled: true
`)

	stdout, err := executeCommand(t, "list", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: docs")
	assert.Contains(t, stdout, "Status [docs]: SUCCESS")
}
