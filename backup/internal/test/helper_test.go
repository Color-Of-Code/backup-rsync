package internal_test

import (
	"bytes"
	"errors"
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

func TestUTCLogWriter_FormatsISO8601UTC(t *testing.T) {
	var buf bytes.Buffer

	writer := &UTCLogWriter{
		W:   &buf,
		Now: fixedTime,
	}

	_, err := writer.Write([]byte("hello\n"))
	require.NoError(t, err)

	assert.Equal(t, "2025-06-15T14:30:45Z hello\n", buf.String())
}

func TestUTCLogWriter_ConvertsToUTC(t *testing.T) {
	var buf bytes.Buffer

	eastern := time.FixedZone("EST", -5*60*60)
	nonUTCTime := time.Date(2025, 6, 15, 10, 0, 0, 0, eastern)

	writer := &UTCLogWriter{
		W: &buf,
		Now: func() time.Time {
			return nonUTCTime
		},
	}

	_, err := writer.Write([]byte("test\n"))
	require.NoError(t, err)

	assert.Equal(t, "2025-06-15T15:00:00Z test\n", buf.String())
}

var errWriteFailed = errors.New("write failed")

type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, errWriteFailed
}

func TestUTCLogWriter_PropagatesWriteError(t *testing.T) {
	writer := &UTCLogWriter{
		W:   &failWriter{},
		Now: fixedTime,
	}

	_, err := writer.Write([]byte("hello\n"))

	require.ErrorIs(t, err, errWriteFailed)
}

func TestNewUTCLogger_WritesISO8601(t *testing.T) {
	var buf bytes.Buffer

	logger := NewUTCLogger(&buf)

	logger.Print("test message")

	output := buf.String()
	// Should contain ISO 8601 timestamp format (RFC3339) and the message
	assert.Contains(t, output, "test message")
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z `, output)
}
