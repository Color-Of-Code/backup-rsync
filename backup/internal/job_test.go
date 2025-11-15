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
	job := internal.Job{
		Source:     "/home/user/Music/",
		Target:     "/target/user/music/home",
		Delete:     true,
		Exclusions: []string{"*.tmp", "node_modules/"},
	}
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

	job := internal.Job{
		Name:       "test_job",
		Source:     "/home/test/",
		Target:     "/mnt/backup1/test/",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{"*.tmp"},
	}
	simulate := true

	status := internal.ExecuteJobWithExecutor(job, simulate, false, "", mockExecutor)
	if status != statusSuccess {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

	disabledJob := internal.Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: false,
	}

	status = internal.ExecuteJobWithExecutor(disabledJob, simulate, false, "", mockExecutor)
	if status != "SKIPPED" {
		t.Errorf("Expected status SKIPPED, got %s", status)
	}

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := internal.Job{
		Name:    "invalid_job",
		Source:  "/invalid/source/path",
		Target:  "/mnt/backup1/invalid/",
		Delete:  true,
		Enabled: true,
	}

	status = internal.ExecuteJobWithExecutor(invalidJob, false, false, "", mockExecutor)
	if status != "FAILURE" {
		t.Errorf("Expected status FAILURE, got %s", status)
	}
}

// Ensure all references to ExecuteJob are prefixed with internal.
func TestJobSkippedEnabledTrue(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := internal.Job{
		Name:    "test_job",
		Source:  "/home/test/",
		Target:  "/mnt/backup1/test/",
		Enabled: true,
	}

	status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)
	if status != statusSuccess {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}

func TestJobSkippedEnabledFalse(t *testing.T) {
	// Create mock executor (won't be used since job is disabled)
	mockExecutor := &MockCommandExecutor{}

	disabledJob := internal.Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: false,
	}

	status := internal.ExecuteJobWithExecutor(disabledJob, true, false, "", mockExecutor)
	if status != "SKIPPED" {
		t.Errorf("Expected status SKIPPED, got %s", status)
	}
}

func TestJobSkippedEnabledOmitted(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := internal.Job{
		Name:    "omitted_enabled_job",
		Source:  "/home/omitted/",
		Target:  "/mnt/backup1/omitted/",
		Delete:  true,
		Enabled: true,
	}

	status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)
	if status != statusSuccess {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}

func TestExecuteJobWithMockedRsync(t *testing.T) {
	// Create mock executor
	mockExecutor := &MockCommandExecutor{}

	job := internal.Job{
		Name:       "test_job",
		Source:     "/home/test/",
		Target:     "/mnt/backup1/test/",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{"*.tmp"},
	}
	status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)

	if status != statusSuccess {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

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
