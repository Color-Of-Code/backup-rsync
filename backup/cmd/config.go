package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `The config subcommand allows you to manage configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement config logic
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
