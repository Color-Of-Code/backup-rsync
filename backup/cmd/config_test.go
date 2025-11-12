package cmd

import (
	"bytes"
	"strings"
	"testing"

	"backup-rsync/backup/internal"
)

func TestSubstituteVariables(t *testing.T) {
	variables := map[string]string{
		"target_base": "/mnt/backup1",
	}
	input := "${target_base}/user/music/home"
	expected := "/mnt/backup1/user/music/home"
	result := substituteVariables(input, variables)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestLoadConfig(t *testing.T) {
	yamlData := `
variables:
  target_base: "/mnt/backup1"

jobs:
  - name: "test_job"
    source: "/home/test/"
    target: "${target_base}/test/"
    enabled: true
`
	reader := bytes.NewReader([]byte(yamlData))
	cfg, err := loadConfig(reader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Variables["target_base"] != "/mnt/backup1" {
		t.Errorf("Expected /mnt/backup1, got %s", cfg.Variables["target_base"])
	}

	if len(cfg.Jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(cfg.Jobs))
	}

	job := cfg.Jobs[0]
	if job.Name != "test_job" {
		t.Errorf("Expected job name test_job, got %s", job.Name)
	}
	if job.Source != "/home/test/" {
		t.Errorf("Expected source /home/test/, got %s", job.Source)
	}
	if job.Target != "${target_base}/test/" {
		t.Errorf("Expected target ${target_base}/test/, got %s", job.Target)
	}
}

func TestValidateJobNames(t *testing.T) {
	tests := []struct {
		name         string
		jobs         []internal.Job
		expectsError bool
		errorMessage string
	}{
		{
			name: "Valid job names",
			jobs: []internal.Job{
				{Name: "job1"},
				{Name: "job2"},
			},
			expectsError: false,
		},
		{
			name: "Duplicate job names",
			jobs: []internal.Job{
				{Name: "job1"},
				{Name: "job1"},
			},
			expectsError: true,
			errorMessage: "duplicate job name: job1",
		},
		{
			name: "Invalid characters in job name",
			jobs: []internal.Job{
				{Name: "job 1"},
			},
			expectsError: true,
			errorMessage: "invalid characters in job name: job 1",
		},
		{
			name: "Mixed errors",
			jobs: []internal.Job{
				{Name: "job1"},
				{Name: "job 1"},
				{Name: "job1"},
			},
			expectsError: true,
			errorMessage: "duplicate job name: job1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateJobNames(test.jobs)
			if test.expectsError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), test.errorMessage) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", test.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name         string
		jobPath      string
		paths        []internal.Path
		pathType     string
		jobName      string
		expectsError bool
		errorMessage string
	}{
		{
			name:         "Valid source path",
			jobPath:      "/home/user/documents",
			paths:        []internal.Path{{Path: "/home/user"}},
			pathType:     "source",
			jobName:      "job1",
			expectsError: false,
		},
		{
			name:         "Invalid source path",
			jobPath:      "/invalid/source",
			paths:        []internal.Path{{Path: "/home/user"}},
			pathType:     "source",
			jobName:      "job1",
			expectsError: true,
			errorMessage: "invalid source path for job 'job1': /invalid/source",
		},
		{
			name:         "Valid target path",
			jobPath:      "/mnt/backup/documents",
			paths:        []internal.Path{{Path: "/mnt/backup"}},
			pathType:     "target",
			jobName:      "job1",
			expectsError: false,
		},
		{
			name:         "Invalid target path",
			jobPath:      "/invalid/target",
			paths:        []internal.Path{{Path: "/mnt/backup"}},
			pathType:     "target",
			jobName:      "job1",
			expectsError: true,
			errorMessage: "invalid target path for job 'job1': /invalid/target",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePath(test.jobPath, test.paths, test.pathType, test.jobName)
			if test.expectsError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != test.errorMessage {
					t.Errorf("Expected error message '%s', but got '%s'", test.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidatePaths(t *testing.T) {
	tests := []struct {
		name         string
		cfg          internal.Config
		expectsError bool
		errorMessage string
	}{
		{
			name: "Valid paths",
			cfg: internal.Config{
				Sources: []internal.Path{
					{Path: "/home/user"},
				},
				Targets: []internal.Path{
					{Path: "/mnt/backup"},
				},
				Jobs: []internal.Job{
					{Name: "job1", Source: "/home/user/documents", Target: "/mnt/backup/documents"},
				},
			},
			expectsError: false,
		},
		{
			name: "Invalid paths",
			cfg: internal.Config{
				Sources: []internal.Path{
					{Path: "/home/user"},
				},
				Targets: []internal.Path{
					{Path: "/mnt/backup"},
				},
				Jobs: []internal.Job{
					{Name: "job1", Source: "/invalid/source", Target: "/invalid/target"},
				},
			},
			expectsError: true,
			errorMessage: "path validation errors: [invalid source path for job 'job1': /invalid/source invalid target path for job 'job1': /invalid/target]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePaths(test.cfg)
			if test.expectsError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != test.errorMessage {
					t.Errorf("Expected error message '%s', but got '%s'", test.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
