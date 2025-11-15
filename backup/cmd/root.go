// Package cmd contains the commands for the backup-tool CLI application.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	var configPath string

	rootCmd := &cobra.Command{
		Use:   "backup-tool",
		Short: "A tool for managing backups",
		Long:  `backup-tool is a CLI tool for managing backups and configurations.`,
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "Path to the configuration file")

	// Parse flags before adding commands to ensure configPath is available.
	err := rootCmd.ParseFlags(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	AddConfigCommands(rootCmd, configPath)
	AddBackupCommands(rootCmd, configPath)
	AddCheckCommands(rootCmd, configPath)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
