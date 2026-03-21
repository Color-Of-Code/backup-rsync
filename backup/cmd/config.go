package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"

	"github.com/spf13/cobra"
)

type configVerb struct {
	use     string
	short   string
	errCtx  string
	success func(cmd *cobra.Command, cfg internal.Config)
}

func buildConfigCommand() *cobra.Command {
	configVerbs := []configVerb{
		{
			use:    "show",
			short:  "Show resolved configuration",
			errCtx: "loading config",
			success: func(cmd *cobra.Command, cfg internal.Config) {
				fmt.Fprintf(cmd.OutOrStdout(), "Resolved Configuration:\n%s\n", cfg)
			},
		},
		{
			use:    "validate",
			short:  "Validate configuration",
			errCtx: "validating config",
			success: func(cmd *cobra.Command, _ internal.Config) {
				fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid.")
			},
		},
	}
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	for _, verb := range configVerbs {
		configCmd.AddCommand(&cobra.Command{
			Use:   verb.use,
			Short: verb.short,
			RunE:  configRunE(verb),
		})
	}

	return configCmd
}

func configRunE(verb configVerb) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")

		cfg, err := internal.LoadResolvedConfig(configPath)
		if err != nil {
			return fmt.Errorf("%s: %w", verb.errCtx, err)
		}

		verb.success(cmd, cfg)

		return nil
	}
}
