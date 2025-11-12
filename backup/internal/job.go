package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

var execCommand = exec.Command

func buildRsyncCmd(job Job, simulate bool, logPath string) []string {
	args := []string{"-aiv", "--stats"}
	if job.Delete == nil || *job.Delete {
		args = append(args, "--delete")
	}
	if logPath != "" {
		args = append(args, fmt.Sprintf("--log-file=%s", logPath))
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

func ExecuteJob(job Job, simulate bool, show bool, logPath string) string {
	if job.Enabled != nil && !*job.Enabled {
		return "SKIPPED"
	}

	args := buildRsyncCmd(job, simulate, logPath)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: rsync %s\n", strings.Join(args, " "))

	if show {
		return "SUCCESS"
	}

	cmd := execCommand("rsync", args...)
	out, err := cmd.CombinedOutput()
	fmt.Printf("Output:\n%s\n", string(out))
	if err != nil {
		return "FAILURE"
	}
	return "SUCCESS"
}
