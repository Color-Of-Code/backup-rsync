package internal

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var execCommand = exec.Command

func buildRsyncCmd(job Job, simulate bool) []string {
	args := []string{"-aiv", "--info=progress2"}
	if job.Delete == nil || *job.Delete {
		args = append(args, "--delete")
	}
	for _, excl := range job.Exclusions {
		args = append(args, fmt.Sprintf("--exclude=%s", excl))
	}
	args = append(args, job.Source, job.Target)
	if simulate {
		args = append([]string{"--dry-run"}, args...)
	}
	return args
}

func ExecuteJob(job Job, simulate bool, show bool, logger *log.Logger) string {
	if job.Enabled != nil && !*job.Enabled {
		if logger != nil {
			logger.Printf("SKIPPED [%s]: Job is disabled", job.Name)
		}
		return "SKIPPED"
	}

	args := buildRsyncCmd(job, simulate)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: rsync %s\n", strings.Join(args, " "))

	if show {
		return "SUCCESS"
	}

	cmd := execCommand("rsync", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if logger != nil {
			logger.Printf("ERROR [%s]: %v\nOutput: %s", job.Name, err, string(out))
		}
		return "FAILURE"
	}
	if logger != nil {
		logger.Printf("SUCCESS [%s]: %s", job.Name, string(out))
	}
	return "SUCCESS"
}
