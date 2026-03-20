package cmd

import (
	"fmt"

	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
)

func buildVersionCommand(shell internal.Exec) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the rsync version, protocol version, and full path to the rsync binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")
			out := cmd.OutOrStdout()
			rsync := internal.NewSyncCommand(rsyncPath, "", shell, out)

			output, _, err := rsync.GetVersionInfo()
			if err != nil {
				return fmt.Errorf("getting version info: %w", err)
			}

			fmt.Fprintf(out, "Rsync Binary Path: %s\n", rsyncPath)
			fmt.Fprintf(out, "Version Info: %s", output)

			return nil
		},
	}
}
