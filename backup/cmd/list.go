package cmd

import (
	"fmt"

	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func listCommands(cfg internal.Config) {
	logPath := internal.GetLogPath(false)
	for _, job := range cfg.Jobs {
		jobLogPath := fmt.Sprintf("%s/job-%s.log", logPath, job.Name)
		internal.ExecuteJob(job, false, true, jobLogPath)
	}
}

func buildListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the commands that will be executed",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			cfg := internal.LoadResolvedConfig(configPath)
			listCommands(cfg)
		},
	}
}
