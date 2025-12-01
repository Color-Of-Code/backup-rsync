package internal_test

import (
	. "backup-rsync/backup/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newMockSyncCommand() SyncCommand {
	return SyncCommand{
		SharedCommand: SharedCommand{
			BinPath: "/usr/bin/rsync",
			Shell:   &MockExec{},
		},
	}
}

func newMockSimulateCommand() SimulateCommand {
	return SimulateCommand{
		SharedCommand: SharedCommand{
			BinPath: "/usr/bin/rsync",
			Shell:   &MockExec{},
		},
	}
}

func TestApply(t *testing.T) {
	rsync := newMockSimulateCommand()

	job := NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
		WithExclusions([]string{"*.tmp"}),
	)

	status := job.Apply(rsync)
	assert.Equal(t, Success, status)
}
func TestApply_Disabled(t *testing.T) {
	command := newMockSyncCommand()

	disabledJob := NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status := disabledJob.Apply(command)
	assert.Equal(t, Skipped, status)
}

func TestApply_Invalid(t *testing.T) {
	rsync := newMockSyncCommand()

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := NewJob(
		WithName("invalid_job"),
		WithSource("/invalid/source/path"),
		WithTarget("/mnt/backup1/invalid/"),
	)

	status := invalidJob.Apply(rsync)
	assert.Equal(t, Failure, status)
}

func TestJobSkippedEnabledTrue(t *testing.T) {
	rsync := newMockSyncCommand()

	job := NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
	)

	status := job.Apply(rsync)
	assert.Equal(t, Success, status)
}

func TestJobSkippedEnabledFalse(t *testing.T) {
	rsync := newMockSyncCommand()

	disabledJob := NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status := disabledJob.Apply(rsync)
	assert.Equal(t, Skipped, status)
}

func TestJobSkippedEnabledOmitted(t *testing.T) {
	rsync := newMockSyncCommand()

	job := NewJob(
		WithName("omitted_enabled_job"),
		WithSource("/home/omitted/"),
		WithTarget("/mnt/backup1/omitted/"),
	)

	status := job.Apply(rsync)
	assert.Equal(t, Success, status)
}

func TestApplyWithMockedRsync(t *testing.T) {
	mockExecutor := &MockExec{}
	rsync := newMockSimulateCommand()
	rsync.Shell = mockExecutor

	job := NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
		WithExclusions([]string{"*.tmp"}),
	)
	status := job.Apply(rsync)

	assert.Equal(t, Success, status)
	assert.NotEmpty(t, mockExecutor.CapturedCommands)

	cmd := mockExecutor.CapturedCommands[0]

	assert.Equal(t, "/usr/bin/rsync", cmd.Name, "Command name mismatch")
	assert.Contains(t, cmd.Args, "--dry-run", "Expected --dry-run flag in command arguments")
}
