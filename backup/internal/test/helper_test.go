package internal_test

import (
	"os"
	"testing"
	"time"

	. "backup-rsync/backup/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to//normalize/", "/path/to/normalize"},
		{"/path//with//double/slashes/", "/path/with/double/slashes"},
		{"/trailing/slash/", "/trailing/slash"},
		{"/no/trailing/slash", "/no/trailing/slash"},
	}

	for _, test := range tests {
		result := NormalizePath(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func fixedTime() time.Time {
	return time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
}

func TestCreateMainLogger_Title_IsPresent(t *testing.T) {
	logger, logPath, cleanup, err := CreateMainLogger("title", true, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Contains(t, logPath, "title")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_IsSimulate_HasSimSuffix(t *testing.T) {
	logger, logPath, cleanup, err := CreateMainLogger("", true, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Contains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_NotSimulate_HasNoSimSuffix(t *testing.T) {
	logger, logPath, cleanup, err := CreateMainLogger("", false, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.NotContains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_DeterministicLogPath(t *testing.T) {
	_, logPath, cleanup, err := CreateMainLogger("backup.yaml", true, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-backup-sim", logPath)
}

func TestCreateMainLogger_DeterministicLogPath_NoSimulate(t *testing.T) {
	_, logPath, cleanup, err := CreateMainLogger("sync.yaml", false, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-sync", logPath)
}

func TestCreateMainLogger_MkdirError(t *testing.T) {
	// Use t.Chdir to a temp dir so we control the filesystem
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create "logs" as a regular file to block MkdirAll
	err := os.WriteFile("logs", []byte("block"), 0600)
	require.NoError(t, err)

	_, _, cleanup, err := CreateMainLogger("test.yaml", false, fixedTime())
	_ = cleanup

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create log directory")
}

func TestCreateMainLogger_OpenFileError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Pre-create the log path directory and make summary.log a directory to block OpenFile
	logDir := "logs/sync-2025-06-15T14-30-45-test"

	err := os.MkdirAll(logDir+"/summary.log", 0750)
	require.NoError(t, err)

	_, _, cleanup, err := CreateMainLogger("test.yaml", false, fixedTime())
	_ = cleanup

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open overall log file")
}
