package internal_test

import (
	"backup-rsync/backup/internal"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const statusSuccess = "SUCCESS"

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

func TestBuildRsyncCmd(t *testing.T) {
	job := *NewJob(
		WithSource("/home/user/Music/"),
		WithTarget("/target/user/music/home"),
		WithExclusions([]string{"*.tmp", "node_modules/"}),
	)
	args := internal.BuildRsyncCmd(job, true, "")

	expectedArgs := []string{
		"--dry-run", "-aiv", "--stats", "--delete",
		"--exclude=*.tmp", "--exclude=node_modules/",
		"/home/user/Music/", "/target/user/music/home",
	}

	assert.Equal(t, strings.Join(expectedArgs, " "), strings.Join(args, " "))
}

func TestExecuteJob(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := *NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
		WithExclusions([]string{"*.tmp"}),
	)
	simulate := true

	status := internal.ExecuteJobWithExecutor(job, simulate, false, "", mockExecutor)
	assert.Equal(t, statusSuccess, status)

	disabledJob := *NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status = internal.ExecuteJobWithExecutor(disabledJob, simulate, false, "", mockExecutor)
	assert.Equal(t, "SKIPPED", status)

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := *NewJob(
		WithName("invalid_job"),
		WithSource("/invalid/source/path"),
		WithTarget("/mnt/backup1/invalid/"),
	)

	status = internal.ExecuteJobWithExecutor(invalidJob, false, false, "", mockExecutor)
	assert.Equal(t, "FAILURE", status)
}

// Ensure all references to ExecuteJob are prefixed with internal.
func TestJobSkippedEnabledTrue(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := *NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
	)

	status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)
	assert.Equal(t, statusSuccess, status)
}

func TestJobSkippedEnabledFalse(t *testing.T) {
	// Create mock executor (won't be used since job is disabled)
	mockExecutor := &MockCommandExecutor{}

	disabledJob := *NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status := internal.ExecuteJobWithExecutor(disabledJob, true, false, "", mockExecutor)
	assert.Equal(t, "SKIPPED", status)
}

func TestJobSkippedEnabledOmitted(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := *NewJob(
		WithName("omitted_enabled_job"),
		WithSource("/home/omitted/"),
		WithTarget("/mnt/backup1/omitted/"),
	)

	status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)
	assert.Equal(t, statusSuccess, status)
}

func TestExecuteJobWithMockedRsync(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := *NewJob(
		WithName("test_job"),
		WithSource("/home/test/"),
		WithTarget("/mnt/backup1/test/"),
		WithExclusions([]string{"*.tmp"}),
	)
	status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)

	assert.Equal(t, statusSuccess, status)
	assert.NotEmpty(t, mockExecutor.CapturedCommands)

	cmd := mockExecutor.CapturedCommands[0]

	assert.Equal(t, "rsync", cmd.Name, "Command name mismatch")
	assert.Contains(t, cmd.Args, "--dry-run", "Expected --dry-run flag in command arguments")
}
