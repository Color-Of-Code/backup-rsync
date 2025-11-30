package internal

import (
	"fmt"
	"strings"
)

func (job Job) Apply(rsync RSyncCommand, logPath string) string {
	if !job.Enabled {
		return "SKIPPED"
	}

	args := rsync.ArgumentsForJob(job, logPath)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: rsync %s %s\n", rsync.BinPath, strings.Join(args, " "))

	if rsync.ListOnly {
		return "SUCCESS"
	}

	out, err := rsync.Run(args...)
	fmt.Printf("Output:\n%s\n", string(out))

	if err != nil {
		return "FAILURE"
	}

	return "SUCCESS"
}
