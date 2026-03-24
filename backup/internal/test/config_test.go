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

mappings:
  - name: "test"
    source: "/home/test"
    target: "${target_base}/test"
    jobs:
      - name: "test_job"
        source: ""
        target: ""
        enabled: true
`
	reader := bytes.NewReader([]byte(yamlData))

	cfg, err := LoadConfig(reader)
	require.NoError(t, err)

	assert.Equal(t, "/mnt/backup1", cfg.Variables["target_base"])
	assert.Len(t, cfg.Mappings, 1)
	assert.Len(t, cfg.Mappings[0].Jobs, 1)

	job := cfg.Mappings[0].Jobs[0]
	assert.Equal(t, "test_job", job.Name)
}

func TestLoadConfig2(t *testing.T) {
	yamlData := `
mappings:
  - name: "m1"
    source: "/source"
    target: "/target"
    jobs:
      - name: "job1"
        source: "a"
        target: "a"
      - name: "job2"
        source: "b"
        target: "b"
        delete: false
        enabled: false
`

	reader := bytes.NewReader([]byte(yamlData))

	cfg, err := LoadConfig(reader)
	require.NoError(t, err)

	require.Len(t, cfg.Mappings, 1)
	jobs := cfg.Mappings[0].Jobs
	require.Len(t, jobs, 2)

	assert.Equal(t, "job1", jobs[0].Name)
	assert.True(t, jobs[0].Delete)
	assert.True(t, jobs[0].Enabled)

	assert.Equal(t, "job2", jobs[1].Name)
	assert.False(t, jobs[1].Delete)
	assert.False(t, jobs[1].Enabled)
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

func TestConfigString_ValidConfig(t *testing.T) {
	cfg := Config{
		Mappings: []Mapping{},
	}

	actualOutput := cfg.String()
	assert.Contains(t, actualOutput, "mappings: []")
}

func TestResolveConfig(t *testing.T) {
	cfg := Config{
		Variables: map[string]string{
			"source_base": "/home/user",
			"target_base": "/backup/user",
		},
		Mappings: []Mapping{
			{
				Name:   "data",
				Source: "${source_base}",
				Target: "${target_base}",
				Jobs: []Job{
					{Name: "job1", Source: "Documents", Target: "Documents", Enabled: true, Delete: true},
					{Name: "job2", Source: "Pictures", Target: "Pictures", Enabled: true, Delete: true},
				},
			},
		},
	}

	resolvedCfg, err := ResolveConfig(cfg)
	require.NoError(t, err)

	allJobs := resolvedCfg.AllJobs()
	assert.Equal(t, "/home/user/Documents/", allJobs[0].Source)
	assert.Equal(t, "/backup/user/Documents", allJobs[0].Target)
	assert.Equal(t, "/home/user/Pictures/", allJobs[1].Source)
	assert.Equal(t, "/backup/user/Pictures", allJobs[1].Target)
}

func TestLoadResolvedConfig(t *testing.T) {
	tests := []struct {
		name, config, wantErr, wantTarget string
		wantJobs                          int
	}{
		{name: "FileNotFound", wantErr: "failed to open config"},
		{name: "InvalidYAML", config: "{{invalid yaml", wantErr: "failed to parse YAML"},
		{name: "DuplicateJobNames", wantErr: "duplicate job name: dup",
			config: testutil.NewConfigBuilder().
				AddMapping("m", "/src", "/tgt").
				AddJobToMapping("dup", "a", "a").AddJobToMapping("dup", "b", "b").Build()},
		{name: "OverlappingSourcePaths", wantErr: "job source path validation failed",
			config: testutil.NewConfigBuilder().
				AddMapping("m", "/home", "/backup").
				AddJobToMapping("parent", "user", "user").
				AddJobToMapping("child", "user/docs", "docs").Build()},
		{name: "OverlappingAllowedByExclusion", wantJobs: 2,
			config: testutil.NewConfigBuilder().
				AddMapping("m", "/home", "/backup").
				AddJobToMapping("parent", "user", "user", testutil.Exclusions("docs")).
				AddJobToMapping("child", "user/docs", "docs").Build()},
		{name: "OverlappingTargetPaths", wantErr: "job target path validation failed",
			config: testutil.NewConfigBuilder().
				AddMapping("m", "/home", "/backup").
				AddJobToMapping("job1", "docs", "all").
				AddJobToMapping("job2", "photos", "all/photos").Build()},
		{name: "ValidConfig", wantJobs: 1, wantTarget: "/backup/docs",
			config: testutil.NewConfigBuilder().
				Variable("base", "/backup").
				AddMapping("m", "/home", "${base}").
				AddJobToMapping("docs", "docs", "docs").Build()},
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

			allJobs := cfg.AllJobs()
			if test.wantJobs > 0 {
				assert.Len(t, allJobs, test.wantJobs)
			}

			if test.wantTarget != "" {
				assert.Equal(t, test.wantTarget, allJobs[0].Target)
			}
		})
	}
}

func TestConfigApply_VersionInfoSuccess(t *testing.T) {
	mockCmd := NewMockJobCommand(t)

	var logBuf bytes.Buffer

	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		Mappings: []Mapping{
			{
				Name:   "test",
				Source: "/src",
				Target: "/dst",
				Jobs: []Job{
					{Name: "job1", Source: "/src/a/", Target: "/dst/a", Enabled: true},
					{Name: "job2", Source: "/src/b/", Target: "/dst/b", Enabled: false},
				},
			},
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
		Mappings: []Mapping{
			{
				Name:   "test",
				Source: "/data",
				Target: "/bak",
				Jobs: []Job{
					{Name: "backup", Source: "/data/stuff/", Target: "/bak/stuff", Enabled: true},
				},
			},
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
			"user": "alice",
		},
		Mappings: []Mapping{
			{
				Name:   "${user}_home",
				Source: "/home/${user}",
				Target: "/mnt/backup1/${user}",
				Jobs: []Job{
					{
						Name:   "${user}_docs",
						Source: "Documents",
						Target: "documents",
					},
				},
			},
		},
	}

	resolved, err := ResolveConfig(cfg)

	require.NoError(t, err)
	assert.Equal(t, "alice_home", resolved.Mappings[0].Name)
	assert.Equal(t, "/home/alice", resolved.Mappings[0].Source)
	assert.Equal(t, "/mnt/backup1/alice", resolved.Mappings[0].Target)

	allJobs := resolved.AllJobs()
	assert.Equal(t, "alice_docs", allJobs[0].Name)
	assert.Equal(t, "/home/alice/Documents/", allJobs[0].Source)
	assert.Equal(t, "/mnt/backup1/alice/documents", allJobs[0].Target)
}

func TestResolveConfig_VariableChaining(t *testing.T) {
	cfg := Config{
		Variables: map[string]string{
			"user":        "bob",
			"source_home": "/home/${user}",
			"target_base": "/mnt/backup1/${user}",
		},
		Mappings: []Mapping{
			{
				Name:   "${user}_home",
				Source: "${source_home}",
				Target: "${target_base}",
				Jobs: []Job{
					{
						Name:   "${user}_mail",
						Source: ".thunderbird",
						Target: "mail",
					},
				},
			},
		},
	}

	resolved, err := ResolveConfig(cfg)

	require.NoError(t, err)
	assert.Equal(t, "/home/bob", resolved.Mappings[0].Source)
	assert.Equal(t, "/mnt/backup1/bob", resolved.Mappings[0].Target)

	allJobs := resolved.AllJobs()
	assert.Equal(t, "bob_mail", allJobs[0].Name)
	assert.Equal(t, "/home/bob/.thunderbird/", allJobs[0].Source)
	assert.Equal(t, "/mnt/backup1/bob/mail", allJobs[0].Target)
}

func TestResolveConfig_MappingSourceMacroError(t *testing.T) {
	cfg := Config{
		Mappings: []Mapping{
			{
				Name:   "test",
				Source: "/home/@{bogus:val}",
				Target: "/dst",
				Jobs:   []Job{{Name: "job1", Source: "a", Target: "a"}},
			},
		},
	}

	_, err := ResolveConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving mapping source")
}

func TestResolveConfig_MappingTargetMacroError(t *testing.T) {
	cfg := Config{
		Mappings: []Mapping{
			{
				Name:   "test",
				Source: "/src",
				Target: "/backup/@{bogus:val}",
				Jobs:   []Job{{Name: "job1", Source: "a", Target: "a"}},
			},
		},
	}

	_, err := ResolveConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving mapping target")
}

func TestResolveConfig_MappingNameMacroError(t *testing.T) {
	cfg := Config{
		Mappings: []Mapping{
			{
				Name:   "@{bogus:val}",
				Source: "/src",
				Target: "/dst",
				Jobs:   []Job{{Name: "job1", Source: "a", Target: "a"}},
			},
		},
	}

	_, err := ResolveConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving mapping name")
}

func TestLoadResolvedConfig_WithOverrides(t *testing.T) {
	config := testutil.NewConfigBuilder().
		Variable("user", "default").
		AddMapping("home", "/home/${user}", "/mnt/backup1/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path, map[string]string{"user": "alice"})

	require.NoError(t, err)

	allJobs := cfg.AllJobs()
	assert.Equal(t, "alice_docs", allJobs[0].Name)
	assert.Equal(t, "/home/alice/docs/", allJobs[0].Source)
	assert.Equal(t, "/mnt/backup1/alice/docs", allJobs[0].Target)
}

func TestLoadResolvedConfig_OverridesNewVariable(t *testing.T) {
	config := testutil.NewConfigBuilder().
		AddMapping("home", "/home/${user}", "/mnt/backup1/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path, map[string]string{"user": "bob"})

	require.NoError(t, err)

	allJobs := cfg.AllJobs()
	assert.Equal(t, "bob_docs", allJobs[0].Name)
	assert.Equal(t, "/home/bob/docs/", allJobs[0].Source)
	assert.Equal(t, "/mnt/backup1/bob/docs", allJobs[0].Target)
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
		Variable("user", "alice").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("docs", "docs", "docs").
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
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path, map[string]string{"user": "alice", "user_cap": "Alice"})

	require.NoError(t, err)
	assert.Equal(t, "alice_docs", cfg.AllJobs()[0].Name)
}

func TestLoadResolvedConfig_TemplateVarsAllInYAML(t *testing.T) {
	config := testutil.NewConfigBuilder().
		TemplateVar("user").
		Variable("user", "bob").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()

	path := testutil.WriteConfigFile(t, config)

	cfg, err := LoadResolvedConfig(path)

	require.NoError(t, err)
	assert.Equal(t, "bob_docs", cfg.AllJobs()[0].Name)
}

// --- LoadResolvedConfig with include ---

func TestLoadResolvedConfig_BasicInclude(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)

	allJobs := cfg.AllJobs()
	require.Len(t, allJobs, 1)
	assert.Equal(t, "alice_docs", allJobs[0].Name)
	assert.Equal(t, "/home/alice/docs/", allJobs[0].Source)
	assert.Equal(t, "/backup/alice/docs", allJobs[0].Target)
}

func TestLoadResolvedConfig_MultipleIncludes(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		AddInclude("template.yaml", map[string]string{"user": "bob"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)

	allJobs := cfg.AllJobs()
	require.Len(t, allJobs, 2)
	assert.Equal(t, "alice_docs", allJobs[0].Name)
	assert.Equal(t, "bob_docs", allJobs[1].Name)
}

func TestLoadResolvedConfig_IncludeMissingTemplateVars(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").TemplateVar("user_cap").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
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
		AddMapping("m", "/src", "/dst").
		AddJobToMapping("inner", "a", "a").
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
		Variable("home", "/home/${user}").
		AddMapping("home", "${home}", "/backup/${user}").
		AddJobToMapping("${user}_mail", ".thunderbird", "mail").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)

	allJobs := cfg.AllJobs()
	require.Len(t, allJobs, 1)
	assert.Equal(t, "alice_mail", allJobs[0].Name)
	assert.Equal(t, "/home/alice/.thunderbird/", allJobs[0].Source)
	assert.Equal(t, "/backup/alice/mail", allJobs[0].Target)
}

func TestLoadResolvedConfig_IncludeWithOverridesOnMainConfig(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		Variable("target_root", "backup").
		AddMapping("home", "/home/${user}", "/${target_root}/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)
	assert.Equal(t, "/backup/alice/docs", cfg.AllJobs()[0].Target)
}

func TestLoadResolvedConfig_IncludeMergesMappings(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
		Build()
	testutil.WriteConfigFileInDir(t, dir, "template.yaml", template)

	main := testutil.NewConfigBuilder().
		AddMapping("shared", "/shared", "/shared-backup").
		AddJobToMapping("shared_data", "data", "data").
		AddInclude("template.yaml", map[string]string{"user": "alice"}).
		Build()
	mainPath := testutil.WriteConfigFileInDir(t, dir, "main.yaml", main)

	cfg, err := LoadResolvedConfig(mainPath)

	require.NoError(t, err)

	allJobs := cfg.AllJobs()
	require.Len(t, allJobs, 2)

	// Main config's own mappings + included template's mappings
	assert.Len(t, cfg.Mappings, 2)
}

func TestLoadResolvedConfig_IncludeMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "@{bogus:val}", "docs").
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

func TestLoadResolvedConfig_IncludeTemplateMappingSourceMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "@{bogus:val}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
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

func TestLoadResolvedConfig_IncludeTemplateMappingTargetMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "@{bogus:val}").
		AddJobToMapping("${user}_docs", "docs", "docs").
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

func TestLoadResolvedConfig_IncludeTemplateMappingNameMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("@{bogus:val}", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "docs").
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

func TestConfigBuilder_AddMappingWithExclusions(t *testing.T) {
	yaml := testutil.NewConfigBuilder().
		AddMappingWithExclusions("m", "/src", "/tgt", "cache", "tmp").
		AddJobToMapping("j", "docs", "docs").
		Build()

	assert.Contains(t, yaml, "exclusions:")
	assert.Contains(t, yaml, "cache")
	assert.Contains(t, yaml, "tmp")
}

func TestLoadResolvedConfig_IncludeTemplateJobNameMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("@{bogus:val}", "docs", "docs").
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

func TestLoadResolvedConfig_IncludeTemplateJobTargetMacroError(t *testing.T) {
	dir := t.TempDir()

	template := testutil.NewConfigBuilder().
		TemplateVar("user").
		AddMapping("home", "/home/${user}", "/backup/${user}").
		AddJobToMapping("${user}_docs", "docs", "@{bogus:val}").
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

func TestResolveConfig_JobNameMacroError(t *testing.T) {
	cfg := Config{
		Mappings: []Mapping{{Name: "m", Source: "/src", Target: "/tgt",
			Jobs: []Job{{Name: "@{bogus:val}", Source: "", Target: ""}}}},
	}

	_, err := ResolveConfig(cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving job")
}

// --- AllJobs helper ---

func TestAllJobs(t *testing.T) {
	cfg := Config{
		Mappings: []Mapping{
			{Name: "m1", Source: "/s1", Target: "/t1", Jobs: []Job{{Name: "j1"}, {Name: "j2"}}},
			{Name: "m2", Source: "/s2", Target: "/t2", Jobs: []Job{{Name: "j3"}}},
		},
	}

	allJobs := cfg.AllJobs()
	require.Len(t, allJobs, 3)
	assert.Equal(t, "j1", allJobs[0].Name)
	assert.Equal(t, "j2", allJobs[1].Name)
	assert.Equal(t, "j3", allJobs[2].Name)
}
