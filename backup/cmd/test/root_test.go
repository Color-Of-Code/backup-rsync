package cmd_test

import (
	"bytes"
	"testing"

	"backup-rsync/backup/cmd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integration test to verify the root command and its sub-commands are set up correctly.
func TestBuildRootCommand_HelpOutput(t *testing.T) {
	rootCmd := cmd.BuildRootCommand()

	// Capture the help output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	helpOutput := buf.String()
	// Verify the help output contains expected content
	assert.Contains(t, helpOutput, "backup is a CLI tool for managing backups and configurations.",
		"Help output should contain the long description")
	assert.Contains(t, helpOutput, "backup [command]", "Help output should contain usage")

	// check persistent flags
	assert.Contains(t, helpOutput, "--config string       Path to the configuration file (default \"config.yaml\")")
	assert.Contains(t, helpOutput, "--rsync-path string   Path to the rsync binary (default \"/usr/bin/rsync\")")

	// check each sub-command is listed
	subCommands := []string{"list", "run", "simulate", "config", "check-coverage", "version"}
	for _, cmdName := range subCommands {
		assert.Regexp(t, "(?m)^  "+cmdName, helpOutput, "Help output should list the sub-command: "+cmdName)
	}
}
