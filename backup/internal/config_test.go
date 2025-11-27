package internal_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"backup-rsync/backup/internal"
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

	cfg, err := internal.LoadConfig(reader)
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

	cfg, err := internal.LoadConfig(reader)
	require.NoError(t, err)

	expected := []internal.Job{
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
	expected := internal.Job{
		Name:    "test_job",
		Source:  "/source",
		Target:  "/target",
		Delete:  true,
		Enabled: true,
	}

	var job internal.Job

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
	expected := internal.Job{
		Name:    "test_job",
		Source:  "/source",
		Target:  "/target",
		Delete:  false,
		Enabled: false,
	}

	var job internal.Job

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
	expected := internal.Job{
		Name:    "test_job",
		Source:  "/source",
		Target:  "/target",
		Delete:  false,
		Enabled: true, // default
	}

	var job internal.Job

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

	result := internal.SubstituteVariables(input, variables)
	assert.Equal(t, expected, result, "SubstituteVariables result mismatch")
}

func TestValidateJobNames_ValidJobNames(t *testing.T) {
	jobs := []internal.Job{
		{Name: "job1"},
		{Name: "job2"},
	}

	err := internal.ValidateJobNames(jobs)
	assert.NoError(t, err)
}

func TestValidateJobNames_DuplicateJobNames(t *testing.T) {
	jobs := []internal.Job{
		{Name: "job1"},
		{Name: "job1"},
	}

	err := internal.ValidateJobNames(jobs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate job name: job1")
}

func TestValidateJobNames_InvalidCharactersInJobName(t *testing.T) {
	jobs := []internal.Job{
		{Name: "job 1"},
	}

	err := internal.ValidateJobNames(jobs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters in job name: job 1")
}

func TestValidateJobNames_MixedErrors(t *testing.T) {
	jobs := []internal.Job{
		{Name: "job1"},
		{Name: "job 1"},
		{Name: "job1"},
	}

	err := internal.ValidateJobNames(jobs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate job name: job1")
}

func TestValidatePath_ValidSourcePath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []internal.Path
		pathType string
	}{
		jobPath:  "/home/user/documents",
		paths:    []internal.Path{{Path: "/home/user"}},
		pathType: "source",
	}

	err := internal.ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	assert.NoError(t, err)
}

func TestValidatePath_InvalidSourcePath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []internal.Path
		pathType string
	}{
		jobPath:  "/invalid/source",
		paths:    []internal.Path{{Path: "/home/user"}},
		pathType: "source",
	}

	err := internal.ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	require.Error(t, err)
	assert.EqualError(t, err, "invalid path for job 'job1': source /invalid/source")
}

func TestValidatePath_ValidTargetPath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []internal.Path
		pathType string
	}{
		jobPath:  "/mnt/backup/documents",
		paths:    []internal.Path{{Path: "/mnt/backup"}},
		pathType: "target",
	}

	err := internal.ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	assert.NoError(t, err)
}

func TestValidatePath_InvalidTargetPath(t *testing.T) {
	test := struct {
		jobPath  string
		paths    []internal.Path
		pathType string
	}{
		jobPath:  "/invalid/target",
		paths:    []internal.Path{{Path: "/mnt/backup"}},
		pathType: "target",
	}

	err := internal.ValidatePath(test.jobPath, test.paths, test.pathType, "job1")

	require.Error(t, err)
	assert.EqualError(t, err, "invalid path for job 'job1': target /invalid/target")
}

func TestValidatePaths_ValidPaths(t *testing.T) {
	test := struct {
		name         string
		cfg          internal.Config
		expectsError bool
	}{
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
	}

	t.Run(test.name, func(t *testing.T) {
		err := internal.ValidatePaths(test.cfg)
		assert.NoError(t, err)
	})
}

func TestValidatePaths_InvalidPaths(t *testing.T) {
	test := struct {
		name         string
		cfg          internal.Config
		expectsError bool
		errorMessage string
	}{
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
		errorMessage: "path validation failed: [" +
			"invalid path for job 'job1': source /invalid/source " +
			"invalid path for job 'job1': target /invalid/target]",
	}

	t.Run(test.name, func(t *testing.T) {
		err := internal.ValidatePaths(test.cfg)
		require.Error(t, err)
		assert.EqualError(t, err, test.errorMessage)
	})
}
