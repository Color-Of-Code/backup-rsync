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
			BaseLogPath: logPath + "-sim",
			Shell:       &OsExec{},
		},
	}
}

func (c SimulateCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", c.BaseLogPath, job.Name)
	args := ArgumentsForJob(job, logPath, true)

	return c.RunWithArgs(job, args)
}
