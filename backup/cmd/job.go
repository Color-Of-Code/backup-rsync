package cmd

import (
	"backup-rsync/backup/internal"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// LoggerFactory creates a logger, returning the logger, log directory path, cleanup function, and any error.
type LoggerFactory func(fs afero.Fs, configPath string, now time.Time) (*slog.Logger, string, func() error, error)

func discardLoggerFactory(_ afero.Fs, _ string, _ time.Time) (*slog.Logger, string, func() error, error) {
	return slog.New(slog.DiscardHandler), "", func() error { return nil }, nil
}

type jobCommandOptions struct {
	use          string
	short        string
	factory      func(rsyncPath string, logPath string, out io.Writer) internal.JobCommand
	createLogger LoggerFactory
}

// parseSetFlags parses --set flag values (key=value) into a map.
func parseSetFlags(cmd *cobra.Command) map[string]string {
	setFlags, _ := cmd.Flags().GetStringArray("set")
	overrides := make(map[string]string, len(setFlags))

	for _, s := range setFlags {
		key, value, ok := strings.Cut(s, "=")
		if ok {
			overrides[key] = value
		}
	}

	return overrides
}

func buildJobCommand(fs afero.Fs, opts jobCommandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   opts.use,
		Short: opts.short,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			rsyncPath, _ := cmd.Flags().GetString("rsync-path")
			overrides := parseSetFlags(cmd)

			cfg, err := internal.LoadResolvedConfig(configPath, overrides)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			out := cmd.OutOrStdout()

			createLogger := opts.createLogger
			if createLogger == nil {
				createLogger = discardLoggerFactory
			}

			logger, logPath, cleanup, err := createLogger(fs, configPath, time.Now())
			if err != nil {
				return fmt.Errorf("creating logger: %w", err)
			}

			defer cleanup()

			command := opts.factory(rsyncPath, logPath, out)

			return cfg.Apply(command, logger)
		},
	}
}
