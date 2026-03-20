package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func buildSimulateCommand(shell internal.Exec) *cobra.Command {
	return &cobra.Command{
		Use:   "simulate",
		Short: "Simulate the sync jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			cfg, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			logger, logPath, cleanup, err := internal.CreateMainLogger(configPath, true, time.Now())
			if err != nil {
				return fmt.Errorf("creating logger: %w", err)
			}

			defer cleanup()

			out := cmd.OutOrStdout()
			command := internal.NewSimulateCommand(rsyncPath, logPath, shell, out)

			return cfg.Apply(command, logger, out)
		},
	}
}
