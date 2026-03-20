package cmd

import (
	"fmt"

	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func buildVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the rsync version, protocol version, and full path to the rsync binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")
			rsync := internal.NewSyncCommand(rsyncPath, "")

			output, _, err := rsync.GetVersionInfo()
			if err != nil {
				return fmt.Errorf("getting version info: %w", err)
			}

			fmt.Printf("Rsync Binary Path: %s\n", rsyncPath)
			fmt.Printf("Version Info: %s", output)

			return nil
		},
	}
}
