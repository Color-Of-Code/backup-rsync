package internal

type ListCommand struct {
	SharedCommand
}

func NewListCommand(binPath string) ListCommand {
	return ListCommand{
		SharedCommand: NewSharedCommand(binPath, "", &OsExec{}),
	}
}

func (c ListCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	c.PrintArgs(job, args)

	return Success
}
