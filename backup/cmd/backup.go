package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func executeSyncJobs(cfg internal.Config, simulate bool) {
	logPath := fmt.Sprintf("logs/sync-%s", time.Now().Format("2006-01-02T15-04-05"))

	if err := os.MkdirAll(logPath, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	overallLogPath := fmt.Sprintf("%s/summary.log", logPath)
	overallLogFile, err := os.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open overall log file: %v", err)
	}
	defer overallLogFile.Close()
	overallLogger := log.New(overallLogFile, "", log.LstdFlags)

	for _, job := range cfg.Jobs {
		jobLogPath := fmt.Sprintf("%s/job-%s.log", logPath, job.Name)
		logFile, err := os.OpenFile(jobLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			overallLogger.Printf("ERROR [%s]: Failed to create job log file: %v", job.Name, err)
			continue
		}
		defer logFile.Close()
		jobLogger := log.New(logFile, "", log.LstdFlags)

		status := internal.ExecuteJob(job, simulate, jobLogger)
		overallLogger.Printf("STATUS [%s]: %s", job.Name, status)
		fmt.Printf("Status [%s]: %s\n", job.Name, status)
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the sync jobs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadResolvedConfig(configPath)
		executeSyncJobs(cfg, false)
	},
}

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate the sync jobs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadResolvedConfig(configPath)
		executeSyncJobs(cfg, true)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(simulateCmd)
}
