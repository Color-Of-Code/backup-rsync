package internal

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var ErrInvalidRsyncVersion = errors.New("invalid rsync version output")
var ErrInvalidRsyncPath = errors.New("rsync path must be an absolute path")

type RSyncCommand struct {
	BinPath  string
	Simulate bool
	ListOnly bool
	Executor JobRunner
}

func NewRSyncCommand(binPath string) RSyncCommand {
	return RSyncCommand{
		BinPath:  binPath,
		Executor: &RealSync{},
	}
}

func (command RSyncCommand) GetVersionInfo() (string, error) {
	rsyncPath := command.BinPath

	if !filepath.IsAbs(rsyncPath) {
		return "", fmt.Errorf("%w: \"%s\"", ErrInvalidRsyncPath, rsyncPath)
	}

	output, err := command.Run("--version")
	if err != nil {
		return "", fmt.Errorf("error fetching rsync version: %w", err)
	}

	// Validate output
	if !strings.Contains(string(output), "rsync") || !strings.Contains(string(output), "protocol version") {
		return "", fmt.Errorf("%w: %s", ErrInvalidRsyncVersion, output)
	}

	return string(output), nil
}

func (command RSyncCommand) ArgumentsForJob(job Job, logPath string) []string {
	args := []string{"-aiv", "--stats"}
	if job.Delete {
		args = append(args, "--delete")
	}

	if logPath != "" {
		args = append(args, "--log-file="+logPath)
	}

	for _, excl := range job.Exclusions {
		args = append(args, "--exclude="+excl)
	}

	args = append(args, job.Source, job.Target)
	if command.Simulate {
		args = append([]string{"--dry-run"}, args...)
	}

	return args
}

func (command RSyncCommand) Run(args ...string) ([]byte, error) {
	return command.Executor.Execute(command.BinPath, args...)
}
