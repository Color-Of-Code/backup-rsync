package cmd

import (
	"fmt"

	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var AppFs = afero.NewOsFs()

var checkCmd = &cobra.Command{
	Use:   "check-coverage",
	Short: "Check path coverage",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadResolvedConfig(configPath)
		uncoveredPaths := internal.ListUncoveredPaths(AppFs, cfg)
		fmt.Println("Uncovered paths:")
		for _, path := range uncoveredPaths {
			fmt.Println(path)
		}
	},
}

func init() {
	RootCmd.AddCommand(checkCmd)
}
