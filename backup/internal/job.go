package internal

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func buildRsyncCmd(job Job, simulate bool) *exec.Cmd {
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
	return exec.Command("rsync", args...)
}

func ExecuteJob(job Job, simulate bool, logger *log.Logger) string {
	if job.Enabled != nil && !*job.Enabled {
		logger.Printf("SKIPPED [%s]: Job is disabled", job.Name)
		return "SKIPPED"
	}

	cmd := buildRsyncCmd(job, simulate)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: %s\n", strings.Join(cmd.Args, " "))
	if !simulate {
		out, err := cmd.CombinedOutput()
		if err != nil {
			logger.Printf("ERROR [%s]: %v\nOutput: %s", job.Name, err, string(out))
			return "FAILURE"
		}
		logger.Printf("SUCCESS [%s]: %s", job.Name, string(out))
	}
	return "SUCCESS"
}
