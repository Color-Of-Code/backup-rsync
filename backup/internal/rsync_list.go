package internal

import (
	"fmt"
)

type ListCommand struct {
	SharedCommand
}

func NewListCommand(binPath string) ListCommand {
	return ListCommand{
		SharedCommand: SharedCommand{
			BinPath:     binPath,
			BaseLogPath: "",
			Shell:       &OsExec{},
		},
	}
}
func (c ListCommand) Run(job Job) JobStatus {
	logPath := fmt.Sprintf("%s/job-%s.log", c.BaseLogPath, job.Name)
	args := ArgumentsForJob(job, logPath, false)

	c.PrintArgs(job, args)

	return Success
}
