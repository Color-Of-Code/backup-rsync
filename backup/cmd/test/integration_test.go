//go:build integration

package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"backup-rsync/backup/cmd"
	"backup-rsync/backup/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

// setupDirs creates source and target temp directories.
// Returns (sourceDir, targetDir).
func setupDirs(t *testing.T) (string, string) {
	t.Helper()

	base := t.TempDir()
	src := filepath.Join(base, "source")
	dst := filepath.Join(base, "target")

	require.NoError(t, os.MkdirAll(src, 0750))
	require.NoError(t, os.MkdirAll(dst, 0750))

	return src, dst
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0750))
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func fileExists(t *testing.T, path string) bool {
	t.Helper()

	_, err := os.Stat(path)

	return err == nil
}

func readFileContent(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

func executeIntegrationCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	rootCmd := cmd.BuildRootCommand()

	var stdout bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	return stdout.String(), err
}

// --- run: basic sync from source to target ---

func TestIntegration_Run_BasicSync(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "hello.txt"), "hello world")
	writeFile(t, filepath.Join(src, "subdir", "nested.txt"), "nested content")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("basic", "", "", testutil.Delete(false)).
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: basic")
	assert.Contains(t, stdout, "Status [basic]: SUCCESS")

	assert.Equal(t, "hello world", readFileContent(t, filepath.Join(dst, "hello.txt")))
	assert.Equal(t, "nested content", readFileContent(t, filepath.Join(dst, "subdir", "nested.txt")))
}

// --- run: idempotent second sync produces no changes ---

func TestIntegration_Run_IdempotentSync(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "data.txt"), "same content")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("idem", "", "").
		Build())

	// First sync
	_, err := executeIntegrationCommand(t, "run", "--config", cfgPath)
	require.NoError(t, err)

	// Second sync - should still succeed, nothing new to transfer
	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [idem]: SUCCESS")
	assert.Equal(t, "same content", readFileContent(t, filepath.Join(dst, "data.txt")))
}

// --- run: delete mode removes extra files from target ---

func TestIntegration_Run_DeleteRemovesExtraFiles(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "keep.txt"), "keep me")
	writeFile(t, filepath.Join(dst, "keep.txt"), "keep me")
	writeFile(t, filepath.Join(dst, "stale.txt"), "should be removed")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("cleanup", "", "", testutil.Delete(true)).
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [cleanup]: SUCCESS")

	assert.True(t, fileExists(t, filepath.Join(dst, "keep.txt")))
	assert.False(t, fileExists(t, filepath.Join(dst, "stale.txt")), "stale.txt should have been deleted")
}

// --- run: no-delete mode preserves extra files in target ---

func TestIntegration_Run_NoDeletePreservesExtraFiles(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "a.txt"), "a")
	writeFile(t, filepath.Join(dst, "extra.txt"), "should remain")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("nodelete", "", "", testutil.Delete(false)).
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [nodelete]: SUCCESS")

	assert.True(t, fileExists(t, filepath.Join(dst, "a.txt")))
	assert.True(t, fileExists(t, filepath.Join(dst, "extra.txt")), "extra.txt should be preserved")
}

// --- run: exclusions prevent syncing excluded paths ---

func TestIntegration_Run_Exclusions(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "docs", "readme.txt"), "documentation")
	writeFile(t, filepath.Join(src, "cache", "tmp.dat"), "temporary data")
	writeFile(t, filepath.Join(src, "logs", "app.log"), "log data")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("filtered", "", "", testutil.Exclusions("cache", "logs")).
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [filtered]: SUCCESS")

	assert.True(t, fileExists(t, filepath.Join(dst, "docs", "readme.txt")))
	assert.False(t, fileExists(t, filepath.Join(dst, "cache", "tmp.dat")), "cache should be excluded")
	assert.False(t, fileExists(t, filepath.Join(dst, "logs", "app.log")), "logs should be excluded")
}

// --- run: disabled job is skipped ---

func TestIntegration_Run_DisabledJobSkipped(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "file.txt"), "content")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("disabled-job", "", "", testutil.Enabled(false)).
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [disabled-job]: SKIPPED")
	assert.False(t, fileExists(t, filepath.Join(dst, "file.txt")), "disabled job should not sync files")
}

// --- run: multiple jobs with mixed outcomes ---

func TestIntegration_Run_MultipleJobs(t *testing.T) {
	base := t.TempDir()

	srcA := filepath.Join(base, "srcA")
	dstA := filepath.Join(base, "dstA")
	srcB := filepath.Join(base, "srcB")
	dstB := filepath.Join(base, "dstB")

	require.NoError(t, os.MkdirAll(srcA, 0750))
	require.NoError(t, os.MkdirAll(dstA, 0750))
	require.NoError(t, os.MkdirAll(srcB, 0750))
	require.NoError(t, os.MkdirAll(dstB, 0750))

	writeFile(t, filepath.Join(srcA, "a.txt"), "alpha")
	writeFile(t, filepath.Join(srcB, "b.txt"), "bravo")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("mA", srcA, dstA).
		AddJobToMapping("jobA", "", "").
		AddMapping("mB", srcB, dstB).
		AddJobToMapping("jobB", "", "").
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [jobA]: SUCCESS")
	assert.Contains(t, stdout, "Status [jobB]: SUCCESS")
	assert.Contains(t, stdout, "Summary: 2 succeeded, 0 failed, 0 skipped")

	assert.Equal(t, "alpha", readFileContent(t, filepath.Join(dstA, "a.txt")))
	assert.Equal(t, "bravo", readFileContent(t, filepath.Join(dstB, "b.txt")))
}

// --- run: partial changes — only modified files are synced ---

func TestIntegration_Run_PartialChanges(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "unchanged.txt"), "same")
	writeFile(t, filepath.Join(src, "modified.txt"), "original")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("partial", "", "").
		Build())

	// Initial sync
	_, err := executeIntegrationCommand(t, "run", "--config", cfgPath)
	require.NoError(t, err)

	assert.Equal(t, "original", readFileContent(t, filepath.Join(dst, "modified.txt")))

	// Modify source file
	writeFile(t, filepath.Join(src, "modified.txt"), "updated")

	// Second sync - should pick up the change
	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [partial]: SUCCESS")

	assert.Equal(t, "updated", readFileContent(t, filepath.Join(dst, "modified.txt")))
	assert.Equal(t, "same", readFileContent(t, filepath.Join(dst, "unchanged.txt")))
}

// --- simulate: dry-run does NOT modify target ---

func TestIntegration_Simulate_NoChanges(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "new.txt"), "should not appear in target")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("dryrun", "", "").
		Build())

	stdout, err := executeIntegrationCommand(t, "simulate", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: dryrun")
	assert.Contains(t, stdout, "Status [dryrun]: SUCCESS")
	assert.Contains(t, stdout, "--dry-run")

	assert.False(t, fileExists(t, filepath.Join(dst, "new.txt")),
		"simulate should not create files in target")
}

// --- simulate: shows what would be transferred ---

func TestIntegration_Simulate_ShowsChanges(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "report.txt"), "quarterly report")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("preview", "", "").
		Build())

	stdout, err := executeIntegrationCommand(t, "simulate", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: preview")
	assert.Contains(t, stdout, "--dry-run")
	assert.Contains(t, stdout, "Status [preview]: SUCCESS")
}

// --- simulate then run: simulate doesn't interfere with subsequent run ---

func TestIntegration_SimulateThenRun(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "data.txt"), "important data")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("workflow", "", "").
		Build())

	// Simulate first
	_, err := executeIntegrationCommand(t, "simulate", "--config", cfgPath)
	require.NoError(t, err)

	assert.False(t, fileExists(t, filepath.Join(dst, "data.txt")), "simulate should not modify target")

	// Now actually run
	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [workflow]: SUCCESS")
	assert.Equal(t, "important data", readFileContent(t, filepath.Join(dst, "data.txt")))
}

// --- list: lists commands without executing rsync ---

func TestIntegration_List_ShowsCommands(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "x.txt"), "x")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("listjob", "", "", testutil.Exclusions("temp")).
		Build())

	stdout, err := executeIntegrationCommand(t, "list", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Job: listjob")
	assert.Contains(t, stdout, "--exclude=temp")
	assert.Contains(t, stdout, src+"/")
	assert.Contains(t, stdout, dst)
	assert.NotContains(t, stdout, "Status [listjob]:")
	assert.NotContains(t, stdout, "Summary:")

	// list should not actually sync files
	assert.False(t, fileExists(t, filepath.Join(dst, "x.txt")), "list should not sync files")
}

// --- run: variable substitution works end-to-end ---

func TestIntegration_Run_VariableSubstitution(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "v.txt"), "vars work")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Variable("src_dir", src).Variable("dst_dir", dst).
		AddMapping("m", "${src_dir}", "${dst_dir}").
		AddJobToMapping("var-job", "", "").
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [var-job]: SUCCESS")
	assert.Equal(t, "vars work", readFileContent(t, filepath.Join(dst, "v.txt")))
}

// --- run: mixed enabled/disabled/multiple results with summary ---

func TestIntegration_Run_MixedJobsSummary(t *testing.T) {
	base := t.TempDir()

	srcOK := filepath.Join(base, "srcOK")
	dstOK := filepath.Join(base, "dstOK")
	srcSkip := filepath.Join(base, "srcSkip")
	dstSkip := filepath.Join(base, "dstSkip")

	require.NoError(t, os.MkdirAll(srcOK, 0750))
	require.NoError(t, os.MkdirAll(dstOK, 0750))
	require.NoError(t, os.MkdirAll(srcSkip, 0750))
	require.NoError(t, os.MkdirAll(dstSkip, 0750))

	writeFile(t, filepath.Join(srcOK, "ok.txt"), "ok")
	writeFile(t, filepath.Join(srcSkip, "skip.txt"), "skip")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("mOK", srcOK, dstOK).
		AddJobToMapping("active", "", "", testutil.Enabled(true)).
		AddMapping("mSkip", srcSkip, dstSkip).
		AddJobToMapping("inactive", "", "", testutil.Enabled(false)).
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [active]: SUCCESS")
	assert.Contains(t, stdout, "Status [inactive]: SKIPPED")
	assert.Contains(t, stdout, "Summary: 1 succeeded, 0 failed, 1 skipped")

	assert.True(t, fileExists(t, filepath.Join(dstOK, "ok.txt")))
	assert.False(t, fileExists(t, filepath.Join(dstSkip, "skip.txt")))
}

// --- run: empty source directory syncs nothing ---

func TestIntegration_Run_EmptySource(t *testing.T) {
	src, dst := setupDirs(t)

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("empty", "", "").
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [empty]: SUCCESS")
}

// --- run: deep directory hierarchy is synced correctly ---

func TestIntegration_Run_DeepHierarchy(t *testing.T) {
	src, dst := setupDirs(t)

	writeFile(t, filepath.Join(src, "a", "b", "c", "d", "deep.txt"), "deep file")

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("deep", "", "").
		Build())

	stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Status [deep]: SUCCESS")
	assert.Equal(t, "deep file", readFileContent(t, filepath.Join(dst, "a", "b", "c", "d", "deep.txt")))
}

// --- check-coverage: fully covered config reports no uncovered paths ---

func TestIntegration_CheckCoverage_FullCoverage(t *testing.T) {
	src := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(src, "docs"), 0750))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "photos"), 0750))

	dst := t.TempDir()

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("docs", "docs", "docs").
		AddJobToMapping("photos", "photos", "photos").
		Build())

	stdout, err := executeIntegrationCommand(t, "check-coverage", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Uncovered paths:")
	// Both subdirectories are covered, so no uncovered paths should appear after the header
	lines := splitNonEmpty(stdout)
	assert.Equal(t, 1, len(lines), "only the header line expected; got: %v", lines)
}

// --- check-coverage: incomplete coverage reports uncovered paths ---

func TestIntegration_CheckCoverage_IncompleteCoverage(t *testing.T) {
	src := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(src, "docs"), 0750))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "music"), 0750))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "videos"), 0750))

	dst := t.TempDir()

	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", src, dst).
		AddJobToMapping("docs-only", "docs", "docs").
		Build())

	stdout, err := executeIntegrationCommand(t, "check-coverage", "--config", cfgPath)

	require.NoError(t, err)
	// The source root itself should be reported as uncovered (since music and videos aren't covered)
	assert.Contains(t, stdout, src)
}

// --- config show: end-to-end with variable resolution ---

func TestIntegration_ConfigShow(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		Variable("base", "/backup").
		AddMapping("m", "/data", "${base}").
		AddJobToMapping("resolved", "files", "files").
		Build())

	stdout, err := executeIntegrationCommand(t, "config", "show", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "/backup/files")
	assert.Contains(t, stdout, "resolved")
}

// --- config validate: valid config passes ---

func TestIntegration_ConfigValidate_Valid(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", "/data", "/backup").
		AddJobToMapping("valid", "stuff", "stuff").
		Build())

	stdout, err := executeIntegrationCommand(t, "config", "validate", "--config", cfgPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Configuration is valid.")
}

// --- config validate: overlapping sources are rejected ---

func TestIntegration_ConfigValidate_OverlappingSources(t *testing.T) {
	cfgPath := testutil.WriteConfigFile(t, testutil.NewConfigBuilder().
		AddMapping("m", "/data", "/backup").
		AddJobToMapping("parent", "user", "user").
		AddJobToMapping("child", "user/docs", "docs").
		Build())

	_, err := executeIntegrationCommand(t, "config", "validate", "--config", cfgPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validating config")
}

// --- version: real rsync version output ---

func TestIntegration_Version(t *testing.T) {
	stdout, err := executeIntegrationCommand(t, "version")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Rsync Binary Path: /usr/bin/rsync")
	assert.Contains(t, stdout, "Version Info:")
	assert.Contains(t, stdout, "rsync")
}

// splitNonEmpty splits a string by newlines and returns non-empty trimmed lines.
func splitNonEmpty(s string) []string {
	var result []string

	for _, line := range bytes.Split([]byte(s), []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 {
			result = append(result, string(trimmed))
		}
	}

	return result
}
