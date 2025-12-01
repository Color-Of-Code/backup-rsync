package internal_test

import (
	. "backup-rsync/backup/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newJob() Job {
	return Job{
		Name:       "job",
		Source:     "",
		Target:     "",
		Delete:     true,
		Enabled:    true,
		Exclusions: []string{},
	}
}

func TestApply_DisabledJob_ReturnsSkippedAndRunIsNotCalled(t *testing.T) {
	mockJobCommand := NewMockJobCommand(t)

	disabledJob := newJob()
	disabledJob.Enabled = false

	// No expectations set - Run should NOT be called for disabled jobs

	status := disabledJob.Apply(mockJobCommand)

	assert.Equal(t, Skipped, status)
}

func TestApply_JobFailing_RunIsCalledAndReturnsFailure(t *testing.T) {
	mockJobCommand := NewMockJobCommand(t)

	job := newJob()

	// Set expectation that Run will be called and return Failure
	mockJobCommand.EXPECT().Run(job).Return(Failure).Once()

	status := job.Apply(mockJobCommand)

	assert.Equal(t, Failure, status)
}

func TestApply_JobSucceeds_RunIsCalledAndReturnsSuccess(t *testing.T) {
	mockJobCommand := NewMockJobCommand(t)

	job := newJob()

	// Set expectation that Run will be called and return Success
	mockJobCommand.EXPECT().Run(job).Return(Success).Once()

	status := job.Apply(mockJobCommand)

	assert.Equal(t, Success, status)
}
