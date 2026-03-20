package internal

import "io"

type SyncCommand struct {
	SharedCommand
}

func NewSyncCommand(binPath string, logPath string, shell Exec, output io.Writer) SyncCommand {
	return SyncCommand{
		SharedCommand: NewSharedCommand(binPath, logPath, shell, output),
	}
}

func (c SyncCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	args := ArgumentsForJob(job, logPath, false)

	return c.RunWithArgs(job, args)
}
