package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildCheckCoverageCommand(fs afero.Fs) *cobra.Command {
	return &cobra.Command{
		Use:   "check-coverage",
		Short: "Check path coverage",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			overrides := parseSetFlags(cmd)

			cfg, err := internal.LoadResolvedConfig(configPath, overrides)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			checker := &internal.CoverageChecker{
				Logger: slog.New(internal.NewUTCTextHandler(os.Stderr)),
				Fs:     fs,
			}

			uncoveredPaths := checker.ListUncoveredPaths(cfg)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Uncovered paths:")

			for _, path := range uncoveredPaths {
				fmt.Fprintln(out, path)
			}

			return nil
		},
	}
}
