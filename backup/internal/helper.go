// Package internal provides helper functions for internal use within the application.
package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func NormalizePath(path string) string {
	return strings.TrimSuffix(strings.ReplaceAll(path, "//", "/"), "/")
}

const FilePermission = 0644
const LogDirPermission = 0755

func GetLogPath() string {
	logPath := "logs/sync-" + time.Now().Format("2006-01-02T15-04-05")

	err := os.MkdirAll(logPath, LogDirPermission)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	return logPath
}

func createFileLogger() (*log.Logger, string) {
	logPath := GetLogPath()

	overallLogPath := logPath + "/summary.log"

	overallLogFile, err := os.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, FilePermission)
	if err != nil {
		log.Fatalf("Failed to open overall log file: %v", err)
	}

	defer func() {
		err := overallLogFile.Close()
		if err != nil {
			log.Fatalf("Failed to close overall log file: %v", err)
		}
	}()

	logger := log.New(overallLogFile, "", log.LstdFlags)

	return logger, logPath
}

func createLogger(rsync RSyncCommand) (*log.Logger, string) {
	if rsync.ListOnly {
		return log.New(io.Discard, "", 0), ""
	}

	return createFileLogger()
}

func (cfg Config) Apply(rsync RSyncCommand) {
	overallLogger, logPath := createLogger(rsync)

	versionInfo, err := rsync.GetVersionInfo()
	if err != nil {
		overallLogger.Printf("Failed to fetch rsync version: %v", err)
	} else {
		overallLogger.Printf("Rsync Binary Path: %s", rsync.BinPath)
		overallLogger.Printf("Rsync Version Info: %s", versionInfo)
	}

	for _, job := range cfg.Jobs {
		jobLogPath := fmt.Sprintf("%s/job-%s.log", logPath, job.Name)
		status := job.Apply(rsync, jobLogPath)
		overallLogger.Printf("STATUS [%s]: %s", job.Name, status)
		fmt.Printf("Status [%s]: %s\n", job.Name, status)
	}
}
