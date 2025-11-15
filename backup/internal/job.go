package internal

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandExecutor interface for executing commands.
type CommandExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using actual os/exec.
type RealCommandExecutor struct{}

// Execute runs the actual command.
func (r *RealCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command '%s %s': %w", name, strings.Join(args, " "), err)
	}

	return output, nil
}

func BuildRsyncCmd(job Job, simulate bool, logPath string) []string {
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

func ExecuteJob(job Job, simulate bool, show bool, logPath string) string {
	var osExec CommandExecutor = &RealCommandExecutor{}

	return ExecuteJobWithExecutor(job, simulate, show, logPath, osExec)
}

func ExecuteJobWithExecutor(job Job, simulate bool, show bool, logPath string, executor CommandExecutor) string {
	if !job.Enabled {
		return "SKIPPED"
	}

	args := BuildRsyncCmd(job, simulate, logPath)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: rsync %s\n", strings.Join(args, " "))

	if show {
		return "SUCCESS"
	}

	out, err := executor.Execute("rsync", args...)
	fmt.Printf("Output:\n%s\n", string(out))

	if err != nil {
		return "FAILURE"
	}

	return "SUCCESS"
}
