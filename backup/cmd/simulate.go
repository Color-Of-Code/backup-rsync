package cmd

import (
	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func buildSimulateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "simulate",
		Short: "Simulate the sync jobs",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			cfg := internal.LoadResolvedConfig(configPath)
			command := internal.NewRSyncSimulateCommand(rsyncPath)

			cfg.Apply(command)
		},
	}
}
