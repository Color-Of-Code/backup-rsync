// Package internal provides helper functions for internal use within the application.
package internal

import (
	"log"
	"os"
	"strings"
	"time"
)

func NormalizePath(path string) string {
	return strings.TrimSuffix(strings.ReplaceAll(path, "//", "/"), "/")
}

const LogFilePermission = 0644
const LogDirPermission = 0755

func getLogPath() string {
	logPath := "logs/sync-" + time.Now().Format("2006-01-02T15-04-05")

	err := os.MkdirAll(logPath, LogDirPermission)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	return logPath
}

func createFileLogger() (*log.Logger, string) {
	logPath := getLogPath()

	overallLogPath := logPath + "/summary.log"

	overallLogFile, err := os.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogFilePermission)
	if err != nil {
		log.Fatalf("Failed to open overall log file: %v", err)
	}

	logger := log.New(overallLogFile, "", log.LstdFlags)

	return logger, logPath
}
