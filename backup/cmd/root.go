package cmd

import (
	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func BuildRootCommand() *cobra.Command {
	return BuildRootCommandWithDeps(afero.NewOsFs(), &internal.OsExec{})
}

func BuildRootCommandWithFs(fs afero.Fs) *cobra.Command {
	return BuildRootCommandWithDeps(fs, &internal.OsExec{})
}

func BuildRootCommandWithDeps(fs afero.Fs, shell internal.Exec) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "backup",
		Short: "A tool for managing backups",
		Long:  `backup is a CLI tool for managing backups and configurations.`,
	}

	rootCmd.PersistentFlags().String("config", "config.yaml", "Path to the configuration file")
	rootCmd.PersistentFlags().String("rsync-path", "/usr/bin/rsync", "Path to the rsync binary")

	rootCmd.AddCommand(
		buildListCommand(shell),
		buildRunCommand(shell),
		buildSimulateCommand(shell),
		buildConfigCommand(),
		buildCheckCoverageCommand(fs),
		buildVersionCommand(shell),
	)

	return rootCmd
}
