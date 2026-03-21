package cmd

import (
	"backup-rsync/backup/internal"
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildSimulateCommand(fs afero.Fs, shell internal.Exec) *cobra.Command {
	return buildJobCommand(fs, jobCommandOptions{
		use:      "simulate",
		short:    "Simulate the sync jobs",
		needsLog: true,
		simulate: true,
		factory: func(rsyncPath string, logPath string, out io.Writer) internal.JobCommand {
			return internal.NewSimulateCommand(rsyncPath, logPath, shell, out)
		},
	})
}
