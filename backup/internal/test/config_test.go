package internal_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	. "backup-rsync/backup/internal"
	"backup-rsync/backup/internal/testutil"
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
	require.NoError(t, err)

	assert.Equal(t, "/mnt/backup1", cfg.Variables["target_base"])
	assert.Len(t, cfg.Jobs, 1)

	job := cfg.Jobs[0]
	assert.Equal(t, "test_job", job.Name)
	assert.Equal(t, "/home/test/", job.Source)
	assert.Equal(t, "${target_base}/test/", job.Target)
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

	reader := bytes.NewReader([]byte(yamlData))

	cfg, err := LoadConfig(reader)
	require.NoError(t, err)

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

	for i, job := range cfg.Jobs {
		assert.Equal(t, expected[i], job, "Job mismatch at index %d", i)
	}
}

func TestYAMLUnmarshalingDefaults(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected Job
	}{
		{
			name: "FieldsOmitted",
			yaml: `
name: "test_job"
source: "/source"
target: "/target"
`,
			expected: Job{
				Name: "test_job", Source: "/source", Target: "/target",
				Delete: true, Enabled: true,
			},
		},
		{
			name: "ExplicitFalseValues",
			yaml: `
name: "test_job"
source: "/source"
target: "/target"
delete: false
enabled: false
`,
			expected: Job{
				Name: "test_job", Source: "/source", Target: "/target",
				Delete: false, Enabled: false,
			},
		},
		{
			name: "MixedValues",
			yaml: `
name: "test_job"
source: "/source"
target: "/target"
delete: false
`,
			expected: Job{
				Name: "test_job", Source: "/source", Target: "/target",
				Delete: false, Enabled: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var job Job

			err := yaml.Unmarshal([]byte(test.yaml), &job)
			require.NoError(t, err)
			assert.Equal(t, test.expected, job)
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	variables := map[string]string{
		"target_base": "/mnt/backup1",
	}
	input := "${target_base}/user/music/home"
	expected := "/mnt/backup1/user/music/home"

	result := SubstituteVariables(input, variables)
	assert.Equal(t, expected, result, "SubstituteVariables result mismatch")
}

func TestValidateJobNames(t *testing.T) {
	tests := []struct {
		name    string
		jobs    []Job
		wantErr string
	}{
		{"ValidJobNames", []Job{{Name: "job1"}, {Name: "job2"}}, ""},
		{"DuplicateJobNames", []Job{{Name: "job1"}, {Name: "job1"}}, "duplicate job name: job1"},
		{"InvalidCharacters", []Job{{Name: "job 1"}}, "invalid characters in job name: job 1"},
		{"MixedErrors", []Job{{Name: "job1"}, {Name: "job 1"}, {Name: "job1"}}, "duplicate job name: job1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateJobNames(test.jobs)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name     string
		jobPath  string
		paths    []Path
		pathType string
		wantErr  string
	}{
		{
			name:     "ValidSourcePath",
			jobPath:  "/home/user/documents",
			paths:    []Path{{Path: "/home/user"}},
			pathType: "source",
		},
		{
			name:     "InvalidSourcePath",
			jobPath:  "/invalid/source",
			paths:    []Path{{Path: "/home/user"}},
			pathType: "source",
			wantErr:  "invalid path for job 'job1': source /invalid/source",
		},
		{
			name:     "ValidTargetPath",
			jobPath:  "/mnt/backup/documents",
			paths:    []Path{{Path: "/mnt/backup"}},
			pathType: "target",
		},
		{
			name:     "InvalidTargetPath",
			jobPath:  "/invalid/target",
			paths:    []Path{{Path: "/mnt/backup"}},
			pathType: "target",
			wantErr:  "invalid path for job 'job1': target /invalid/target",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

			if test.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePaths(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "ValidPaths",
			cfg: Config{
				Sources: []Path{{Path: "/home/user"}},
				Targets: []Path{{Path: "/mnt/backup"}},
				Jobs:    []Job{{Name: "job1", Source: "/home/user/documents", Target: "/mnt/backup/documents"}},
			},
		},
		{
			name: "InvalidPaths",
			cfg: Config{
				Sources: []Path{{Path: "/home/user"}},
				Targets: []Path{{Path: "/mnt/backup"}},
				Jobs:    []Job{{Name: "job1", Source: "/invalid/source", Target: "/invalid/target"}},
			},
			wantErr: "path validation failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePaths(test.cfg)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigString_ValidConfig(t *testing.T) {
	cfg := Config{
		Sources:   []Path{},
		Targets:   []Path{},
		Variables: map[string]string{},
		Jobs:      []Job{},
	}

	expectedOutput := "sources: []\ntargets: []\nvariables: {}\njobs: []\n"
	actualOutput := cfg.String()

	assert.Equal(t, expectedOutput, actualOutput)
}

func TestResolveConfig(t *testing.T) {
	cfg := Config{
		Variables: map[string]string{
			"source_base": "/home/user",
			"target_base": "/backup/user",
		},
		Jobs: []Job{
			{
				Name:   "job1",
				Source: "${source_base}/Documents",
				Target: "${target_base}/Documents",
			},
			{
				Name:   "job2",
				Source: "${source_base}/Pictures",
				Target: "${target_base}/Pictures",
			},
		},
	}

	resolvedCfg, err := ResolveConfig(cfg)
	require.NoError(t, err)

	assert.Equal(t, "/home/user/Documents", resolvedCfg.Jobs[0].Source)
	assert.Equal(t, "/backup/user/Documents", resolvedCfg.Jobs[0].Target)
	assert.Equal(t, "/home/user/Pictures", resolvedCfg.Jobs[1].Source)
	assert.Equal(t, "/backup/user/Pictures", resolvedCfg.Jobs[1].Target)
}

func TestLoadResolvedConfig(t *testing.T) {
	tests := []struct {
		name, config, wantErr, wantTarget string
		wantJobs                          int
	}{
		{name: "FileNotFound", wantErr: "failed to open config"},
		{name: "InvalidYAML", config: "{{invalid yaml", wantErr: "failed to parse YAML"},
		{name: "DuplicateJobNames", wantErr: "duplicate job name: dup",
			config: testutil.NewConfigBuilder().Source("/src").Target("/tgt").
				AddJob("dup", "/src/a", "/tgt/a").AddJob("dup", "/src/b", "/tgt/b").Build()},
		{name: "InvalidSourcePath", wantErr: "path validation failed",
			config: testutil.NewConfigBuilder().Source("/home").Target("/backup").
				AddJob("job1", "/invalid/source", "/backup/stuff").Build()},
		{name: "OverlappingSourcePaths", wantErr: "job source path validation failed",
			config: testutil.NewConfigBuilder().Source("/home").Target("/backup").
				AddJob("parent", "/home/user", "/backup/user").
				AddJob("child", "/home/user/docs", "/backup/docs").Build()},
		{name: "OverlappingAllowedByExclusion", wantJobs: 2,
			config: testutil.NewConfigBuilder().Source("/home").Target("/backup").
				AddJob("parent", "/home/user", "/backup/user", testutil.Exclusions("docs")).
				AddJob("child", "/home/user/docs", "/backup/docs").Build()},
		{name: "OverlappingTargetPaths", wantErr: "job target path validation failed",
			config: testutil.NewConfigBuilder().Source("/home").Target("/backup").
				AddJob("job1", "/home/docs", "/backup/all").
				AddJob("job2", "/home/photos", "/backup/all/photos").Build()},
		{name: "ValidConfig", wantJobs: 1, wantTarget: "/backup/docs",
			config: testutil.NewConfigBuilder().Source("/home").Target("/backup").
				Variable("base", "/backup").AddJob("docs", "/home/docs", "${base}/docs").Build()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := "/nonexistent/path/config.yaml"
			if test.config != "" {
				path = testutil.WriteConfigFile(t, test.config)
			}

			cfg, err := LoadResolvedConfig(path)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)

				return
			}

			require.NoError(t, err)

			if test.wantJobs > 0 {
				assert.Len(t, cfg.Jobs, test.wantJobs)
			}

			if test.wantTarget != "" {
				assert.Equal(t, test.wantTarget, cfg.Jobs[0].Target)
			}
		})
	}
}

func TestConfigApply_VersionInfoSuccess(t *testing.T) {
	mockCmd := NewMockJobCommand(t)

	var logBuf bytes.Buffer

	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		Jobs: []Job{
			{Name: "job1", Source: "/src/", Target: "/dst/", Enabled: true},
			{Name: "job2", Source: "/src2/", Target: "/dst2/", Enabled: false},
		},
	}

	mockCmd.EXPECT().GetVersionInfo().Return("rsync version 3.2.3", "/usr/bin/rsync", nil).Once()
	mockCmd.EXPECT().Run(mock.AnythingOfType("internal.Job")).Return(Success).Once()
	mockCmd.EXPECT().ReportJobStatus("job1", Success, logger).Once()
	mockCmd.EXPECT().ReportJobStatus("job2", Skipped, logger).Once()
	mockCmd.EXPECT().ReportSummary(map[JobStatus]int{Success: 1, Skipped: 1}, logger).Once()

	err := cfg.Apply(mockCmd, logger)

	require.NoError(t, err)
	assert.Contains(t, logBuf.String(), "Rsync Binary Path: /usr/bin/rsync")
	assert.Contains(t, logBuf.String(), "Rsync Version Info: rsync version 3.2.3")
}

func TestConfigApply_VersionInfoError(t *testing.T) {
	mockCmd := NewMockJobCommand(t)

	var logBuf bytes.Buffer

	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		Jobs: []Job{
			{Name: "backup", Source: "/data/", Target: "/bak/", Enabled: true},
		},
	}

	mockCmd.EXPECT().GetVersionInfo().Return("", "", errCommandNotFound).Once()
	mockCmd.EXPECT().Run(mock.AnythingOfType("internal.Job")).Return(Failure).Once()
	mockCmd.EXPECT().ReportJobStatus("backup", Failure, logger).Once()
	mockCmd.EXPECT().ReportSummary(map[JobStatus]int{Failure: 1}, logger).Once()

	err := cfg.Apply(mockCmd, logger)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrJobFailure)
	assert.Contains(t, logBuf.String(), "Failed to fetch rsync version: command not found")
	assert.NotContains(t, logBuf.String(), "Rsync Binary Path")
}
