package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func buildSimulateCommand() *cobra.Command {
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

			logger, logPath, err := internal.CreateMainLogger(configPath, true)
			if err != nil {
				return fmt.Errorf("creating logger: %w", err)
			}

			command := internal.NewSimulateCommand(rsyncPath, logPath, &internal.OsExec{}, os.Stdout)

			cfg.Apply(command, logger, os.Stdout)

			return nil
		},
	}
}
