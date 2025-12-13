package internal_test

import (
	"testing"

	. "backup-rsync/backup/internal"

	"github.com/stretchr/testify/assert"
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

func TestCreateMainLogger_Title_IsPresent(t *testing.T) {
	logger, logPath := CreateMainLogger("title", true)
	assert.Contains(t, logPath, "title")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_IsSimulate_HasSimSuffix(t *testing.T) {
	logger, logPath := CreateMainLogger("", true)
	assert.Contains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateMainLogger_NotSimulate_HasNoSimSuffix(t *testing.T) {
	logger, logPath := CreateMainLogger("", false)
	assert.NotContains(t, logPath, "-sim")
	assert.NotNil(t, logger)
}

func TestCreateLogPath_IsSimulate_ContainsTimestamp(t *testing.T) {
	_, logPath := CreateMainLogger("", true)
	// Check if the logPath contains a timestamp in the format '2006-01-02T15-04-05'
	timestampRegex := `\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}`
	assert.Regexp(t, timestampRegex, logPath)
}

func TestCreateLogPath_NotSimulate_ContainsTimestamp(t *testing.T) {
	_, logPath := CreateMainLogger("", false)
	// Check if the logPath contains a timestamp in the format '2006-01-02T15-04-05'
	timestampRegex := `\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}`
	assert.Regexp(t, timestampRegex, logPath)
}
