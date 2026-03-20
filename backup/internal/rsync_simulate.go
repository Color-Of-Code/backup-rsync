package internal

import "io"

type SimulateCommand struct {
	SharedCommand
}

func NewSimulateCommand(binPath string, logPath string, shell Exec, output io.Writer) SimulateCommand {
	return SimulateCommand{
		SharedCommand: NewSharedCommand(binPath, logPath, shell, output),
	}
}

func (c SimulateCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	// Don't use --log-file in simulate mode as rsync doesn't log file changes to it in dry-run
	args := ArgumentsForJob(job, "", true)

	return c.RunWithArgsAndCaptureOutput(job, args, logPath)
}
