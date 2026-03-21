package internal

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Job represents a backup job configuration for a source/target pair.
//
//nolint:recvcheck // UnmarshalYAML requires pointer receiver while Apply uses value receiver
type Job struct {
	Name       string   `yaml:"name"`
	Source     string   `yaml:"source"`
	Target     string   `yaml:"target"`
	Delete     bool     `yaml:"delete"`
	Enabled    bool     `yaml:"enabled"`
	Exclusions []string `yaml:"exclusions,omitempty"`
}

// JobYAML is a helper struct for proper YAML unmarshaling with defaults.
type JobYAML struct {
	Name       string   `yaml:"name"`
	Source     string   `yaml:"source"`
	Target     string   `yaml:"target"`
	Delete     *bool    `yaml:"delete"`
	Enabled    *bool    `yaml:"enabled"`
	Exclusions []string `yaml:"exclusions,omitempty"`
}

func (job Job) Apply(rsync JobCommand) JobStatus {
	if !job.Enabled {
		return Skipped
	}

	return rsync.Run(job)
}

func boolDefault(ptr *bool, defaultVal bool) bool {
	if ptr != nil {
		return *ptr
	}

	return defaultVal
}

// UnmarshalYAML implements custom YAML unmarshaling to handle defaults properly.
func (job *Job) UnmarshalYAML(node *yaml.Node) error {
	var jobYAML JobYAML

	err := node.Decode(&jobYAML)
	if err != nil {
		return fmt.Errorf("failed to decode YAML node: %w", err)
	}

	job.Name = jobYAML.Name
	job.Source = jobYAML.Source
	job.Target = jobYAML.Target
	job.Exclusions = jobYAML.Exclusions
	job.Delete = boolDefault(jobYAML.Delete, true)
	job.Enabled = boolDefault(jobYAML.Enabled, true)

	return nil
}
