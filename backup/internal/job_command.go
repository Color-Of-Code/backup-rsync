package internal

type JobStatus string

const (
	Success JobStatus = "SUCCESS"
	Failure JobStatus = "FAILURE"
	Skipped JobStatus = "SKIPPED"
)

type JobCommand interface {
	Run(job Job) JobStatus
	GetVersionInfo() (string, string, error)
}
