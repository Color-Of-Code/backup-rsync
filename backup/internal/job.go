package internal

func (job Job) Apply(rsync RSyncCommand, logPath string) string {
	if !job.Enabled {
		return "SKIPPED"
	}

	return rsync.Run(job, logPath)
}
