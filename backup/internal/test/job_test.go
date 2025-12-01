package internal_test

import (
	"backup-rsync/backup/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Option func(*internal.Job)

func NewJob(opts ...Option) *internal.Job {
	// Default values
	job := &internal.Job{
		Name:       "job",
		Source:     "",
		Target:     "",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{},
	}

	// Apply all options (overrides defaults)
	for _, opt := range opts {
		opt(job)
	}

	return job
}

func WithName(name string) Option {
	return func(p *internal.Job) {
		p.Name = name
	}
}

func WithSource(source string) Option {
	return func(p *internal.Job) {
		p.Source = source
	}
}

func WithTarget(target string) Option {
	return func(p *internal.Job) {
		p.Target = target
	}
}

func WithEnabled(enabled bool) Option {
	return func(p *internal.Job) {
		p.Enabled = enabled
	}
}

func WithExclusions(exclusions []string) Option {
	return func(p *internal.Job) {
		p.Exclusions = exclusions
	}
}

func newMockSyncCommand() internal.SyncCommand {
	return internal.SyncCommand{
		BinPath:  "/usr/bin/rsync",
		Executor: &MockCommandExecutor{},
	}
}

func newMockSimulateCommand() internal.SimulateCommand {
	return internal.SimulateCommand{
		SyncCommand: internal.SyncCommand{
			BinPath:  "/usr/bin/rsync",
			Executor: &MockCommandExecutor{},
		},
	}
}

func TestApply(t *testing.T) {
	rsync := newMockSimulateCommand()

	job := *NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
		WithExclusions([]string{"*.tmp"}),
	)

	status := job.Apply(rsync)
	assert.Equal(t, internal.Success, status)
}
func TestApply_Disabled(t *testing.T) {
	command := newMockSyncCommand()

	disabledJob := *NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status := disabledJob.Apply(command)
	assert.Equal(t, internal.Skipped, status)
}

func TestApply_Invalid(t *testing.T) {
	rsync := newMockSyncCommand()

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := *NewJob(
		WithName("invalid_job"),
		WithSource("/invalid/source/path"),
		WithTarget("/mnt/backup1/invalid/"),
	)

	status := invalidJob.Apply(rsync)
	assert.Equal(t, internal.Failure, status)
}

func TestJobSkippedEnabledTrue(t *testing.T) {
	rsync := newMockSyncCommand()

	job := *NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
	)

	status := job.Apply(rsync)
	assert.Equal(t, internal.Success, status)
}

func TestJobSkippedEnabledFalse(t *testing.T) {
	rsync := newMockSyncCommand()

	disabledJob := *NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status := disabledJob.Apply(rsync)
	assert.Equal(t, internal.Skipped, status)
}

func TestJobSkippedEnabledOmitted(t *testing.T) {
	rsync := newMockSyncCommand()

	job := *NewJob(
		WithName("omitted_enabled_job"),
		WithSource("/home/omitted/"),
		WithTarget("/mnt/backup1/omitted/"),
	)

	status := job.Apply(rsync)
	assert.Equal(t, internal.Success, status)
}

func TestApplyWithMockedRsync(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}
	rsync := newMockSimulateCommand()
	rsync.Executor = mockExecutor

	job := *NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
		WithExclusions([]string{"*.tmp"}),
	)
	status := job.Apply(rsync)

	assert.Equal(t, internal.Success, status)
	assert.NotEmpty(t, mockExecutor.CapturedCommands)

	cmd := mockExecutor.CapturedCommands[0]

	assert.Equal(t, "/usr/bin/rsync", cmd.Name, "Command name mismatch")
	assert.Contains(t, cmd.Args, "--dry-run", "Expected --dry-run flag in command arguments")
}
