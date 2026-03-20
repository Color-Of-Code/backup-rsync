package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"

	"github.com/spf13/cobra"
)

func buildRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Execute the sync jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			cfg, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			logger, logPath, err := internal.CreateMainLogger(configPath, false)
			if err != nil {
				return fmt.Errorf("creating logger: %w", err)
			}

			out := cmd.OutOrStdout()
			command := internal.NewSyncCommand(rsyncPath, logPath, &internal.OsExec{}, out)

			cfg.Apply(command, logger, out)

			return nil
		},
	}
}
