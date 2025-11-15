package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

const filePermission = 0644
const logDirPermission = 0755

func getLogPath(create bool) string {
	logPath := "logs/sync-" + time.Now().Format("2006-01-02T15-04-05")
	if create {
		err := os.MkdirAll(logPath, logDirPermission)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	return logPath
}

func executeSyncJobs(cfg internal.Config, simulate bool) {
	logPath := getLogPath(true)

	overallLogPath := logPath + "/summary.log"

	overallLogFile, err := os.OpenFile(overallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
	if err != nil {
		log.Fatalf("Failed to open overall log file: %v", err)
	}
	defer overallLogFile.Close()

	overallLogger := log.New(overallLogFile, "", log.LstdFlags)

	for _, job := range cfg.Jobs {
		jobLogPath := fmt.Sprintf("%s/job-%s.log", logPath, job.Name)
		status := internal.ExecuteJob(job, simulate, false, jobLogPath)
		overallLogger.Printf("STATUS [%s]: %s", job.Name, status)
		fmt.Printf("Status [%s]: %s\n", job.Name, status)
	}
}

func listCommands(cfg internal.Config) {
	logPath := getLogPath(false)
	for _, job := range cfg.Jobs {
		jobLogPath := fmt.Sprintf("%s/job-%s.log", logPath, job.Name)
		internal.ExecuteJob(job, false, true, jobLogPath)
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the sync jobs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := internal.LoadResolvedConfig(configPath)
		executeSyncJobs(cfg, false)
	},
}

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate the sync jobs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := internal.LoadResolvedConfig(configPath)
		executeSyncJobs(cfg, true)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the commands that will be executed",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := internal.LoadResolvedConfig(configPath)
		listCommands(cfg)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(simulateCmd)
	RootCmd.AddCommand(listCmd)
}
