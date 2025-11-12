package internal

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// Helper function to create a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}

func TestBuildRsyncCmd(t *testing.T) {
	job := Job{
		Source:     "/home/user/Music/",
		Target:     "/target/user/music/home",
		Delete:     nil,
		Exclusions: []string{"*.tmp", "node_modules/"},
	}
	dryRun := true
	cmd := buildRsyncCmd(job, dryRun)

	expectedArgs := []string{
		"rsync", "--dry-run", "-aiv", "--info=progress2", "--delete",
		"--exclude=*.tmp", "--exclude=node_modules/",
		"/home/user/Music/", "/target/user/music/home",
	}

	if strings.Join(cmd.Args, " ") != strings.Join(expectedArgs, " ") {
		t.Errorf("Expected %v, got %v", expectedArgs, cmd.Args)
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
	dryRun := true
	logger := log.New(&bytes.Buffer{}, "", log.LstdFlags)

	status := ExecuteJob(job, dryRun, logger)
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}

	disabledJob := Job{
		Name:    "disabled_job",
		Source:  "/home/disabled/",
		Target:  "/mnt/backup1/disabled/",
		Enabled: boolPtr(false),
	}

	status = ExecuteJob(disabledJob, dryRun, logger)
	if status != "SKIPPED" {
		t.Errorf("Expected status SKIPPED, got %s", status)
	}

	// Test case for failure (simulate by providing invalid source path)
	invalidJob := Job{
		Name:   "invalid_job",
		Source: "/invalid/source/path",
		Target: "/mnt/backup1/invalid/",
	}

	status = ExecuteJob(invalidJob, false, logger)
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
	dryRun := true
	logger := log.New(&bytes.Buffer{}, "", log.LstdFlags)

	status := ExecuteJob(job, dryRun, logger)
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
	dryRun := true
	logger := log.New(&bytes.Buffer{}, "", log.LstdFlags)

	status := ExecuteJob(disabledJob, dryRun, logger)
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
	dryRun := true
	logger := log.New(&bytes.Buffer{}, "", log.LstdFlags)

	status := ExecuteJob(job, dryRun, logger)
	if status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", status)
	}
}
