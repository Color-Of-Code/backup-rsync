package internal

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConfig1(t *testing.T) {
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

	cfg, err := LoadConfig(reader)
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

func TestLoadConfig2(t *testing.T) {
	yamlData := `
jobs:
  - name: "job1"
    source: "/source1"
    target: "/target1"
  - name: "job2"
    source: "/source2"
    target: "/target2"
    delete: false
    enabled: false
`

	// Use a reader instead of a mock file
	reader := bytes.NewReader([]byte(yamlData))

	cfg, err := LoadConfig(reader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expected := []Job{
		{
			Name:    "job1",
			Source:  "/source1",
			Target:  "/target1",
			Delete:  true,
			Enabled: true,
		},
		{
			Name:    "job2",
			Source:  "/source2",
			Target:  "/target2",
			Delete:  false,
			Enabled: false,
		},
	}

	if !reflect.DeepEqual(cfg.Jobs, expected) {
		t.Errorf("got %+v, want %+v", cfg.Jobs, expected)
	}
}

func TestYAMLUnmarshalingDefaults(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected Job
	}{
		{
			name: "Defaults applied when fields omitted",
			yamlData: `
name: "test_job"
source: "/source"
target: "/target"
`,
			expected: Job{
				Name:    "test_job",
				Source:  "/source",
				Target:  "/target",
				Delete:  true,
				Enabled: true,
			},
		},
		{
			name: "Explicit false values preserved",
			yamlData: `
name: "test_job"
source: "/source"
target: "/target"
delete: false
enabled: false
`,
			expected: Job{
				Name:    "test_job",
				Source:  "/source",
				Target:  "/target",
				Delete:  false,
				Enabled: false,
			},
		},
		{
			name: "Mixed explicit and default values",
			yamlData: `
name: "test_job"
source: "/source"
target: "/target"
delete: false
`,
			expected: Job{
				Name:    "test_job",
				Source:  "/source",
				Target:  "/target",
				Delete:  false,
				Enabled: true, // default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var job Job

			err := yaml.Unmarshal([]byte(tt.yamlData), &job)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			if !reflect.DeepEqual(job, tt.expected) {
				t.Errorf("got %+v, want %+v", job, tt.expected)
			}
		})
	}
}

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

func TestValidateJobNames(t *testing.T) {
	tests := []struct {
		name         string
		jobs         []Job
		expectsError bool
		errorMessage string
	}{
		{
			name: "Valid job names",
			jobs: []Job{
				{Name: "job1"},
				{Name: "job2"},
			},
			expectsError: false,
		},
		{
			name: "Duplicate job names",
			jobs: []Job{
				{Name: "job1"},
				{Name: "job1"},
			},
			expectsError: true,
			errorMessage: "duplicate job name: job1",
		},
		{
			name: "Invalid characters in job name",
			jobs: []Job{
				{Name: "job 1"},
			},
			expectsError: true,
			errorMessage: "invalid characters in job name: job 1",
		},
		{
			name: "Mixed errors",
			jobs: []Job{
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
		paths        []Path
		pathType     string
		jobName      string
		expectsError bool
		errorMessage string
	}{
		{
			name:         "Valid source path",
			jobPath:      "/home/user/documents",
			paths:        []Path{{Path: "/home/user"}},
			pathType:     "source",
			jobName:      "job1",
			expectsError: false,
		},
		{
			name:         "Invalid source path",
			jobPath:      "/invalid/source",
			paths:        []Path{{Path: "/home/user"}},
			pathType:     "source",
			jobName:      "job1",
			expectsError: true,
			errorMessage: "invalid source path for job 'job1': /invalid/source",
		},
		{
			name:         "Valid target path",
			jobPath:      "/mnt/backup/documents",
			paths:        []Path{{Path: "/mnt/backup"}},
			pathType:     "target",
			jobName:      "job1",
			expectsError: false,
		},
		{
			name:         "Invalid target path",
			jobPath:      "/invalid/target",
			paths:        []Path{{Path: "/mnt/backup"}},
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
		cfg          Config
		expectsError bool
		errorMessage string
	}{
		{
			name: "Valid paths",
			cfg: Config{
				Sources: []Path{
					{Path: "/home/user"},
				},
				Targets: []Path{
					{Path: "/mnt/backup"},
				},
				Jobs: []Job{
					{Name: "job1", Source: "/home/user/documents", Target: "/mnt/backup/documents"},
				},
			},
			expectsError: false,
		},
		{
			name: "Invalid paths",
			cfg: Config{
				Sources: []Path{
					{Path: "/home/user"},
				},
				Targets: []Path{
					{Path: "/mnt/backup"},
				},
				Jobs: []Job{
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
