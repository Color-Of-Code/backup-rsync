package internal_test

import (
	. "backup-rsync/backup/internal"
	"errors"
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

func TestGetVersionInfo_Success(t *testing.T) {
	mockExec := NewMockExec(t)
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell:   mockExec,
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
	}

	// No expectations set - should fail before calling Execute due to path validation

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"bin/rsync\"")
	assert.Empty(t, versionInfo)
	assert.Empty(t, fullpath)
}

func TestNewSimulateCommand_BaseLogPath_ShallHaveSimSuffix(t *testing.T) {
	binPath := "/usr/bin/simulate"
	logPath := "/var/log/simulate"

	simulateCmd := NewSimulateCommand(binPath, logPath)

	assert.Equal(t, logPath+"-sim", simulateCmd.BaseLogPath)
}
