package internal

import (
	"fmt"
)

type SimulateCommand struct {
	SharedCommand
}

func NewSimulateCommand(binPath string, logPath string) SimulateCommand {
	return SimulateCommand{
		SharedCommand: SharedCommand{
			BinPath:     binPath,
			BaseLogPath: logPath,
			Shell:       &OsExec{},
		},
	}
}

func (c SimulateCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", c.BaseLogPath, job.Name)
	// Don't use --log-file in simulate mode as rsync doesn't log file changes to it in dry-run
	args := ArgumentsForJob(job, "", true)

	return c.RunWithArgsAndCaptureOutput(job, args, logPath)
}
