package cmd

import (
	"fmt"

	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var AppFs = afero.NewOsFs()

func AddCheckCommands(rootCmd *cobra.Command, configPath string) {
	var checkCmd = &cobra.Command{
		Use:   "check-coverage",
		Short: "Check path coverage",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := internal.LoadResolvedConfig(configPath)
			uncoveredPaths := internal.ListUncoveredPaths(AppFs, cfg)

			fmt.Println("Uncovered paths:")

			for _, path := range uncoveredPaths {
				fmt.Println(path)
			}
		},
	}

	rootCmd.AddCommand(checkCmd)
}
