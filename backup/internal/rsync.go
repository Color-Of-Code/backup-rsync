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

func NewRSyncSimulateCommand(binPath string) RSyncCommand {
	return RSyncCommand{
		BinPath:  binPath,
		Simulate: true,
		Executor: &RealSync{},
	}
}

func NewListCommand(binPath string) RSyncCommand {
	return RSyncCommand{
		BinPath:  binPath,
		ListOnly: true,
		Executor: &RealSync{},
	}
}

func (command RSyncCommand) GetVersionInfo() (string, error) {
	rsyncPath := command.BinPath

	if !filepath.IsAbs(rsyncPath) {
		return "", fmt.Errorf("%w: \"%s\"", ErrInvalidRsyncPath, rsyncPath)
	}

	output, err := command.Executor.Execute("--version")
	if err != nil {
		return "", fmt.Errorf("error fetching rsync version: %w", err)
	}

	// Validate output
	if !strings.Contains(string(output), "rsync") || !strings.Contains(string(output), "protocol version") {
		return "", fmt.Errorf("%w: %s", ErrInvalidRsyncVersion, output)
	}

	return string(output), nil
}

func ArgumentsForJob(job Job, logPath string, simulate bool) []string {
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
	if simulate {
		args = append([]string{"--dry-run"}, args...)
	}

	return args
}

func (rsync RSyncCommand) Run(job Job, logPath string) string {
	args := ArgumentsForJob(job, logPath, rsync.Simulate)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: rsync %s %s\n", rsync.BinPath, strings.Join(args, " "))

	if rsync.ListOnly {
		return "SUCCESS"
	}

	out, err := rsync.Executor.Execute(rsync.BinPath, args...)
	fmt.Printf("Output:\n%s\n", string(out))

	if err != nil {
		return "FAILURE"
	}

	return "SUCCESS"
}
