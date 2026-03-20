package internal

import "io"

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

func (c ListCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	c.PrintArgs(job, args)

	return Success
}
