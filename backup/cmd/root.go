package cmd

import (
	"github.com/spf13/cobra"
)

func BuildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "backup",
		Short: "A tool for managing backups",
		Long:  `backup is a CLI tool for managing backups and configurations.`,
	}

	rootCmd.PersistentFlags().String("config", "config.yaml", "Path to the configuration file")

	rootCmd.AddCommand(
		buildListCommand(),
		buildRunCommand(),
		buildSimulateCommand(),
		buildConfigCommand(),
		buildCheckCoverageCommand(),
	)

	return rootCmd
}
