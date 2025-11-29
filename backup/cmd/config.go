package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"

	"github.com/spf13/cobra"
)

func buildConfigCommand() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	var showVerb = &cobra.Command{
		Use:   "show",
		Short: "Show resolved configuration",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			cfg := internal.LoadResolvedConfig(configPath)

			fmt.Printf("Resolved Configuration:\n%s\n", cfg)
		},
	}

	var validateVerb = &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			internal.LoadResolvedConfig(configPath)
			fmt.Println("Configuration is valid.")
		},
	}

	configCmd.AddCommand(showVerb)
	configCmd.AddCommand(validateVerb)

	return configCmd
}
