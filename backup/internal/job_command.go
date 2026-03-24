package internal

import "log/slog"

// JobStatus represents the outcome of a job execution.
type JobStatus string

const (
	// Success indicates the job completed successfully.
	Success JobStatus = "SUCCESS"
	// Failure indicates the job failed.
	Failure JobStatus = "FAILURE"
	// Skipped indicates the job was skipped (e.g., disabled).
	Skipped JobStatus = "SKIPPED"
)

// JobCommand defines the interface for running backup jobs.
type JobCommand interface {
	Run(job Job) JobStatus
	GetVersionInfo() (string, string, error)
	ReportJobStatus(jobName string, status JobStatus, logger *slog.Logger)
	ReportSummary(counts map[JobStatus]int, logger *slog.Logger)
}
