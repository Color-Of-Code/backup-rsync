package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type jobCommandOptions struct {
	use      string
	short    string
	needsLog bool
	simulate bool
	factory  func(rsyncPath string, logPath string, out io.Writer) internal.JobCommand
}

func buildJobCommand(fs afero.Fs, opts jobCommandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   opts.use,
		Short: opts.short,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")

			cfg, err := internal.LoadResolvedConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			out := cmd.OutOrStdout()
			logger := log.New(io.Discard, "", 0)

			var logPath string

			if opts.needsLog {
				var cleanup func() error

				logger, logPath, cleanup, err = internal.CreateMainLogger(fs, configPath, opts.simulate, time.Now())
				if err != nil {
					return fmt.Errorf("creating logger: %w", err)
				}

				defer cleanup()
			}

			command := opts.factory(rsyncPath, logPath, out)

			return cfg.Apply(command, logger)
		},
	}
}
