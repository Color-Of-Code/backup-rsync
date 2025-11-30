package cmd

import (
	"backup-rsync/backup/internal"
	"io"
	"log"

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
			command := internal.NewListCommand(rsyncPath)

			logger := log.New(io.Discard, "", 0)
			cfg.Apply(command, logger)
		},
	}
}
