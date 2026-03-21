package cmd

import (
	"backup-rsync/backup/internal"
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildRunCommand(fs afero.Fs, shell internal.Exec) *cobra.Command {
	return buildJobCommand(fs, jobCommandOptions{
		use:      "run",
		short:    "Execute the sync jobs",
		needsLog: true,
		factory: func(rsyncPath string, logPath string, out io.Writer) internal.JobCommand {
			return internal.NewSyncCommand(rsyncPath, logPath, shell, out)
		},
	})
}
