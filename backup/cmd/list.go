package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
)

func buildListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the commands that will be executed",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			cfg, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			command := internal.NewListCommand(rsyncPath)

			logger := log.New(io.Discard, "", 0)
			cfg.Apply(command, logger)

			return nil
		},
	}
}
