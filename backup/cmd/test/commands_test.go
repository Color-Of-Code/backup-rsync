package cmd_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"backup-rsync/backup/cmd"
	"backup-rsync/backup/internal"
	"backup-rsync/backup/internal/testutil"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubExec struct {
	output []byte
	err    error
}

func (s *stubExec) Execute(_ string, _ ...string) ([]byte, error) {
	return s.output, s.err
}

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	return executeCommandWithFs(t, afero.NewMemMapFs(), args...)
}

func executeCommandWithFs(t *testing.T, fs afero.Fs, args ...string) (string, error) {
	t.Helper()

	rootCmd := cmd.BuildRootCommandWithFs(fs)

	var stdout bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	return stdout.String(), err
}

func executeCommandWithDeps(t *testing.T, fs afero.Fs, shell internal.Exec, args ...string) (string, error) {
	t.Helper()

	rootCmd := cmd.BuildRootCommandWithDeps(fs, shell)

	var stdout bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	return stdout.String(), err
}

// --- config show ---

func TestConfigShow_ValidConfig(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("docs", "/home/docs/", "/backup/docs/").
		Build())

	stdout, err := executeCommand(t, "config", "show", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "docs")
	assert.Contains(t, stdout, "/home/docs")
	assert.Contains(t, stdout, "/backup/docs")
}

func TestConfigShow_MissingFile(t *testing.T) {
	_, err := executeCommand(t, "config", "show", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

// --- missing config (shared pattern) ---

func TestMissingConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"list", "--config", "/nonexistent/config.yaml"}, "loading config"},
		{"run", []string{"run", "--config", "/nonexistent/config.yaml"}, "loading config"},
		{"simulate", []string{"simulate", "--config", "/nonexistent/config.yaml"}, "loading config"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := executeCommand(t, test.args...)

			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

// --- create logger error (shared pattern) ---

func TestCreateLoggerError(t *testing.T) {
	commands := []string{"run", "simulate"}

	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
				Source("/home").Target("/backup").
				AddJob("docs", "/home/docs/", "/backup/docs/").
				Build())

			fs := afero.NewReadOnlyFs(afero.NewMemMapFs())

			shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}

			_, err := executeCommandWithDeps(t, fs, shell, command, "--config", cfgPath)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "creating logger")
		})
	}
}

func TestConfigShow_InvalidYAML(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, `{{{invalid yaml`)

	_, err := executeCommand(t, "config", "show", "--config", cfgPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

// --- config validate ---

func TestConfigValidate_ValidConfig(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("docs", "/home/docs/", "/backup/docs/").
		Build())

	stdout, err := executeCommand(t, "config", "validate", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Configuration is valid.")
}

func TestConfigValidate_MissingFile(t *testing.T) {
	_, err := executeCommand(t, "config", "validate", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validating config")
}

func TestConfigValidate_DuplicateJobNames(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("same", "/home/a/", "/backup/a/").
		AddJob("same", "/home/b/", "/backup/b/").
		Build())

	_, err := executeCommand(t, "config", "validate", "--config", cfgPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validating config")
}

// --- run ---

func TestRun_ValidConfig(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("docs", "/home/docs/", "/backup/docs/", testutil.Enabled(true), testutil.Delete(true)).
		Build())

	shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}

	stdout, err := executeCommandWithDeps(t, afero.NewMemMapFs(), shell, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: docs")
	assert.Contains(t, stdout, "Status [docs]: SUCCESS")
}

// --- simulate ---

func TestSimulate_ValidConfig(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("docs", "/home/docs/", "/backup/docs/", testutil.Enabled(true)).
		Build())

	shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}

	stdout, err := executeCommandWithDeps(t, afero.NewMemMapFs(), shell, "simulate", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: docs")
	assert.Contains(t, stdout, "Status [docs]: SUCCESS")
}

// --- version ---

func TestVersion_ErrorPaths(t *testing.T) {
	tests := []struct {
		name, rsyncPath string
	}{
		{"InvalidRsyncPath", "not-absolute"},
		{"NonExistentRsyncPath", "/nonexistent/rsync"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := executeCommand(t, "version", "--rsync-path", test.rsyncPath)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "getting version info")
		})
	}
}

// --- check-coverage ---

func TestCheckCoverage_MissingConfig(t *testing.T) {
	_, err := executeCommand(t, "check-coverage", "--config", "/nonexistent/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading config")
}

func TestCheckCoverage_WithUncoveredPaths(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/src").Target("/dst").
		AddJob("docs", "/src/docs/", "/dst/docs/").
		Build())

	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/src/docs", 0755)
	_ = fs.MkdirAll("/src/photos", 0755)

	stdout, err := executeCommandWithFs(t, fs, "check-coverage", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Uncovered paths:")
	assert.Contains(t, stdout, "/src")
}

func TestCheckCoverage_ValidConfig(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/src").Target("/dst").
		AddJob("docs", "/src/docs/", "/dst/docs/").
		Build())

	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/src/docs", 0755)

	stdout, err := executeCommandWithFs(t, fs, "check-coverage", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Uncovered paths:")
}

// --- version (positive path) ---

func TestVersion_ValidRsync(t *testing.T) {
	// Only run if rsync is actually installed
	_, err := os.Stat("/usr/bin/rsync")
	if os.IsNotExist(err) {
		t.Skip("rsync not installed")
	}

	stdout, err := executeCommand(t, "version", "--rsync-path", "/usr/bin/rsync")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Rsync Binary Path: /usr/bin/rsync")
	assert.Contains(t, stdout, "Version Info:")
}

func TestVersion_WithMockExec(t *testing.T) {
	shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}

	stdout, err := executeCommandWithDeps(t, afero.NewMemMapFs(), shell, "version", "--rsync-path", "/usr/bin/rsync")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Rsync Binary Path: /usr/bin/rsync")
	assert.Contains(t, stdout, "rsync version 3.2.7")
}

// --- list (positive path) ---

func TestList_ValidConfig(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("docs", "/home/docs/", "/backup/docs/", testutil.Enabled(true)).
		Build())

	shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}

	stdout, err := executeCommandWithDeps(t, afero.NewMemMapFs(), shell, "list", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: docs")
	assert.Contains(t, stdout, "Status [docs]: SUCCESS")
}

// --- run: logger cleanup happens after cfg.Apply completes ---

func TestRun_LoggerOpenDuringApply(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Source("/home").Target("/backup").
		AddJob("docs", "/home/docs/", "/backup/docs/", testutil.Enabled(true)).
		Build())

	shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}
	fs := afero.NewMemMapFs()

	stdout, err := executeCommandWithDeps(t, fs, shell, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [docs]: SUCCESS")

	// Walk the in-memory filesystem to find the summary log written by the logger.
	// cfg.Apply writes "STATUS [docs]: SUCCESS" via the logger after the if-block
	// where defer cleanup() is registered. If cleanup ran too early (closing the
	// log file before Apply), the log file would be empty or missing this entry.
	var summaryContent string

	_ = afero.Walk(fs, "logs", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, "summary.log") {
			data, readErr := afero.ReadFile(fs, path)
			require.NoError(t, readErr)

			summaryContent = string(data)
		}

		return nil
	})

	require.NotEmpty(t, summaryContent, "summary.log should have been created")
	assert.Contains(t, summaryContent, "STATUS [docs]: SUCCESS",
		"logger must remain open during cfg.Apply — proves defer cleanup() is function-scoped")
}
