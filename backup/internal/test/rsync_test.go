package internal_test

import (
	"backup-rsync/backup/internal"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errCommandNotFound = errors.New("command not found")

const rsyncPath = "/usr/bin/rsync"

func TestFetchRsyncVersion_Success(t *testing.T) {
	executor := &MockCommandExecutor{
		Output: "rsync  version 3.2.3  protocol version 31\n",
		Error:  nil,
	}

	versionInfo, err := internal.FetchRsyncVersion(executor, rsyncPath)

	require.NoError(t, err)
	assert.Equal(t, "rsync  version 3.2.3  protocol version 31\n", versionInfo)
}

func TestFetchRsyncVersion_CommandError(t *testing.T) {
	executor := &MockCommandExecutor{
		Output: "",
		Error:  errCommandNotFound,
	}

	versionInfo, err := internal.FetchRsyncVersion(executor, rsyncPath)

	require.Error(t, err)
	assert.Empty(t, versionInfo)
}

func TestFetchRsyncVersion_InvalidOutput(t *testing.T) {
	executor := &MockCommandExecutor{
		Output: "invalid output",
		Error:  nil,
	}

	versionInfo, err := internal.FetchRsyncVersion(executor, rsyncPath)

	require.Error(t, err)
	assert.Empty(t, versionInfo)
}

func TestFetchRsyncVersion_EmptyPath(t *testing.T) {
	executor := &MockCommandExecutor{
		Output: "",
		Error:  nil,
	}

	versionInfo, err := internal.FetchRsyncVersion(executor, "")

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"\"")
	assert.Empty(t, versionInfo)
}

func TestFetchRsyncVersion_IncompletePath(t *testing.T) {
	executor := &MockCommandExecutor{
		Output: "",
		Error:  nil,
	}

	versionInfo, err := internal.FetchRsyncVersion(executor, "bin/rsync")

	require.Error(t, err)
	require.EqualError(t, err, "rsync path must be an absolute path: \"bin/rsync\"")
	assert.Empty(t, versionInfo)
}
