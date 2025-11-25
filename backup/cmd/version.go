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
			var executor internal.CommandExecutor = &internal.RealCommandExecutor{}

			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			output, err := internal.FetchRsyncVersion(executor, rsyncPath)
			if err != nil {
				fmt.Printf("%v\n", err)

				return
			}

			fmt.Printf("Rsync Binary Path: %s\n", rsyncPath)
			fmt.Printf("Version Info: %s", output)
		},
	}
}
