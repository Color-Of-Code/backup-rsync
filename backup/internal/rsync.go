package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidRsyncVersion = errors.New("invalid rsync version output")
var ErrInvalidRsyncPath = errors.New("rsync path must be an absolute path")

const RsyncVersionFlag = "--version"

type SharedCommand struct {
	BinPath     string
	BaseLogPath string

	Shell  Exec
	Output io.Writer
}

func NewSharedCommand(binPath string, logPath string, shell Exec, output io.Writer) SharedCommand {
	return SharedCommand{
		BinPath:     binPath,
		BaseLogPath: logPath,
		Shell:       shell,
		Output:      output,
	}
}

func (c SharedCommand) JobLogPath(job Job) string {
	return fmt.Sprintf("%s/job-%s.log", c.BaseLogPath, job.Name)
}

func (c SharedCommand) PrintArgs(job Job, args []string) {
	fmt.Fprintf(c.Output, "Job: %s\n", job.Name)
	fmt.Fprintf(c.Output, "Command: %s %s\n", c.BinPath, strings.Join(args, " "))
}

func (c SharedCommand) RunWithArgs(job Job, args []string) JobStatus {
	c.PrintArgs(job, args)

	out, err := c.Shell.Execute(c.BinPath, args...)
	fmt.Fprintf(c.Output, "Output:\n%s\n", string(out))

	if err != nil {
		return Failure
	}

	return Success
}

func (c SharedCommand) RunWithArgsAndCaptureOutput(job Job, args []string, logPath string) JobStatus {
	c.PrintArgs(job, args)

	out, err := c.Shell.Execute(c.BinPath, args...)

	// Write output to log file for simulate commands
	if logPath != "" {
		writeErr := os.WriteFile(logPath, out, LogFilePermission)
		if writeErr != nil {
			fmt.Fprintf(c.Output, "Warning: Failed to write output to log file %s: %v\n", logPath, writeErr)
		}
	}

	if err != nil {
		return Failure
	}

	return Success
}

func (c SharedCommand) GetVersionInfo() (string, string, error) {
	rsyncPath := c.BinPath

	if !filepath.IsAbs(rsyncPath) {
		return "", "", fmt.Errorf("%w: \"%s\"", ErrInvalidRsyncPath, rsyncPath)
	}

	output, err := c.Shell.Execute(c.BinPath, RsyncVersionFlag)
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
