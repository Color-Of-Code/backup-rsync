package internal

import (
	"gopkg.in/yaml.v3"
)

// Centralized type definitions

type Path struct {
	Path       string   `yaml:"path"`
	Exclusions []string `yaml:"exclusions"`
}

type Config struct {
	Sources   []Path            `yaml:"sources"`
	Targets   []Path            `yaml:"targets"`
	Variables map[string]string `yaml:"variables"`
	Jobs      []Job             `yaml:"jobs"`
}

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

// UnmarshalYAML implements custom YAML unmarshaling to handle defaults properly.
func (j *Job) UnmarshalYAML(node *yaml.Node) error {
	var jobYAML JobYAML
	err := node.Decode(&jobYAML)
	if err != nil {
		return err
	}

	// Copy basic fields
	j.Name = jobYAML.Name
	j.Source = jobYAML.Source
	j.Target = jobYAML.Target
	j.Exclusions = jobYAML.Exclusions

	// Handle boolean fields with defaults
	if jobYAML.Delete != nil {
		j.Delete = *jobYAML.Delete
	} else {
		j.Delete = true // default value
	}

	if jobYAML.Enabled != nil {
		j.Enabled = *jobYAML.Enabled
	} else {
		j.Enabled = true // default value
	}

	return nil
}
