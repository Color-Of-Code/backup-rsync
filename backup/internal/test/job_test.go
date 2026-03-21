package internal_test

import (
	. "backup-rsync/backup/internal"
	"backup-rsync/backup/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestApply(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		mockReturn JobStatus
		wantStatus JobStatus
		expectRun  bool
	}{
		{
			name:       "DisabledJob_ReturnsSkipped",
			enabled:    false,
			wantStatus: Skipped,
		},
		{
			name:       "JobFailing_ReturnsFailure",
			enabled:    true,
			mockReturn: Failure,
			wantStatus: Failure,
			expectRun:  true,
		},
		{
			name:       "JobSucceeds_ReturnsSuccess",
			enabled:    true,
			mockReturn: Success,
			wantStatus: Success,
			expectRun:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockJobCommand := NewMockJobCommand(t)
			job := testutil.NewTestJob(testutil.WithEnabled(test.enabled))

			if test.expectRun {
				mockJobCommand.EXPECT().Run(job).Return(test.mockReturn).Once()
			}

			status := job.Apply(mockJobCommand)

			assert.Equal(t, test.wantStatus, status)
		})
	}
}

func TestUnmarshalYAML_InvalidNode(t *testing.T) {
	// A scalar node cannot be decoded into the JobYAML struct
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "not a mapping",
	}

	var job Job

	err := node.Decode(&job)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode YAML node")
}
