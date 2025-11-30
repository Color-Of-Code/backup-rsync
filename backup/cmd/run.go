package cmd

import (
	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func buildRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Execute the sync jobs",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			cfg := internal.LoadResolvedConfig(configPath)
			command := internal.NewRSyncCommand(rsyncPath)

			cfg.Apply(command)
		},
	}
}
