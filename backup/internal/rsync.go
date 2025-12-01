package internal

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var ErrInvalidRsyncVersion = errors.New("invalid rsync version output")
var ErrInvalidRsyncPath = errors.New("rsync path must be an absolute path")

type SyncCommand struct {
	BinPath     string
	BaseLogPath string

	Executor JobRunner
}

func NewSyncCommand(binPath string, logPath string) SyncCommand {
	return SyncCommand{
		BinPath:     binPath,
		BaseLogPath: logPath,
		Executor:    &RealSync{},
	}
}

func (command SyncCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", command.BaseLogPath, job.Name)

	args := ArgumentsForJob(job, logPath, false)

	return command.RunWithArgs(job, args)
}

func (command SyncCommand) PrintArgs(job Job, args []string) {
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: %s %s\n", command.BinPath, strings.Join(args, " "))
}

func (command SyncCommand) RunWithArgs(job Job, args []string) JobStatus {
	command.PrintArgs(job, args)

	out, err := command.Executor.Execute(command.BinPath, args...)
	fmt.Printf("Output:\n%s\n", string(out))

	if err != nil {
		return Failure
	}

	return Success
}

type SimulateCommand struct {
	SyncCommand
}

func NewSimulateCommand(binPath string, logPath string) SimulateCommand {
	return SimulateCommand{
		SyncCommand: SyncCommand{
			BinPath:     binPath,
			BaseLogPath: logPath,
			Executor:    &RealSync{},
		},
	}
}

func (command SimulateCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", command.BaseLogPath, job.Name)
	args := ArgumentsForJob(job, logPath, true)

	return command.RunWithArgs(job, args)
}

type ListCommand struct {
	SyncCommand
}

func NewListCommand(binPath string) ListCommand {
	return ListCommand{
		SyncCommand: SyncCommand{
			BinPath:     binPath,
			BaseLogPath: "",
			Executor:    &RealSync{},
		},
	}
}
func (command ListCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", command.BaseLogPath, job.Name)

	args := ArgumentsForJob(job, logPath, false)
	command.PrintArgs(job, args)

	return Success
}

func (command SyncCommand) GetVersionInfo() (string, string, error) {
	rsyncPath := command.BinPath

	if !filepath.IsAbs(rsyncPath) {
		return "", "", fmt.Errorf("%w: \"%s\"", ErrInvalidRsyncPath, rsyncPath)
	}

	output, err := command.Executor.Execute(command.BinPath, "--version")
	if err != nil {
		return "", "", fmt.Errorf("error fetching rsync version: %w", err)
	}

	// Validate output
	if !strings.Contains(string(output), "rsync") || !strings.Contains(string(output), "protocol version") {
		return "", "", fmt.Errorf("%w: %s", ErrInvalidRsyncVersion, output)
	}

	return string(output), rsyncPath, nil
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
