// Package internal provides helper functions for internal use within the application.
package internal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// NewUTCTextHandler creates a slog.Handler that writes text logs with UTC timestamps.
func NewUTCTextHandler(w io.Writer) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339))
			}

			return a
		},
	})
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
) (*slog.Logger, func() error, error) {
	overallLogPath := logPath + "/summary.log"

	err := fs.MkdirAll(logPath, LogDirPermission)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	overallLogFile, err := fs.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogFilePermission)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open overall log file: %w", err)
	}

	logger := slog.New(NewUTCTextHandler(overallLogFile))

	cleanup := func() error {
		return overallLogFile.Close()
	}

	return logger, cleanup, nil
}
