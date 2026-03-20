package cmd

import (
	"fmt"

	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildCheckCoverageCommand() *cobra.Command {
	var fs = afero.NewOsFs()

	return &cobra.Command{
		Use:   "check-coverage",
		Short: "Check path coverage",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")

			cfg, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			uncoveredPaths := internal.ListUncoveredPaths(fs, cfg)

			fmt.Println("Uncovered paths:")

			for _, path := range uncoveredPaths {
				fmt.Println(path)
			}

			return nil
		},
	}
}
