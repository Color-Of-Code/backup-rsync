package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func buildConfigCommand() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Implementation for the config command
			fmt.Println("Config command executed")
		},
	}

	var showVerb = &cobra.Command{
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
