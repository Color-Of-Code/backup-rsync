package internal_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	. "backup-rsync/backup/internal"
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

func TestYAMLUnmarshalingDefaults_FieldsOmitted(t *testing.T) {
	yamlData := `
name: "test_job"
source: "/source"
target: "/target"
`
	expected := Job{
		Name:    "test_job",
		Source:  "/source",
		Target:  "/target",
		Delete:  true,
		Enabled: true,
	}

	var job Job

	err := yaml.Unmarshal([]byte(yamlData), &job)
	require.NoError(t, err)
	assert.Equal(t, expected, job)
}

func TestYAMLUnmarshalingDefaults_ExplicitFalseValues(t *testing.T) {
	yamlData := `
name: "test_job"
source: "/source"
target: "/target"
delete: false
enabled: false
`
	expected := Job{
		Name:    "test_job",
		Source:  "/source",
		Target:  "/target",
		Delete:  false,
		Enabled: false,
	}

	var job Job

	err := yaml.Unmarshal([]byte(yamlData), &job)
	require.NoError(t, err)
	assert.Equal(t, expected, job)
}

func TestYAMLUnmarshalingDefaults_MixedValues(t *testing.T) {
	yamlData := `
name: "test_job"
source: "/source"
target: "/target"
delete: false
`
	expected := Job{
		Name:    "test_job",
		Source:  "/source",
		Target:  "/target",
		Delete:  false,
		Enabled: true, // default
	}

	var job Job

	err := yaml.Unmarshal([]byte(yamlData), &job)
	require.NoError(t, err)
	assert.Equal(t, expected, job)
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

func TestValidateJobNames_ValidJobNames(t *testing.T) {
	jobs := []Job{
		{Name: "job1"},
		{Name: "job2"},
	}

	err := ValidateJobNames(jobs)
	assert.NoError(t, err)
}

func TestValidateJobNames_DuplicateJobNames(t *testing.T) {
	jobs := []Job{
		{Name: "job1"},
		{Name: "job1"},
	}

	err := ValidateJobNames(jobs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate job name: job1")
}

func TestValidateJobNames_InvalidCharactersInJobName(t *testing.T) {
	jobs := []Job{
		{Name: "job 1"},
	}

	err := ValidateJobNames(jobs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters in job name: job 1")
}

func TestValidateJobNames_MixedErrors(t *testing.T) {
	jobs := []Job{
		{Name: "job1"},
		{Name: "job 1"},
		{Name: "job1"},
	}

	err := ValidateJobNames(jobs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate job name: job1")
}

func TestValidatePath_ValidSourcePath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []Path
		pathType string
	}{
		jobPath:  "/home/user/documents",
		paths:    []Path{{Path: "/home/user"}},
		pathType: "source",
	}

	err := ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	assert.NoError(t, err)
}

func TestValidatePath_InvalidSourcePath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []Path
		pathType string
	}{
		jobPath:  "/invalid/source",
		paths:    []Path{{Path: "/home/user"}},
		pathType: "source",
	}

	err := ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	require.Error(t, err)
	assert.EqualError(t, err, "invalid path for job 'job1': source /invalid/source")
}

func TestValidatePath_ValidTargetPath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []Path
		pathType string
	}{
		jobPath:  "/mnt/backup/documents",
		paths:    []Path{{Path: "/mnt/backup"}},
		pathType: "target",
	}

	err := ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	assert.NoError(t, err)
}

func TestValidatePath_InvalidTargetPath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []Path
		pathType string
	}{
		jobPath:  "/invalid/target",
		paths:    []Path{{Path: "/mnt/backup"}},
		pathType: "target",
	}

	err := ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	require.Error(t, err)
	assert.EqualError(t, err, "invalid path for job 'job1': target /invalid/target")
}

func TestValidatePaths_ValidPaths(t *testing.T) {
	test := struct {
		name         string
		cfg          Config
		expectsError bool
	}{
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
	}

	t.Run(test.name, func(t *testing.T) {
		err := ValidatePaths(test.cfg)
		assert.NoError(t, err)
	})
}

func TestValidatePaths_InvalidPaths(t *testing.T) {
	test := struct {
		name         string
		cfg          Config
		expectsError bool
		errorMessage string
	}{
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
		errorMessage: "path validation failed: [" +
			"invalid path for job 'job1': source /invalid/source " +
			"invalid path for job 'job1': target /invalid/target]",
	}

	t.Run(test.name, func(t *testing.T) {
		err := ValidatePaths(test.cfg)
		require.Error(t, err)
		assert.EqualError(t, err, test.errorMessage)
	})
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

	resolvedCfg := ResolveConfig(cfg)

	assert.Equal(t, "/home/user/Documents", resolvedCfg.Jobs[0].Source)
	assert.Equal(t, "/backup/user/Documents", resolvedCfg.Jobs[0].Target)
	assert.Equal(t, "/home/user/Pictures", resolvedCfg.Jobs[1].Source)
	assert.Equal(t, "/backup/user/Pictures", resolvedCfg.Jobs[1].Target)
}

// writeTestConfig writes YAML content to a temp file and returns its path.
func writeTestConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)

	return path
}

func TestLoadResolvedConfig_FileNotFound(t *testing.T) {
	_, err := LoadResolvedConfig("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open config")
}

func TestLoadResolvedConfig_InvalidYAML(t *testing.T) {
	path := writeTestConfig(t, "{{invalid yaml")

	_, err := LoadResolvedConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestLoadResolvedConfig_DuplicateJobNames(t *testing.T) {
	yaml := `
sources:
  - path: "/src"
targets:
  - path: "/tgt"
jobs:
  - name: "dup"
    source: "/src/a"
    target: "/tgt/a"
  - name: "dup"
    source: "/src/b"
    target: "/tgt/b"
`
	path := writeTestConfig(t, yaml)

	_, err := LoadResolvedConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job validation failed")
	assert.Contains(t, err.Error(), "duplicate job name: dup")
}

func TestLoadResolvedConfig_InvalidSourcePath(t *testing.T) {
	yaml := `
sources:
  - path: "/home"
targets:
  - path: "/backup"
jobs:
  - name: "job1"
    source: "/invalid/source"
    target: "/backup/stuff"
`
	path := writeTestConfig(t, yaml)

	_, err := LoadResolvedConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path validation failed")
}

func TestLoadResolvedConfig_OverlappingSourcePaths(t *testing.T) {
	yaml := `
sources:
  - path: "/home"
targets:
  - path: "/backup"
jobs:
  - name: "parent"
    source: "/home/user"
    target: "/backup/user"
  - name: "child"
    source: "/home/user/docs"
    target: "/backup/docs"
`
	path := writeTestConfig(t, yaml)

	_, err := LoadResolvedConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job source path validation failed")
}

func TestLoadResolvedConfig_ValidConfig(t *testing.T) {
	yaml := `
sources:
  - path: "/home"
targets:
  - path: "/backup"
variables:
  base: "/backup"
jobs:
  - name: "docs"
    source: "/home/docs"
    target: "${base}/docs"
`
	path := writeTestConfig(t, yaml)

	cfg, err := LoadResolvedConfig(path)
	require.NoError(t, err)
	assert.Len(t, cfg.Jobs, 1)
	assert.Equal(t, "/backup/docs", cfg.Jobs[0].Target)
}
