// Package internal provides helper functions for internal use within the application.
package internal

import (
	"fmt"
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

func GetLogPath(create bool) string {
	logPath := "logs/sync-" + time.Now().Format("2006-01-02T15-04-05")
	if create {
		err := os.MkdirAll(logPath, LogDirPermission)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	return logPath
}

func ExecuteSyncJobs(cfg Config, simulate bool) {
	logPath := GetLogPath(true)

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

	overallLogger := log.New(overallLogFile, "", log.LstdFlags)

	for _, job := range cfg.Jobs {
		jobLogPath := fmt.Sprintf("%s/job-%s.log", logPath, job.Name)
		status := ExecuteJob(job, simulate, false, jobLogPath)
		overallLogger.Printf("STATUS [%s]: %s", job.Name, status)
		fmt.Printf("Status [%s]: %s\n", job.Name, status)
	}
}
