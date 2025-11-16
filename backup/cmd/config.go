package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// AddConfigCommands binds the config command and its subcommands to the root command.
func AddConfigCommands(rootCmd *cobra.Command) {
	// configCmd represents the config command.
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Implementation for the config command
			fmt.Println("Config command executed")
		},
	}

	// Extend the config subcommand with the show verb.
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show resolved configuration",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			cfg := internal.LoadResolvedConfig(configPath)

			out, err := yaml.Marshal(cfg)
			if err != nil {
				log.Fatalf("Failed to marshal resolved configuration: %v", err)
			}

			fmt.Printf("Resolved Configuration:\n%s\n", string(out))
		},
	}

	// Extend the config subcommand with the validate verb.
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			internal.LoadResolvedConfig(configPath)
			fmt.Println("Configuration is valid.")
		},
	}

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(validateCmd)
}
