package internal

type SimulateCommand struct {
	SharedCommand
}

func NewSimulateCommand(binPath string, logPath string) SimulateCommand {
	return SimulateCommand{
		SharedCommand: NewSharedCommand(binPath, logPath, &OsExec{}),
	}
}

func (c SimulateCommand) Run(job Job) JobStatus {
	logPath := c.JobLogPath(job)
	// Don't use --log-file in simulate mode as rsync doesn't log file changes to it in dry-run
	args := ArgumentsForJob(job, "", true)

	return c.RunWithArgsAndCaptureOutput(job, args, logPath)
}
