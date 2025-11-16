// Package cmd contains the commands for the backup-tool CLI application.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "backup-tool",
		Short: "A tool for managing backups",
		Long:  `backup-tool is a CLI tool for managing backups and configurations.`,
	}

	rootCmd.PersistentFlags().String("config", "config.yaml", "Path to the configuration file")

	AddConfigCommands(rootCmd)
	AddBackupCommands(rootCmd)
	AddCheckCommands(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
