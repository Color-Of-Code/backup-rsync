package testutil

import (
	"fmt"
	"strings"
)

// JobOpt configures optional fields on a job definition.
type JobOpt func(*jobDef)

type jobDef struct {
	name       string
	source     string
	target     string
	delete     *bool
	enabled    *bool
	exclusions []string
}

// ConfigBuilder constructs YAML config strings declaratively.
type ConfigBuilder struct {
	sources   []string
	targets   []string
	variables map[string]string
	jobs      []jobDef
}

// NewConfigBuilder creates an empty ConfigBuilder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		variables: make(map[string]string),
	}
}

// Source adds a source path.
func (b *ConfigBuilder) Source(path string) *ConfigBuilder {
	b.sources = append(b.sources, path)

	return b
}

// Target adds a target path.
func (b *ConfigBuilder) Target(path string) *ConfigBuilder {
	b.targets = append(b.targets, path)

	return b
}

// Variable adds a variable for substitution.
func (b *ConfigBuilder) Variable(key, value string) *ConfigBuilder {
	b.variables[key] = value

	return b
}

// AddJob adds a job with the given name, source, and target paths.
func (b *ConfigBuilder) AddJob(name, source, target string, opts ...JobOpt) *ConfigBuilder {
	j := jobDef{name: name, source: source, target: target}
	for _, opt := range opts {
		opt(&j)
	}

	b.jobs = append(b.jobs, j)

	return b
}

// Build produces the YAML config string.
func (b *ConfigBuilder) Build() string {
	var result strings.Builder

	result.WriteString("sources:\n")

	for _, s := range b.sources {
		fmt.Fprintf(&result, "  - path: %q\n", s)
	}

	result.WriteString("targets:\n")

	for _, t := range b.targets {
		fmt.Fprintf(&result, "  - path: %q\n", t)
	}

	if len(b.variables) > 0 {
		result.WriteString("variables:\n")

		for k, v := range b.variables {
			fmt.Fprintf(&result, "  %s: %q\n", k, v)
		}
	}

	result.WriteString("jobs:\n")

	for _, job := range b.jobs {
		fmt.Fprintf(&result, "  - name: %q\n", job.name)
		fmt.Fprintf(&result, "    source: %q\n", job.source)
		fmt.Fprintf(&result, "    target: %q\n", job.target)

		if job.delete != nil {
			fmt.Fprintf(&result, "    delete: %v\n", *job.delete)
		}

		if job.enabled != nil {
			fmt.Fprintf(&result, "    enabled: %v\n", *job.enabled)
		}

		if len(job.exclusions) > 0 {
			result.WriteString("    exclusions:\n")

			for _, e := range job.exclusions {
				fmt.Fprintf(&result, "      - %q\n", e)
			}
		}
	}

	return result.String()
}

// Enabled sets the enabled flag on a job.
func Enabled(v bool) JobOpt {
	return func(j *jobDef) { j.enabled = &v }
}

// Delete sets the delete flag on a job.
func Delete(v bool) JobOpt {
	return func(j *jobDef) { j.delete = &v }
}

// Exclusions sets exclusion patterns on a job.
func Exclusions(v ...string) JobOpt {
	return func(j *jobDef) { j.exclusions = v }
}
