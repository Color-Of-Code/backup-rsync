package internal_test

import (
	"errors"
	"strings"
	"testing"

	"backup-rsync/backup/internal"
)

// Static error for testing.
var ErrExitStatus23 = errors.New("exit status 23")

const statusSuccess = "SUCCESS"

// MockCommandExecutor implements CommandExecutor for testing.
type MockCommandExecutor struct {
	CapturedCommands []MockCommand
}

// MockCommand represents a captured command execution.
type MockCommand struct {
	Name string
	Args []string
}

// Option defines a function that modifies a Job.
type Option func(*internal.Job)

// NewJob is a job factory with defaults.
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

// Execute captures the command and simulates execution.
func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	m.CapturedCommands = append(m.CapturedCommands, MockCommand{
		Name: name,
		Args: append([]string{}, args...), // Make a copy of args
	})

	if name == "rsync" {
		// Simulate different scenarios based on arguments
		argsStr := strings.Join(args, " ")

		if strings.Contains(argsStr, "/invalid/source/path") {
			errMsg := "rsync: link_stat \"/invalid/source/path\" failed: No such file or directory"

			return []byte(errMsg), ErrExitStatus23
		}

		return []byte("mocked rsync success"), nil
	}

	return []byte("command not mocked"), nil
}

func TestBuildRsyncCmd(t *testing.T) {
	// This test doesn't need mocking since it only builds args
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

	if strings.Join(args, " ") != strings.Join(expectedArgs, " ") {
		t.Errorf("Expected %v, got %v", expectedArgs, args)
	}
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
	expectStatus(t, status, statusSuccess)

	disabledJob := *NewJob(
		WithName("disabled_job"),
		WithSource("/home/disabled/"),
		WithTarget("/mnt/backup1/disabled/"),
		WithEnabled(false),
	)

	status = internal.ExecuteJobWithExecutor(disabledJob, simulate, false, "", mockExecutor)
	expectStatus(t, status, "SKIPPED")

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := *NewJob(
		WithName("invalid_job"),
		WithSource("/invalid/source/path"),
		WithTarget("/mnt/backup1/invalid/"),
	)

	status = internal.ExecuteJobWithExecutor(invalidJob, false, false, "", mockExecutor)
	expectStatus(t, status, "FAILURE")
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
	expectStatus(t, status, statusSuccess)
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
	expectStatus(t, status, "SKIPPED")
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
	expectStatus(t, status, statusSuccess)
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

	expectStatus(t, status, statusSuccess)

	// Check that rsync was called with the expected arguments
	if len(mockExecutor.CapturedCommands) == 0 {
		t.Errorf("Expected at least one command to be executed")

		return
	}

	cmd := mockExecutor.CapturedCommands[0]
	if cmd.Name != "rsync" {
		t.Errorf("Expected command to be 'rsync', got %s", cmd.Name)
	}

	if len(cmd.Args) == 0 || cmd.Args[0] != "--dry-run" {
		t.Errorf("Expected --dry-run flag, got %v", cmd.Args)
	}
}

func expectStatus(t *testing.T, status, expectedStatus string) {
	t.Helper()

	if status != expectedStatus {
		t.Errorf("Expected status %s, got %s", expectedStatus, status)
	}
}
