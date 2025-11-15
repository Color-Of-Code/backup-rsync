package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

var ExecCommand = exec.Command

func BuildRsyncCmd(job Job, simulate bool, logPath string) []string {
	args := []string{"-aiv", "--stats"}
	if job.Delete {
		args = append(args, "--delete")
	}

	if logPath != "" {
		args = append(args, "--log-file="+logPath)
	}

	for _, excl := range job.Exclusions {
		args = append(args, "--exclude="+excl)
	}

	args = append(args, job.Source, job.Target)
	if simulate {
		args = append([]string{"--dry-run"}, args...)
	}

	return args
}

func ExecuteJob(job Job, simulate bool, show bool, logPath string) string {
	if !job.Enabled {
		return "SKIPPED"
	}

	args := BuildRsyncCmd(job, simulate, logPath)
	fmt.Printf("Job: %s\n", job.Name)
	fmt.Printf("Command: rsync %s\n", strings.Join(args, " "))

	if show {
		return "SUCCESS"
	}

	cmd := ExecCommand("rsync", args...)
	out, err := cmd.CombinedOutput()
	fmt.Printf("Output:\n%s\n", string(out))

	if err != nil {
		return "FAILURE"
	}

	return "SUCCESS"
}
