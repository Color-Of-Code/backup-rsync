package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "backup-tool",
	Short: "A tool for managing backups",
	Long:  `backup-tool is a CLI tool for managing backups and configurations.`,
}

// Define a global configPath variable and flag at the root level
var configPath string

func init() {
	RootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "Path to the configuration file")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
