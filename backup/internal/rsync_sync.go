package internal

type SyncCommand struct {
	SharedCommand
}

func NewSyncCommand(binPath string, logPath string) SyncCommand {
	return SyncCommand{
		SharedCommand: NewSharedCommand(binPath, logPath, &OsExec{}),
	}
}

func (c SyncCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	return c.RunWithArgs(job, args)
}
