// Package internal provides helper functions for internal use within the application.
package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// UTCLogWriter wraps an io.Writer and prepends an ISO 8601 UTC timestamp to each write.
type UTCLogWriter struct {
	W   io.Writer
	Now func() time.Time
}

func (u *UTCLogWriter) Write(data []byte) (int, error) {
	now := u.Now().UTC().Format(time.RFC3339)

	_, err := fmt.Fprintf(u.W, "%s %s", now, data)
	if err != nil {
		return 0, fmt.Errorf("writing log entry: %w", err)
	}

	return len(data), nil
}

// NewUTCLogger creates a *log.Logger that writes ISO 8601 UTC timestamps.
func NewUTCLogger(w io.Writer) *log.Logger {
	return log.New(&UTCLogWriter{W: w, Now: time.Now}, "", 0)
}

// Path represents a source or target path with optional exclusions.
type Path struct {
	Path       string   `yaml:"path"`
	Exclusions []string `yaml:"exclusions"`
}

func NormalizePath(path string) string {
	return strings.TrimSuffix(strings.ReplaceAll(path, "//", "/"), "/")
}

const LogFilePermission = 0644
const LogDirPermission = 0755

func GetLogPath(configPath string, now time.Time) string {
	filename := filepath.Base(configPath)
	filename = strings.TrimSuffix(filename, ".yaml")

	return "logs/sync-" + now.Format("2006-01-02T15-04-05") + "-" + filename
}

func CreateMainLogger(
	fs afero.Fs, logPath string,
) (*log.Logger, func() error, error) {
	overallLogPath := logPath + "/summary.log"

	err := fs.MkdirAll(logPath, LogDirPermission)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	overallLogFile, err := fs.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogFilePermission)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open overall log file: %w", err)
	}

	logger := NewUTCLogger(overallLogFile)

	cleanup := func() error {
		return overallLogFile.Close()
	}

	return logger, cleanup, nil
}
