package internal

type ListCommand struct {
	SharedCommand
}

func NewListCommand(binPath string, shell Exec) ListCommand {
	return ListCommand{
		SharedCommand: NewSharedCommand(binPath, "", shell),
	}
}

func (c ListCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	c.PrintArgs(job, args)

	return Success
}
