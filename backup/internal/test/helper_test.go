package internal_test

import (
	"bytes"
	"log/slog"
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
	logPath := GetLogPath("title", fixedTime())

	logger, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), logPath)
	require.NoError(t, err)

	defer cleanup()

	assert.Contains(t, logPath, "title")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_DeterministicLogPath(t *testing.T) {
	logPath := GetLogPath("backup.yaml", fixedTime())

	_, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), logPath)
	require.NoError(t, err)

	defer cleanup()

	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-backup", logPath)
}

func TestCreateMainLogger_DeterministicLogPath_AnotherConfig(t *testing.T) {
	logPath := GetLogPath("sync.yaml", fixedTime())

	_, cleanup, err := CreateMainLogger(afero.NewMemMapFs(), logPath)
	require.NoError(t, err)

	defer cleanup()

	assert.Equal(t, "logs/sync-2025-06-15T14-30-45-sync", logPath)
}

func TestCreateMainLogger_MkdirError(t *testing.T) {
	// Use a read-only filesystem to block directory creation
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	logPath := GetLogPath("test.yaml", fixedTime())

	_, cleanup, err := CreateMainLogger(fs, logPath)
	_ = cleanup

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create log directory")
}

func TestCreateMainLogger_OpenFileError(t *testing.T) {
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	logPath := GetLogPath("test.yaml", fixedTime())

	_, cleanup, err := CreateMainLogger(fs, logPath)
	_ = cleanup

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create log directory")
}

func TestGetLogPath(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		expected   string
	}{
		{"WithYamlExtension", "backup.yaml", "logs/sync-2025-06-15T14-30-45-backup"},
		{"WithoutYamlExtension", "sync", "logs/sync-2025-06-15T14-30-45-sync"},
		{"WithDirectoryPrefix", "/etc/configs/media.yaml", "logs/sync-2025-06-15T14-30-45-media"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetLogPath(test.configPath, fixedTime())
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestNewUTCTextHandler_FormatsUTCTimestamp(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(NewUTCTextHandler(&buf))

	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, output)
}
