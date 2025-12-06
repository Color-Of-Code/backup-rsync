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
	assert.NotNil(t, logger, "Logger should not be nil")
}

func TestCreateMainLogger_IsSimulate_HasSimSuffix(t *testing.T) {
	logger, logPath := CreateMainLogger("", true)
	assert.Contains(t, logPath, "-sim")
	assert.NotNil(t, logger, "Logger should not be nil")
}

func TestCreateMainLogger_NotSimulate_HasNoSimSuffix(t *testing.T) {
	logger, logPath := CreateMainLogger("", false)
	assert.NotContains(t, logPath, "-sim")
	assert.NotNil(t, logger, "Logger should not be nil")
}
