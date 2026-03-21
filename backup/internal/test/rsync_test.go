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
	tests := []struct {
		name     string
		job      Job
		logPath  string
		simulate bool
		wantArgs []string
	}{
		{
			name: "SimulateWithDelete",
			job: Job{
				Delete: true, Source: "/home/user/Music/", Target: "/target/user/music/home",
				Exclusions: []string{"*.tmp", "node_modules/"},
			},
			simulate: true,
			wantArgs: []string{
				"--dry-run", "-aiv", "--stats", "--delete",
				"--exclude=*.tmp", "--exclude=node_modules/",
				"/home/user/Music/", "/target/user/music/home",
			},
		},
		{
			name: "RealWithLogPath",
			job: Job{
				Delete: false, Source: "/home/user/Documents/", Target: "/backup/user/documents",
				Exclusions: []string{"*.log", "temp/"},
			},
			logPath: "/var/log/rsync.log",
			wantArgs: []string{
				"-aiv", "--stats",
				"--log-file=/var/log/rsync.log",
				"--exclude=*.log", "--exclude=temp/",
				"/home/user/Documents/", "/backup/user/documents",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := ArgumentsForJob(test.job, test.logPath, test.simulate)

			assert.Equal(t, strings.Join(test.wantArgs, " "), strings.Join(args, " "))
		})
	}
}

func TestGetVersionInfo(t *testing.T) {
	tests := []struct {
		name, binPath, wantVersion, wantPath, wantErr string
		mockOutput                                    []byte
		mockErr                                       error
	}{
		{name: "Success", binPath: rsyncPath,
			mockOutput:  []byte("rsync  version 3.2.3  protocol version 31\n"),
			wantVersion: "rsync  version 3.2.3  protocol version 31\n", wantPath: rsyncPath},
		{name: "CommandError", binPath: rsyncPath,
			mockErr: errCommandNotFound, wantErr: "error fetching rsync version"},
		{name: "InvalidOutput", binPath: rsyncPath,
			mockOutput: []byte("invalid output"), wantErr: "invalid rsync version output"},
		{name: "EmptyPath",
			wantErr: `rsync path must be an absolute path: ""`},
		{name: "IncompletePath", binPath: "bin/rsync",
			wantErr: `rsync path must be an absolute path: "bin/rsync"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockExec := NewMockExec(t)
			rsync := SharedCommand{
				BinPath: test.binPath,
				Shell:   mockExec,
				Output:  io.Discard,
			}

			if strings.HasPrefix(test.binPath, "/") {
				mockExec.EXPECT().Execute(rsyncPath, mock.MatchedBy(func(args []string) bool {
					return len(args) == 1 && args[0] == RsyncVersionFlag
				})).Return(test.mockOutput, test.mockErr).Once()
			}

			versionInfo, fullpath, err := rsync.GetVersionInfo()

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				assert.Empty(t, versionInfo)
				assert.Empty(t, fullpath)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantPath, fullpath)
				assert.Equal(t, test.wantVersion, versionInfo)
			}
		})
	}
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
