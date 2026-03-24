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

func TestResolveVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name:     "NoReferences",
			input:    map[string]string{"a": "hello", "b": "world"},
			expected: map[string]string{"a": "hello", "b": "world"},
		},
		{
			name:     "SingleLevel",
			input:    map[string]string{"user": "alice", "home": "/home/${user}"},
			expected: map[string]string{"user": "alice", "home": "/home/alice"},
		},
		{
			name: "MultiLevel",
			input: map[string]string{
				"user":    "bob",
				"home":    "/home/${user}",
				"docs":    "${home}/Documents",
				"archive": "${docs}/archive",
			},
			expected: map[string]string{
				"user":    "bob",
				"home":    "/home/bob",
				"docs":    "/home/bob/Documents",
				"archive": "/home/bob/Documents/archive",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ResolveVariables(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestResolveVariables_CircularReference(t *testing.T) {
	result := ResolveVariables(map[string]string{"a": "${b}", "b": "${a}"})

	// Circular references leave unresolved placeholders — exact values
	// depend on map iteration order, so just verify they remain unresolved.
	assert.Contains(t, result["a"], "${")
	assert.Contains(t, result["b"], "${")
}

func TestResolveConfig_ResolvesAllFields(t *testing.T) {
	cfg := Config{
		Variables: map[string]string{
			"user":     "alice",
			"user_cap": "Jaap",
		},
		Sources: []Path{{Path: "/home/${user}/"}},
		Targets: []Path{{Path: "/mnt/backup1/${user}"}},
		Jobs: []Job{
			{
				Name:   "${user}_docs",
				Source: "/home/${user}/Documents/",
				Target: "/mnt/backup1/${user}/documents",
			},
		},
	}

	resolved, err := ResolveConfig(cfg)

	require.NoError(t, err)
	assert.Equal(t, "/home/alice/", resolved.Sources[0].Path)
	assert.Equal(t, "/mnt/backup1/alice", resolved.Targets[0].Path)
	assert.Equal(t, "alice_docs", resolved.Jobs[0].Name)
	assert.Equal(t, "/home/alice/Documents/", resolved.Jobs[0].Source)
	assert.Equal(t, "/mnt/backup1/alice/documents", resolved.Jobs[0].Target)
}

func TestResolveConfig_VariableChaining(t *testing.T) {
	cfg := Config{
		Variables: map[string]string{
			"user":        "bob",
			"source_home": "/home/${user}",
			"target_base": "/mnt/backup1/${user}",
		},
		Sources: []Path{{Path: "/home/${user}/"}},
		Targets: []Path{{Path: "/mnt/backup1/${user}"}},
		Jobs: []Job{
			{
				Name:   "${user}_mail",
				Source: "${source_home}/.thunderbird/",
				Target: "${target_base}/mail",
			},
		},
	}

	resolved, err := ResolveConfig(cfg)

	require.NoError(t, err)
	assert.Equal(t, "/home/bob/", resolved.Sources[0].Path)
	assert.Equal(t, "/mnt/backup1/bob", resolved.Targets[0].Path)
	assert.Equal(t, "bob_mail", resolved.Jobs[0].Name)
	assert.Equal(t, "/home/bob/.thunderbird/", resolved.Jobs[0].Source)
	assert.Equal(t, "/mnt/backup1/bob/mail", resolved.Jobs[0].Target)
}

func TestResolveConfig_SourceMacroError(t *testing.T) {
	cfg := Config{
		Sources: []Path{{Path: "/home/@{bogus:val}/"}},
		Jobs:    []Job{{Name: "job1", Source: "/src/", Target: "/dst/"}},
	}

	_, err := ResolveConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving source path")
}

func TestResolveConfig_TargetMacroError(t *testing.T) {
	cfg := Config{
		Targets: []Path{{Path: "/backup/@{bogus:val}/"}},
		Jobs:    []Job{{Name: "job1", Source: "/src/", Target: "/dst/"}},
	}

	_, err := ResolveConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving target path")
}

func TestLoadResolvedConfig_WithOverrides(t *testing.T) {
	config := testutil.NewConfigBuilder().
		Source("/home/${user}").
		Target("/mnt/backup1/${user}").
		Variable("user", "default").
		AddJob("${user}_docs", "/home/${user}/docs/", "/mnt/backup1/${user}/docs/").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path, map[string]string{"user": "alice"})

	require.NoError(t, err)
	assert.Equal(t, "alice_docs", cfg.Jobs[0].Name)
	assert.Equal(t, "/home/alice/docs/", cfg.Jobs[0].Source)
	assert.Equal(t, "/mnt/backup1/alice/docs/", cfg.Jobs[0].Target)
}

func TestLoadResolvedConfig_OverridesNewVariable(t *testing.T) {
	config := testutil.NewConfigBuilder().
		Source("/home/${user}").
		Target("/mnt/backup1/${user}").
		AddJob("${user}_docs", "/home/${user}/docs/", "/mnt/backup1/${user}/docs/").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path, map[string]string{"user": "bob"})

	require.NoError(t, err)
	assert.Equal(t, "bob_docs", cfg.Jobs[0].Name)
	assert.Equal(t, "/home/bob/docs/", cfg.Jobs[0].Source)
	assert.Equal(t, "/mnt/backup1/bob/docs/", cfg.Jobs[0].Target)
}

// --- ValidateTemplateVars ---

func TestValidateTemplateVars(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "NoTemplate",
			cfg:  Config{},
		},
		{
			name: "EmptyVariablesList",
			cfg:  Config{Template: &Template{Variables: []string{}}},
		},
		{
			name: "AllVariablesProvided",
			cfg: Config{
				Template:  &Template{Variables: []string{"user", "user_cap"}},
				Variables: map[string]string{"user": "alice", "user_cap": "Alice"},
			},
		},
		{
			name: "OneMissing",
			cfg: Config{
				Template:  &Template{Variables: []string{"user", "user_cap"}},
				Variables: map[string]string{"user": "alice"},
			},
			wantErr: "missing required template variables: [user_cap]",
		},
		{
			name: "AllMissing",
			cfg: Config{
				Template:  &Template{Variables: []string{"user", "user_cap"}},
				Variables: map[string]string{},
			},
			wantErr: "missing required template variables",
		},
		{
			name: "NilVariablesMap",
			cfg: Config{
				Template: &Template{Variables: []string{"user"}},
			},
			wantErr: "missing required template variables: [user]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTemplateVars(test.cfg)

			if test.wantErr != "" {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrMissingTemplateVars)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- LoadResolvedConfig with template: variables ---

func TestLoadResolvedConfig_TemplateVarsMissing(t *testing.T) {
	config := testutil.NewConfigBuilder().
		TemplateVar("user").TemplateVar("user_cap").
		Source("/home/${user}").Target("/backup/${user}").
		Variable("user", "alice").
		AddJob("docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()

	path := testutil.WriteConfigFile(t, config)

	_, err := LoadResolvedConfig(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "template validation failed")
	assert.Contains(t, err.Error(), "user_cap")
}

func TestLoadResolvedConfig_TemplateVarsProvidedViaOverride(t *testing.T) {
	config := testutil.NewConfigBuilder().
		TemplateVar("user").TemplateVar("user_cap").
		Source("/home/${user}").Target("/backup/${user}").
		AddJob("${user}_docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path, map[string]string{"user": "alice", "user_cap": "Alice"})

	require.NoError(t, err)
	assert.Equal(t, "alice_docs", cfg.Jobs[0].Name)
}

func TestLoadResolvedConfig_TemplateVarsAllInYAML(t *testing.T) {
	config := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/backup/${user}").
		Variable("user", "bob").
		AddJob("${user}_docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path)

	require.NoError(t, err)
	assert.Equal(t, "bob_docs", cfg.Jobs[0].Name)
}

// --- LoadResolvedConfig with include ---

func TestLoadResolvedConfig_BasicInclude(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/backup/${user}").
		AddJob("${user}_docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)
	require.Len(t, cfg.Jobs, 1)
	assert.Equal(t, "alice_docs", cfg.Jobs[0].Name)
	assert.Equal(t, "/home/alice/docs/", cfg.Jobs[0].Source)
	assert.Equal(t, "/backup/alice/docs/", cfg.Jobs[0].Target)
}

func TestLoadResolvedConfig_MultipleIncludes(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/backup/${user}").
		AddJob("${user}_docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		AddInclude("template.yaml", map[string]string{"user": "bob"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)
	require.Len(t, cfg.Jobs, 2)
	assert.Equal(t, "alice_docs", cfg.Jobs[0].Name)
	assert.Equal(t, "bob_docs", cfg.Jobs[1].Name)
}

func TestLoadResolvedConfig_IncludeMissingTemplateVars(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").TemplateVar("user_cap").
		Source("/home/${user}").Target("/backup/${user}").
		AddJob("${user}_docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	// Only provides "user" but template requires "user" and "user_cap"
	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	_, err := LoadResolvedConfig(mainPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expanding includes")
	assert.Contains(t, err.Error(), "missing required template variables")
	assert.Contains(t, err.Error(), "user_cap")
}

func TestLoadResolvedConfig_IncludeFileNotFound(t *testing.T) {
	dir := t.TempDir()

	main := testutil.NewConfigBuilder().
		AddInclude("nonexistent.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	_, err := LoadResolvedConfig(mainPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expanding includes")
	assert.Contains(t, err.Error(), "failed to open")
}

func TestLoadResolvedConfig_NestedIncludesRejected(t *testing.T) {
	dir := t.TempDir()

	// inner template itself has an include (nested)
	inner := testutil.NewConfigBuilder().
		AddInclude("other.yaml", map[string]string{"x": "y"}).
		Source("/src").Target("/dst").
		AddJob("inner", "/src/a/", "/dst/a/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "inner.yaml", inner)

	main := testutil.NewConfigBuilder().
		AddInclude("inner.yaml", map[string]string{}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	_, err := LoadResolvedConfig(mainPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expanding includes")
	assert.Contains(t, err.Error(), "nested includes are not supported")
}

func TestLoadResolvedConfig_IncludeInvalidYAML(t *testing.T) {
	dir := t.TempDir()

	testutil.WriteConfigFileInDir(t, dir, "bad.yaml", "{{{not valid yaml")

	main := testutil.NewConfigBuilder().
		AddInclude("bad.yaml", map[string]string{}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	_, err := LoadResolvedConfig(mainPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expanding includes")
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestLoadResolvedConfig_IncludeWithVariableChaining(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/backup/${user}").
		Variable("home", "/home/${user}").
		AddJob("${user}_mail", "${home}/.thunderbird/", "/backup/${user}/mail").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)
	require.Len(t, cfg.Jobs, 1)
	assert.Equal(t, "alice_mail", cfg.Jobs[0].Name)
	assert.Equal(t, "/home/alice/.thunderbird/", cfg.Jobs[0].Source)
	assert.Equal(t, "/backup/alice/mail", cfg.Jobs[0].Target)
}

func TestLoadResolvedConfig_IncludeWithOverridesOnMainConfig(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/${target_root}/${user}").
		Variable("target_root", "backup").
		AddJob("${user}_docs", "/home/${user}/docs/", "/${target_root}/${user}/docs/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)
	assert.Equal(t, "/backup/alice/docs/", cfg.Jobs[0].Target)
}

func TestLoadResolvedConfig_IncludeMergesSources(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/backup/${user}").
		AddJob("${user}_docs", "/home/${user}/docs/", "/backup/${user}/docs/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		Source("/shared").Target("/shared-backup").
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		AddJob("shared_data", "/shared/data/", "/shared-backup/data/").
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)
	require.Len(t, cfg.Jobs, 2)

	// Main config's own sources + included template's sources
	assert.Len(t, cfg.Sources, 2)
	assert.Len(t, cfg.Targets, 2)
}

func TestLoadResolvedConfig_IncludeMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Source("/home/${user}").Target("/backup/${user}").
		AddJob("${user}_docs", "/home/@{bogus:val}/", "/backup/${user}/docs/").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	_, err := LoadResolvedConfig(mainPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expanding includes")
}
