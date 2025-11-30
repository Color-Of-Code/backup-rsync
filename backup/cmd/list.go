package cmd

import (
	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func buildListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the commands that will be executed",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")
			cfg := internal.LoadResolvedConfig(configPath)
			command := internal.NewRSyncCommand(rsyncPath)
			command.ListOnly = true

			cfg.Apply(command)
		},
	}
}
