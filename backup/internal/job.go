package internal

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// JobYAML is a helper struct for proper YAML unmarshaling with defaults.
type JobYAML struct {
	Name       string   `yaml:"name"`
	Source     string   `yaml:"source"`
	Target     string   `yaml:"target"`
	Delete     *bool    `yaml:"delete"`
	Enabled    *bool    `yaml:"enabled"`
	Exclusions []string `yaml:"exclusions,omitempty"`
}

type JobStatus string

const (
	Success JobStatus = "SUCCESS"
	Failure JobStatus = "FAILURE"
	Skipped JobStatus = "SKIPPED"
)

func (job Job) Apply(rsync JobCommand) JobStatus {
	if !job.Enabled {
		return Skipped
	}

	return rsync.Run(job)
}

// UnmarshalYAML implements custom YAML unmarshaling to handle defaults properly.
func (job *Job) UnmarshalYAML(node *yaml.Node) error {
	var jobYAML JobYAML

	err := node.Decode(&jobYAML)
	if err != nil {
		return fmt.Errorf("failed to decode YAML node: %w", err)
	}

	// Copy basic fields
	job.Name = jobYAML.Name
	job.Source = jobYAML.Source
	job.Target = jobYAML.Target
	job.Exclusions = jobYAML.Exclusions

	// Handle boolean fields with defaults
	if jobYAML.Delete != nil {
		job.Delete = *jobYAML.Delete
	} else {
		job.Delete = true // default value
	}

	if jobYAML.Enabled != nil {
		job.Enabled = *jobYAML.Enabled
	} else {
		job.Enabled = true // default value
	}

	return nil
}
