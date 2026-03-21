package internal_test

import (
	. "backup-rsync/backup/internal"
	"backup-rsync/backup/internal/testutil"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errCommandNotFound = errors.New("command not found")

const rsyncPath = "/usr/bin/rsync"

func TestArgumentsForJob(t *testing.T) {
	job := Job{
		Delete:     true,
		Source:     "/home/user/Music/",
		Target:     "/target/user/music/home",
		Exclusions: []string{"*.tmp", "node_modules/"},
	}
	args := ArgumentsForJob(job, "", true)

	expectedArgs := []string{
		"--dry-run", "-aiv", "--stats", "--delete",
		"--exclude=*.tmp", "--exclude=node_modules/",
		"/home/user/Music/", "/target/user/music/home",
	}

	assert.Equal(t, strings.Join(expectedArgs, " "), strings.Join(args, " "))
}

func TestArgumentsForJob_WithLogPath_(t *testing.T) {
	job := Job{
		Delete:     false,
		Source:     "/home/user/Documents/",
		Target:     "/backup/user/documents",
		Exclusions: []string{"*.log", "temp/"},
	}
	args := ArgumentsForJob(job, "/var/log/rsync.log", false)

	expectedArgs := []string{
		"-aiv", "--stats",
		"--log-file=/var/log/rsync.log",
		"--exclude=*.log", "--exclude=temp/",
		"/home/user/Documents/", "/backup/user/documents",
	}

	assert.Equal(t, strings.Join(expectedArgs, " "), strings.Join(args, " "))
}

func TestGetVersionInfo_Success(t *testing.T) {
	mockExec := NewMockExec(t)
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell:   mockExec,
		Output:  io.Discard,
	}

	// Set expectation for Execute call
	mockExec.EXPECT().Execute(rsyncPath, mock.MatchedBy(func(args []string) bool {
		return len(args) == 1 && args[0] == RsyncVersionFlag
	})).Return([]byte("rsync  version 3.2.3  protocol version 31\n"), nil).Once()

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.NoError(t, err)
	assert.Equal(t, rsyncPath, fullpath)
	assert.Equal(t, "rsync  version 3.2.3  protocol version 31\n", versionInfo)
}

func TestGetVersionInfo_CommandError(t *testing.T) {
	mockExec := NewMockExec(t)
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell:   mockExec,
		Output:  io.Discard,
	}

	// Set expectation for Execute call to return error
	mockExec.EXPECT().Execute(rsyncPath, mock.MatchedBy(func(args []string) bool {
		return len(args) == 1 && args[0] == RsyncVersionFlag
	})).Return(nil, errCommandNotFound).Once()

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	assert.Empty(t, fullpath)
	assert.Empty(t, versionInfo)
}

func TestGetVersionInfo_InvalidOutput(t *testing.T) {
	mockExec := NewMockExec(t)
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell:   mockExec,
		Output:  io.Discard,
	}

	// Set expectation for Execute call to return invalid output
	mockExec.EXPECT().Execute(rsyncPath, mock.MatchedBy(func(args []string) bool {
		return len(args) == 1 && args[0] == RsyncVersionFlag
	})).Return([]byte("invalid output"), nil).Once()

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	assert.Empty(t, fullpath)
	assert.Empty(t, versionInfo)
}

func TestGetVersionInfo_EmptyPath(t *testing.T) {
	mockExec := NewMockExec(t)
	rsync := SharedCommand{
		BinPath: "",
		Shell:   mockExec,
		Output:  io.Discard,
	}

	// No expectations set - should fail before calling Execute due to path validation

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"\"")
	assert.Empty(t, versionInfo)
	assert.Empty(t, fullpath)
}

func TestGetVersionInfo_IncompletePath(t *testing.T) {
	mockExec := NewMockExec(t)
	rsync := SharedCommand{
		BinPath: "bin/rsync",
		Shell:   mockExec,
		Output:  io.Discard,
	}

	// No expectations set - should fail before calling Execute due to path validation

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"bin/rsync\"")
	assert.Empty(t, versionInfo)
	assert.Empty(t, fullpath)
}

func TestNewSharedCommand(t *testing.T) {
	mockExec := NewMockExec(t)
	cmd := NewSharedCommand(rsyncPath, "/logs/base", mockExec, io.Discard)

	assert.Equal(t, rsyncPath, cmd.BinPath)
	assert.Equal(t, "/logs/base", cmd.BaseLogPath)
	assert.Equal(t, mockExec, cmd.Shell)
	assert.Equal(t, io.Discard, cmd.Output)
}

func TestJobLogPath(t *testing.T) {
	cmd := NewSharedCommand(rsyncPath, "/logs/sync-2025", nil, io.Discard)
	job := testutil.NewTestJob()

	logPath := cmd.JobLogPath(job)

	assert.Equal(t, "/logs/sync-2025/job-test-job.log", logPath)
}

func TestNewListCommand(t *testing.T) {
	mockExec := NewMockExec(t)
	cmd := NewListCommand(rsyncPath, mockExec, io.Discard)

	assert.Equal(t, rsyncPath, cmd.BinPath)
	assert.Empty(t, cmd.BaseLogPath)
	assert.Equal(t, mockExec, cmd.Shell)
}

func TestListCommand_Run_ReturnsSuccess(t *testing.T) {
	mockExec := NewMockExec(t)

	var buf bytes.Buffer

	cmd := NewListCommand(rsyncPath, mockExec, &buf)
	job := testutil.NewTestJob()

	status := cmd.Run(job)

	assert.Equal(t, Success, status)
	assert.Contains(t, buf.String(), "Job: test-job")
	assert.Contains(t, buf.String(), rsyncPath)
}

func TestNewSyncCommand(t *testing.T) {
	mockExec := NewMockExec(t)
	cmd := NewSyncCommand(rsyncPath, "/logs/base", mockExec, io.Discard)

	assert.Equal(t, rsyncPath, cmd.BinPath)
	assert.Equal(t, "/logs/base", cmd.BaseLogPath)
	assert.Equal(t, mockExec, cmd.Shell)
}

func TestSyncCommand_Run_Success(t *testing.T) {
	mockExec := NewMockExec(t)

	var buf bytes.Buffer

	cmd := NewSyncCommand(rsyncPath, "/logs/base", mockExec, &buf)
	job := testutil.NewTestJob()

	mockExec.EXPECT().Execute(rsyncPath, mock.AnythingOfType("[]string")).
		Return([]byte("sync output"), nil).Once()

	status := cmd.Run(job)

	assert.Equal(t, Success, status)
	assert.Contains(t, buf.String(), "Job: test-job")
	assert.Contains(t, buf.String(), "Output:\nsync output")
}

func TestSyncCommand_Run_Failure(t *testing.T) {
	mockExec := NewMockExec(t)
	cmd := NewSyncCommand(rsyncPath, "/logs/base", mockExec, io.Discard)
	job := testutil.NewTestJob()

	mockExec.EXPECT().Execute(rsyncPath, mock.AnythingOfType("[]string")).
		Return(nil, errCommandNotFound).Once()

	status := cmd.Run(job)

	assert.Equal(t, Failure, status)
}

func TestNewSimulateCommand(t *testing.T) {
	mockExec := NewMockExec(t)
	cmd := NewSimulateCommand(rsyncPath, "/logs/base", mockExec, io.Discard)

	assert.Equal(t, rsyncPath, cmd.BinPath)
	assert.Equal(t, "/logs/base", cmd.BaseLogPath)
	assert.Equal(t, mockExec, cmd.Shell)
}

func TestSimulateCommand_Run_Success(t *testing.T) {
	mockExec := NewMockExec(t)
	logDir := t.TempDir()

	var buf bytes.Buffer

	cmd := NewSimulateCommand(rsyncPath, logDir, mockExec, &buf)
	job := testutil.NewTestJob()

	mockExec.EXPECT().Execute(rsyncPath, mock.AnythingOfType("[]string")).
		Return([]byte("simulated output"), nil).Once()

	status := cmd.Run(job)

	assert.Equal(t, Success, status)
	assert.Contains(t, buf.String(), "Job: test-job")
}

func TestSimulateCommand_Run_Failure(t *testing.T) {
	mockExec := NewMockExec(t)
	logDir := t.TempDir()
	cmd := NewSimulateCommand(rsyncPath, logDir, mockExec, io.Discard)
	job := testutil.NewTestJob()

	mockExec.EXPECT().Execute(rsyncPath, mock.AnythingOfType("[]string")).
		Return(nil, errCommandNotFound).Once()

	status := cmd.Run(job)

	assert.Equal(t, Failure, status)
}

func TestSimulateCommand_Run_LogWriteError(t *testing.T) {
	mockExec := NewMockExec(t)

	var buf bytes.Buffer

	// Use a non-existent directory so WriteFile fails
	cmd := NewSimulateCommand(rsyncPath, "/nonexistent/path", mockExec, &buf)
	job := testutil.NewTestJob()

	mockExec.EXPECT().Execute(rsyncPath, mock.AnythingOfType("[]string")).
		Return([]byte("output"), nil).Once()

	status := cmd.Run(job)

	assert.Equal(t, Success, status)
	assert.Contains(t, buf.String(), "Warning: Failed to write output to log file")
}
