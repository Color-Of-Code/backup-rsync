package internal

import (
	"fmt"
)

type SyncCommand struct {
	SharedCommand
}

func NewSyncCommand(binPath string, logPath string) SyncCommand {
	return SyncCommand{
		SharedCommand: SharedCommand{
			BinPath:     binPath,
			BaseLogPath: logPath,
			Shell:       &OsExec{},
		},
	}
}

func (c SyncCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", c.BaseLogPath, job.Name)
	args := ArgumentsForJob(job, logPath, false)

	return c.RunWithArgs(job, args)
}
