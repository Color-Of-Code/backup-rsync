package cmd

import (
	"backup-rsync/backup/internal"
	"io"

	"github.com/spf13/cobra"
)

func buildListCommand(shell internal.Exec) *cobra.Command {
	return buildJobCommand(nil, jobCommandOptions{
		use:   "list",
		short: "List the commands that will be executed",
		factory: func(rsyncPath string, _ string, out io.Writer) internal.JobCommand {
			return internal.NewListCommand(rsyncPath, shell, out)
		},
	})
}
