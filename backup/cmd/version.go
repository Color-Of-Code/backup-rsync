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
		Run: func(cmd *cobra.Command, args []string) {
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")
			rsync := internal.NewRSyncCommand(rsyncPath)

			output, err := rsync.GetVersionInfo()
			if err != nil {
				fmt.Printf("%v\n", err)

				return
			}

			fmt.Printf("Rsync Binary Path: %s\n", rsyncPath)
			fmt.Printf("Version Info: %s", output)
		},
	}
}
