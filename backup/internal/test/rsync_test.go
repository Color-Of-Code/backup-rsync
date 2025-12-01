package internal_test

import (
	. "backup-rsync/backup/internal"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errCommandNotFound = errors.New("command not found")

const rsyncPath = "/usr/bin/rsync"

func TestArgumentsForJob(t *testing.T) {
	job := NewJob(
		WithSource("/home/user/Music/"),
		WithTarget("/target/user/music/home"),
		WithExclusions([]string{"*.tmp", "node_modules/"}),
	)
	args := ArgumentsForJob(job, "", true)

	expectedArgs := []string{
		"--dry-run", "-aiv", "--stats", "--delete",
		"--exclude=*.tmp", "--exclude=node_modules/",
		"/home/user/Music/", "/target/user/music/home",
	}

	assert.Equal(t, strings.Join(expectedArgs, " "), strings.Join(args, " "))
}

func TestGetVersionInfo_Success(t *testing.T) {
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell: &MockExec{
			Output: "rsync  version 3.2.3  protocol version 31\n",
			Error:  nil,
		},
	}

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.NoError(t, err)
	assert.Equal(t, rsyncPath, fullpath)
	assert.Equal(t, "rsync  version 3.2.3  protocol version 31\n", versionInfo)
}

func TestGetVersionInfo_CommandError(t *testing.T) {
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell: &MockExec{
			Output: "",
			Error:  errCommandNotFound,
		},
	}

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	assert.Empty(t, fullpath)
	assert.Empty(t, versionInfo)
}

func TestGetVersionInfo_InvalidOutput(t *testing.T) {
	rsync := SharedCommand{
		BinPath: rsyncPath,
		Shell: &MockExec{
			Output: "invalid output",
			Error:  nil,
		},
	}

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	assert.Empty(t, fullpath)
	assert.Empty(t, versionInfo)
}

func TestGetVersionInfo_EmptyPath(t *testing.T) {
	rsync := SharedCommand{
		BinPath: "",
		Shell: &MockExec{
			Output: "",
			Error:  nil,
		},
	}

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"\"")
	assert.Empty(t, versionInfo)
	assert.Empty(t, fullpath)
}

func TestGetVersionInfo_IncompletePath(t *testing.T) {
	rsync := SharedCommand{
		BinPath: "bin/rsync",
		Shell: &MockExec{
			Output: "",
			Error:  nil,
		},
	}

	versionInfo, fullpath, err := rsync.GetVersionInfo()

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"bin/rsync\"")
	assert.Empty(t, versionInfo)
	assert.Empty(t, fullpath)
}
