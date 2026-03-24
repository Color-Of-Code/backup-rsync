package internal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Static errors for wrapping..
var (
	ErrJobValidation       = errors.New("job validation failed")
	ErrOverlappingPath     = errors.New("overlapping path detected")
	ErrJobFailure          = errors.New("one or more jobs failed")
	ErrMissingTemplateVars = errors.New("missing required template variables")
	ErrNestedIncludes      = errors.New("nested includes are not supported")
)

// Template declares required variables for a template config file.
type Template struct {
	Variables []string `yaml:"variables"`
}

// Include references a template config to instantiate with specific variable values.
type Include struct {
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
}

// Mapping defines a source-to-target directory pair with its own list of backup jobs.
// Job paths within a mapping are relative to the mapping's Source and Target.
type Mapping struct {
	Name       string   `yaml:"name"`
	Source     string   `yaml:"source"`
	Target     string   `yaml:"target"`
	Exclusions []string `yaml:"exclusions,omitempty"`
	Jobs       []Job    `yaml:"jobs"`
}

// Config represents the overall backup configuration.
type Config struct {
	Template  *Template         `yaml:"template,omitempty"`
	Include   []Include         `yaml:"include,omitempty"`
	Variables map[string]string `yaml:"variables,omitempty"`
	Mappings  []Mapping         `yaml:"mappings"`
}

// AllJobs returns a flat list of all jobs across all mappings.
func (cfg Config) AllJobs() []Job {
	var jobs []Job
	for _, m := range cfg.Mappings {
		jobs = append(jobs, m.Jobs...)
	}

	return jobs
}

func (cfg Config) String() string {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("error marshaling config: %v", err)
	}

	return string(out)
}

func (cfg Config) Apply(rsync JobCommand, logger *log.Logger) error {
	versionInfo, fullpath, err := rsync.GetVersionInfo()
	if err != nil {
		logger.Printf("Failed to fetch rsync version: %v", err)
	} else {
		logger.Printf("Rsync Binary Path: %s", fullpath)
		logger.Printf("Rsync Version Info: %s", versionInfo)
	}

	counts := make(map[JobStatus]int)
	allJobs := cfg.AllJobs()

	for _, job := range allJobs {
		status := job.Apply(rsync)
		rsync.ReportJobStatus(job.Name, status, logger)
		counts[status]++
	}

	rsync.ReportSummary(counts, logger)

	if counts[Failure] > 0 {
		return fmt.Errorf("%w: %d of %d jobs", ErrJobFailure, counts[Failure], len(allJobs))
	}

	return nil
}

func LoadConfig(reader io.Reader) (Config, error) {
	var cfg Config

	err := yaml.NewDecoder(reader).Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decode YAML: %w", err)
	}

	// Defaults are handled in Job.UnmarshalYAML

	return cfg, nil
}

func SubstituteVariables(input string, variables map[string]string) string {
	oldnew := make([]string, 0, len(variables)*2) //nolint:mnd // 2 entries per variable: key placeholder + value
	for key, value := range variables {
		oldnew = append(oldnew, "${"+key+"}", value)
	}

	return strings.NewReplacer(oldnew...).Replace(input)
}

func resolveField(input string, variables map[string]string) (string, error) {
	result := SubstituteVariables(input, variables)

	resolved, err := ResolveMacros(result)
	if err != nil {
		return "", err
	}

	return resolved, nil
}

const maxResolvePasses = 10

// ResolveVariables resolves variable-to-variable references within the variables map.
// Variables can reference other variables (e.g., source_home: "/home/${user}").
// Performs multiple passes until no further substitutions occur or maxResolvePasses is reached.
func ResolveVariables(variables map[string]string) map[string]string {
	resolved := make(map[string]string, len(variables))
	maps.Copy(resolved, variables)

	for range maxResolvePasses {
		changed := false

		for k, v := range resolved {
			newV := SubstituteVariables(v, resolved)
			if newV != v {
				resolved[k] = newV
				changed = true
			}
		}

		if !changed {
			break
		}
	}

	return resolved
}

// resolveFields resolves variables and macros in all mapping and job fields.
// When joinPaths is true, job paths are joined with their mapping base paths
// and unresolved macros are validated — this is the final resolution step.
// When joinPaths is false, only variable/macro substitution is performed,
// used by expandIncludes so that path joining happens only once.
func resolveFields(cfg Config, joinPaths bool) (Config, error) {
	resolved := cfg
	resolved.Variables = ResolveVariables(cfg.Variables)

	for mIdx := range resolved.Mappings {
		err := resolveMapping(&resolved.Mappings[mIdx], resolved.Variables, joinPaths)
		if err != nil {
			return Config{}, err
		}
	}

	if joinPaths {
		err := ValidateNoUnresolvedMacros(resolved)
		if err != nil {
			return Config{}, fmt.Errorf("macro resolution incomplete: %w", err)
		}
	}

	return resolved, nil
}

func resolveMapping(mapping *Mapping, variables map[string]string, joinPaths bool) error {
	var err error

	mapping.Name, err = resolveField(mapping.Name, variables)
	if err != nil {
		return fmt.Errorf("resolving mapping name %q: %w", mapping.Name, err)
	}

	mapping.Source, err = resolveField(mapping.Source, variables)
	if err != nil {
		return fmt.Errorf("resolving mapping source %q: %w", mapping.Source, err)
	}

	mapping.Target, err = resolveField(mapping.Target, variables)
	if err != nil {
		return fmt.Errorf("resolving mapping target %q: %w", mapping.Target, err)
	}

	for jIdx := range mapping.Jobs {
		job := &mapping.Jobs[jIdx]

		job.Name, err = resolveField(job.Name, variables)
		if err != nil {
			return fmt.Errorf("resolving job name %q: %w", job.Name, err)
		}

		job.Source, err = resolveField(job.Source, variables)
		if err != nil {
			return fmt.Errorf("resolving job source %q: %w", job.Source, err)
		}

		job.Target, err = resolveField(job.Target, variables)
		if err != nil {
			return fmt.Errorf("resolving job target %q: %w", job.Target, err)
		}

		if joinPaths {
			job.Source = filepath.Join(mapping.Source, job.Source) + "/"
			job.Target = filepath.Join(mapping.Target, job.Target)
		}
	}

	return nil
}

// ResolveConfig resolves all variables, macros, and joins job paths with mapping base paths.
func ResolveConfig(cfg Config) (Config, error) {
	return resolveFields(cfg, true)
}

func ValidateJobNames(jobs []Job) error {
	var invalidNames []string

	nameSet := make(map[string]bool)

	for _, job := range jobs {
		if nameSet[job.Name] {
			invalidNames = append(invalidNames, "duplicate job name: "+job.Name)
		} else {
			nameSet[job.Name] = true
		}

		if strings.ContainsFunc(job.Name, func(r rune) bool { return r > 127 || r == ' ' }) {
			invalidNames = append(invalidNames, "invalid characters in job name: "+job.Name)
		}
	}

	if len(invalidNames) > 0 {
		return fmt.Errorf("%w: %v", ErrJobValidation, invalidNames)
	}

	return nil
}

func validateJobPaths(jobs []Job, pathType string, getPath func(job Job) string) error {
	for i, job1 := range jobs {
		for j, job2 := range jobs {
			if i != j {
				path1, path2 := NormalizePath(getPath(job1)), NormalizePath(getPath(job2))

				excluded := pathType == "source" && slices.ContainsFunc(job2.Exclusions, func(exclusion string) bool {
					return strings.HasPrefix(path1, NormalizePath(filepath.Join(job2.Source, exclusion)))
				})

				if !excluded && strings.HasPrefix(path1, path2) {
					return fmt.Errorf("%w: job '%s' has a %s path overlapping with job '%s'",
						ErrOverlappingPath, job1.Name, pathType, job2.Name)
				}
			}
		}
	}

	return nil
}

// ValidateTemplateVars checks that all variables declared in the template section have values.
func ValidateTemplateVars(cfg Config) error {
	if cfg.Template == nil || len(cfg.Template.Variables) == 0 {
		return nil
	}

	var missing []string

	for _, v := range cfg.Template.Variables {
		if _, ok := cfg.Variables[v]; !ok {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %v", ErrMissingTemplateVars, missing)
	}

	return nil
}

func loadTemplateConfig(templatePath string) (Config, error) {
	templateFile, err := os.Open(templatePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open: %w", err)
	}
	defer templateFile.Close()

	cfg, err := LoadConfig(templateFile)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse: %w", err)
	}

	if len(cfg.Include) > 0 {
		return Config{}, ErrNestedIncludes
	}

	return cfg, nil
}

func expandIncludes(cfg *Config, configDir string) error {
	for _, inc := range cfg.Include {
		templatePath := inc.Uses
		if !filepath.IsAbs(templatePath) {
			templatePath = filepath.Join(configDir, templatePath)
		}

		tmplCfg, err := loadTemplateConfig(templatePath)
		if err != nil {
			return fmt.Errorf("include %q: %w", inc.Uses, err)
		}

		if tmplCfg.Variables == nil {
			tmplCfg.Variables = make(map[string]string)
		}

		maps.Copy(tmplCfg.Variables, inc.With)

		err = ValidateTemplateVars(tmplCfg)
		if err != nil {
			return fmt.Errorf("include %q: %w", inc.Uses, err)
		}

		resolved, err := resolveFields(tmplCfg, false)
		if err != nil {
			return fmt.Errorf("include %q: resolving config: %w", inc.Uses, err)
		}

		cfg.Mappings = append(cfg.Mappings, resolved.Mappings...)
	}

	cfg.Include = nil

	return nil
}

func mergeOverrides(cfg Config, overrides []map[string]string) Config {
	for _, override := range overrides {
		if cfg.Variables == nil {
			cfg.Variables = make(map[string]string)
		}

		maps.Copy(cfg.Variables, override)
	}

	return cfg
}

func LoadResolvedConfig(configPath string, overrides ...map[string]string) (Config, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config: %w", err)
	}
	defer configFile.Close()

	cfg, err := LoadConfig(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	cfg = mergeOverrides(cfg, overrides)

	return resolveAndValidate(cfg, filepath.Dir(configPath))
}

func resolveAndValidate(cfg Config, configDir string) (Config, error) {
	err := expandIncludes(&cfg, configDir)
	if err != nil {
		return Config{}, fmt.Errorf("expanding includes: %w", err)
	}

	err = ValidateTemplateVars(cfg)
	if err != nil {
		return Config{}, fmt.Errorf("template validation failed: %w", err)
	}

	resolvedCfg, err := ResolveConfig(cfg)
	if err != nil {
		return Config{}, fmt.Errorf("config resolution failed: %w", err)
	}

	allJobs := resolvedCfg.AllJobs()

	err = ValidateJobNames(allJobs)
	if err != nil {
		return Config{}, fmt.Errorf("job validation failed: %w", err)
	}

	err = validateJobPaths(allJobs, "source", func(job Job) string { return job.Source })
	if err != nil {
		return Config{}, fmt.Errorf("job source path validation failed: %w", err)
	}

	err = validateJobPaths(allJobs, "target", func(job Job) string { return job.Target })
	if err != nil {
		return Config{}, fmt.Errorf("job target path validation failed: %w", err)
	}

	return resolvedCfg, nil
}
