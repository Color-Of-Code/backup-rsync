package internal

import "io"

// SyncCommand runs rsync to perform the actual backup.
type SyncCommand struct {
	SharedCommand
}

// NewSyncCommand creates a SyncCommand with the given dependencies.
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
