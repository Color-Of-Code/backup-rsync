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
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")

			cfg, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Resolved Configuration:\n%s\n", cfg)

			return nil
		},
	}

	var validateVerb = &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")

			_, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("validating config: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid.")

			return nil
		},
	}

	configCmd.AddCommand(showVerb)
	configCmd.AddCommand(validateVerb)

	return configCmd
}
