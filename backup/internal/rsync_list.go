package internal

import "io"

type ListCommand struct {
	SharedCommand
}

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
