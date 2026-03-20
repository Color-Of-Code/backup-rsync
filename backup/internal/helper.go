// Package internal provides helper functions for internal use within the application.
package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

func getLogPath(simulate bool, configPath string, now time.Time) string {
	filename := filepath.Base(configPath)
	filename = strings.TrimSuffix(filename, ".yaml")
	logPath := "logs/sync-" + now.Format("2006-01-02T15-04-05") + "-" + filename

	if simulate {
		logPath += "-sim"
	}

	return logPath
}

func CreateMainLogger(configPath string, simulate bool, now time.Time) (*log.Logger, string, func() error, error) {
	logPath := getLogPath(simulate, configPath, now)
	overallLogPath := logPath + "/summary.log"

	err := os.MkdirAll(logPath, LogDirPermission)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	overallLogFile, err := os.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogFilePermission)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to open overall log file: %w", err)
	}

	logger := log.New(overallLogFile, "", log.LstdFlags)

	cleanup := func() error {
		return overallLogFile.Close()
	}

	return logger, logPath, cleanup, nil
}
