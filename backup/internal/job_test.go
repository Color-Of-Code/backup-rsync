package internal

import (
	"os/exec"
	"strings"
	"testing"
)

// Helper function to create a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}

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
	execCommand = mockExecCommand
}

func TestBuildRsyncCmd(t *testing.T) {
	job := Job{
		Source:     "/home/user/Music/",
		Target:     "/target/user/music/home",
		Delete:     nil,
		Exclusions: []string{"*.tmp", "node_modules/"},
	}
	args := buildRsyncCmd(job, true, "")

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
	job := Job{
		Name:       "test_job",
		Source:     "/home/test/",
		Target:     "/mnt/backup1/test/",
		Delete:     nil,
		Exclusions: []string{"*.tmp"},
	}
	simulate := true

	status := ExecuteJob(job, simulate, false, "")
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

	disabledJob := Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: boolPtr(false),
	}

	status = ExecuteJob(disabledJob, simulate, false, "")
	if status != "SKIPPED" {
		t.Errorf("Expected status SKIPPED, got %s", status)
	}

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := Job{
		Name:   "invalid_job",
		Source: "/invalid/source/path",
		Target: "/mnt/backup1/invalid/",
	}

	status = ExecuteJob(invalidJob, false, false, "")
	if status != "FAILURE" {
		t.Errorf("Expected status FAILURE, got %s", status)
	}
}

func TestJobSkippedEnabledTrue(t *testing.T) {
	job := Job{
		Name:    "test_job",
		Source:  "/home/test/",
		Target:  "/mnt/backup1/test/",
		Enabled: boolPtr(true),
	}
	status := ExecuteJob(job, true, false, "")
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}

func TestJobSkippedEnabledFalse(t *testing.T) {
	disabledJob := Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: boolPtr(false),
	}
	status := ExecuteJob(disabledJob, true, false, "")
	if status != "SKIPPED" {
		t.Errorf("Expected status SKIPPED, got %s", status)
	}
}

func TestJobSkippedEnabledOmitted(t *testing.T) {
	job := Job{
		Name:   "omitted_enabled_job",
		Source: "/home/omitted/",
		Target: "/mnt/backup1/omitted/",
	}
	status := ExecuteJob(job, true, false, "")
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}

func TestExecuteJobWithMockedRsync(t *testing.T) {
	// Reset capturedArgs before the test
	capturedArgs = nil

	job := Job{
		Name:       "test_job",
		Source:     "/home/test/",
		Target:     "/mnt/backup1/test/",
		Delete:     nil,
		Exclusions: []string{"*.tmp"},
	}
	status := ExecuteJob(job, true, false, "")

	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

	if len(capturedArgs) == 0 || capturedArgs[0] != "--dry-run" {
		t.Errorf("Expected --dry-run flag, got %v", capturedArgs)
	}
}
