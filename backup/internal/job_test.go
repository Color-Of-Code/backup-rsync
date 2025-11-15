package internal_test

import (
	"os/exec"
	"strings"
	"testing"

	"backup-rsync/backup/internal"
)

var capturedArgs []string

var mockExecCommand = func(name string, args ...string) *exec.Cmd {
	if name == "rsync" {
		capturedArgs = append(capturedArgs, args...) // Append arguments for assertions
		if strings.Contains(strings.Join(args, " "), "--dry-run") {
			return exec.Command("echo", "mocked rsync success") // Simulate success for dry-run
		}
		if strings.Contains(strings.Join(args, " "), "/invalid/source/path") {
			return exec.Command("false") // Simulate failure for invalid paths
		}

		return exec.Command("echo", "mocked rsync success") // Simulate general success
	}

	return exec.Command(name, args...)
}

func init() {
	internal.ExecCommand = mockExecCommand
}

func TestBuildRsyncCmd(t *testing.T) {
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
	job := internal.Job{
		Name:       "test_job",
		Source:     "/home/test/",
		Target:     "/mnt/backup1/test/",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{"*.tmp"},
	}
	simulate := true

	status := internal.ExecuteJob(job, simulate, false, "")
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

	disabledJob := internal.Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: false,
	}

	status = internal.ExecuteJob(disabledJob, simulate, false, "")
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

	status = internal.ExecuteJob(invalidJob, false, false, "")
	if status != "FAILURE" {
		t.Errorf("Expected status FAILURE, got %s", status)
	}
}

// Ensure all references to ExecuteJob are prefixed with internal
func TestJobSkippedEnabledTrue(t *testing.T) {
	job := internal.Job{
		Name:    "test_job",
		Source:  "/home/test/",
		Target:  "/mnt/backup1/test/",
		Enabled: true,
	}

	status := internal.ExecuteJob(job, true, false, "")
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}

func TestJobSkippedEnabledFalse(t *testing.T) {
	disabledJob := internal.Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: false,
	}

	status := internal.ExecuteJob(disabledJob, true, false, "")
	if status != "SKIPPED" {
		t.Errorf("Expected status SKIPPED, got %s", status)
	}
}

func TestJobSkippedEnabledOmitted(t *testing.T) {
	job := internal.Job{
		Name:    "omitted_enabled_job",
		Source:  "/home/omitted/",
		Target:  "/mnt/backup1/omitted/",
		Delete:  true,
		Enabled: true,
	}

	status := internal.ExecuteJob(job, true, false, "")
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}

func TestExecuteJobWithMockedRsync(t *testing.T) {
	// Reset capturedArgs before the test
	capturedArgs = nil

	job := internal.Job{
		Name:       "test_job",
		Source:     "/home/test/",
		Target:     "/mnt/backup1/test/",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{"*.tmp"},
	}
	status := internal.ExecuteJob(job, true, false, "")

	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

	if len(capturedArgs) == 0 || capturedArgs[0] != "--dry-run" {
		t.Errorf("Expected --dry-run flag, got %v", capturedArgs)
	}
}
