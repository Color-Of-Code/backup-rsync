package cmd

import (
	"backup-rsync/backup/internal"
	"io"
	"log"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildRunCommand(fs afero.Fs, shell internal.Exec) *cobra.Command {
	return buildJobCommand(fs, jobCommandOptions{
		use:   "run",
		short: "Execute the sync jobs",
		createLogger: func(fs afero.Fs, configPath string, now time.Time) (*log.Logger, string, func() error, error) {
			logPath := internal.GetLogPath(configPath, now)

			logger, cleanup, err := internal.CreateMainLogger(fs, logPath)

			return logger, logPath, cleanup, err
		},
		factory: func(rsyncPath string, logPath string, out io.Writer) internal.JobCommand {
			return internal.NewSyncCommand(rsyncPath, logPath, shell, out)
		},
	})
}
