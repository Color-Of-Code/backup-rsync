package internal

type SyncCommand struct {
	SharedCommand
}

func NewSyncCommand(binPath string, logPath string, shell Exec) SyncCommand {
	return SyncCommand{
		SharedCommand: NewSharedCommand(binPath, logPath, shell),
	}
}

func (c SyncCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	return c.RunWithArgs(job, args)
}
