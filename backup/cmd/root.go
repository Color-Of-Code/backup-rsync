package cmd

import (
	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// BuildRootCommand creates the root cobra command with production defaults.
func BuildRootCommand() *cobra.Command {
	return BuildRootCommandWithDeps(afero.NewOsFs(), &internal.OsExec{})
}

// BuildRootCommandWithFs creates the root command with a custom filesystem.
func BuildRootCommandWithFs(fs afero.Fs) *cobra.Command {
	return BuildRootCommandWithDeps(fs, &internal.OsExec{})
}

// BuildRootCommandWithDeps creates the root command with full dependency injection.
func BuildRootCommandWithDeps(fs afero.Fs, shell internal.Exec) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "backup",
		Short: "A tool for managing backups",
		Long:  `backup is a CLI tool for managing backups and configurations.`,
	}

	rootCmd.PersistentFlags().String("config", "config.yaml", "Path to the configuration file")
	rootCmd.PersistentFlags().String("rsync-path", "/usr/bin/rsync", "Path to the rsync binary")
	rootCmd.PersistentFlags().StringArray("set", nil, "Set a variable override (key=value), can be repeated")

	rootCmd.AddCommand(
		buildListCommand(shell),
		buildRunCommand(fs, shell),
		buildSimulateCommand(fs, shell),
		buildConfigCommand(),
		buildCheckCoverageCommand(fs),
		buildVersionCommand(shell),
	)

	return rootCmd
}
