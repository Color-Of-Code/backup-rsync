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

type includeDef struct {
	uses string
	with map[string]string
}

// ConfigBuilder constructs YAML config strings declaratively.
type ConfigBuilder struct {
	sources      []string
	targets      []string
	variables    map[string]string
	jobs         []jobDef
	templateVars []string
	includes     []includeDef
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

// TemplateVar adds a required template variable name.
func (b *ConfigBuilder) TemplateVar(name string) *ConfigBuilder {
	b.templateVars = append(b.templateVars, name)

	return b
}

// AddInclude adds a template include with variable bindings.
func (b *ConfigBuilder) AddInclude(uses string, with map[string]string) *ConfigBuilder {
	b.includes = append(b.includes, includeDef{uses: uses, with: with})

	return b
}

// Build produces the YAML config string.
func (b *ConfigBuilder) Build() string {
	var result strings.Builder

	b.writeTemplate(&result)
	b.writeIncludes(&result)

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
		writeJob(&result, job)
	}

	return result.String()
}

func (b *ConfigBuilder) writeTemplate(writer *strings.Builder) {
	if len(b.templateVars) == 0 {
		return
	}

	writer.WriteString("template:\n  variables:\n")

	for _, v := range b.templateVars {
		fmt.Fprintf(writer, "    - %q\n", v)
	}
}

func (b *ConfigBuilder) writeIncludes(writer *strings.Builder) {
	if len(b.includes) == 0 {
		return
	}

	writer.WriteString("include:\n")

	for _, inc := range b.includes {
		fmt.Fprintf(writer, "  - uses: %q\n", inc.uses)

		if len(inc.with) > 0 {
			writer.WriteString("    with:\n")

			for k, v := range inc.with {
				fmt.Fprintf(writer, "      %s: %q\n", k, v)
			}
		}
	}
}

func writeJob(writer *strings.Builder, job jobDef) {
	fmt.Fprintf(writer, "  - name: %q\n", job.name)
	fmt.Fprintf(writer, "    source: %q\n", job.source)
	fmt.Fprintf(writer, "    target: %q\n", job.target)

	if job.delete != nil {
		fmt.Fprintf(writer, "    delete: %v\n", *job.delete)
	}

	if job.enabled != nil {
		fmt.Fprintf(writer, "    enabled: %v\n", *job.enabled)
	}

	if len(job.exclusions) > 0 {
		writer.WriteString("    exclusions:\n")

		for _, e := range job.exclusions {
			fmt.Fprintf(writer, "      - %q\n", e)
		}
	}
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
