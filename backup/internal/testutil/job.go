package testutil

import (
	"backup-rsync/backup/internal"
	"slices"
)

// TestJobOpt configures a test Job struct.
type TestJobOpt func(*internal.Job)

// NewTestJob creates a Job with sensible defaults for testing.
// Override individual fields with option functions.
func NewTestJob(opts ...TestJobOpt) internal.Job {
	job := internal.Job{
		Name:       "test-job",
		Source:     "/home/user/docs/",
		Target:     "/backup/user/docs/",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{"*.tmp"},
	}

	for opt := range slices.Values(opts) {
		opt(&job)
	}

	return job
}

// WithEnabled sets the job enabled flag.
func WithEnabled(enabled bool) TestJobOpt {
	return func(job *internal.Job) { job.Enabled = enabled }
}
