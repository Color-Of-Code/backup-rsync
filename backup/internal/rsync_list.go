package internal

import (
	"io"
	"log"
)

// ListCommand prints the rsync commands that would be executed without running them.
type ListCommand struct {
	SharedCommand
}

// NewListCommand creates a ListCommand with the given dependencies.
func NewListCommand(binPath string, shell Exec, output io.Writer) ListCommand {
	return ListCommand{
		SharedCommand: NewSharedCommand(binPath, "", shell, output),
	}
}

func (ListCommand) ReportJobStatus(_ string, _ JobStatus, _ *log.Logger) {}

func (ListCommand) ReportSummary(_ map[JobStatus]int, _ *log.Logger) {}

func (c ListCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	c.PrintArgs(job, args)

	return Success
}
