package internal

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
	Delete     *bool    `yaml:"delete"`
	Exclusions []string `yaml:"exclusions"`
	Enabled    *bool    `yaml:"enabled"`
}
