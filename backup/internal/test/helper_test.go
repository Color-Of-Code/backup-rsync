package internal_test

import (
	"testing"
	"time"

	. "backup-rsync/backup/internal"

	"github.com/spf13/afero"
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
	logger, logPath, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), "title", true, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Contains(t, logPath, "title")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_IsSimulate_HasSimSuffix(t *testing.T) {
	logger, logPath, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), "", true, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Contains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_NotSimulate_HasNoSimSuffix(t *testing.T) {
	logger, logPath, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), "", false, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.NotContains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_DeterministicLogPath(t *testing.T) {
	_, logPath, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), "backup.yaml", true, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-backup-sim", logPath)
}

func TestCreateMainLogger_DeterministicLogPath_NoSimulate(t *testing.T) {
	_, logPath, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), "sync.yaml", false, fixedTime())
	require.NoError(t, err)

	defer cleanup()

	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-sync", logPath)
}

func TestCreateMainLogger_MkdirError(t *testing.T) {
	// Use a read-only filesystem to block directory creation
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())

	_, _, cleanup, err := CreateMainLogger(fs, "test.yaml", false, fixedTime())
	_ = cleanup

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create log directory")
}

func TestCreateMainLogger_OpenFileError(t *testing.T) {
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())

	_, _, cleanup, err := CreateMainLogger(fs, "test.yaml", false, fixedTime())
	_ = cleanup

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create log directory")
}
