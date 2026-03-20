package internal_test

import (
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
	logger, logPath, err := CreateMainLogger("title", true, fixedTime())
	require.NoError(t, err)
	assert.Contains(t, logPath, "title")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_IsSimulate_HasSimSuffix(t *testing.T) {
	logger, logPath, err := CreateMainLogger("", true, fixedTime())
	require.NoError(t, err)
	assert.Contains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_NotSimulate_HasNoSimSuffix(t *testing.T) {
	logger, logPath, err := CreateMainLogger("", false, fixedTime())
	require.NoError(t, err)
	assert.NotContains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_DeterministicLogPath(t *testing.T) {
	_, logPath, err := CreateMainLogger("backup.yaml", true, fixedTime())
	require.NoError(t, err)
	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-backup-sim", logPath)
}

func TestCreateMainLogger_DeterministicLogPath_NoSimulate(t *testing.T) {
	_, logPath, err := CreateMainLogger("sync.yaml", false, fixedTime())
	require.NoError(t, err)
	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-sync", logPath)
}
