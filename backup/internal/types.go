package internal

// Path represents a source or target path with optional exclusions.
type Path struct {
	Path       string   `yaml:"path"`
	Exclusions []string `yaml:"exclusions"`
}

// Config represents the overall backup configuration.
type Config struct {
	Sources   []Path            `yaml:"sources"`
	Targets   []Path            `yaml:"targets"`
	Variables map[string]string `yaml:"variables"`
	Jobs      []Job             `yaml:"jobs"`
}

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

type JobCommand interface {
	Run(job Job) JobStatus
	GetVersionInfo() (string, string, error)
}
