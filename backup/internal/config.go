package internal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Static errors for wrapping..
var (
	ErrJobValidation   = errors.New("job validation failed")
	ErrInvalidPath     = errors.New("invalid path")
	ErrPathValidation  = errors.New("path validation failed")
	ErrOverlappingPath = errors.New("overlapping path detected")
	ErrJobFailure      = errors.New("one or more jobs failed")
)

// Config represents the overall backup configuration.
type Config struct {
	Sources   []Path            `yaml:"sources"`
	Targets   []Path            `yaml:"targets"`
	Variables map[string]string `yaml:"variables"`
	Jobs      []Job             `yaml:"jobs"`
}

func (cfg Config) String() string {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("error marshaling config: %v", err)
	}

	return string(out)
}

func (cfg Config) Apply(rsync JobCommand, logger *log.Logger, output io.Writer) error {
	versionInfo, fullpath, err := rsync.GetVersionInfo()
	if err != nil {
		logger.Printf("Failed to fetch rsync version: %v", err)
	} else {
		logger.Printf("Rsync Binary Path: %s", fullpath)
		logger.Printf("Rsync Version Info: %s", versionInfo)
	}

	counts := make(map[JobStatus]int)

	for _, job := range cfg.Jobs {
		status := job.Apply(rsync)
		logger.Printf("STATUS [%s]: %s", job.Name, status)
		fmt.Fprintf(output, "Status [%s]: %s\n", job.Name, status)
		counts[status]++
	}

	summary := fmt.Sprintf("Summary: %d succeeded, %d failed, %d skipped",
		counts[Success], counts[Failure], counts[Skipped])
	logger.Print(summary)
	fmt.Fprintln(output, summary)

	if counts[Failure] > 0 {
		return fmt.Errorf("%w: %d of %d jobs", ErrJobFailure, counts[Failure], len(cfg.Jobs))
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

func ResolveConfig(cfg Config) Config {
	resolvedCfg := cfg
	for i, job := range resolvedCfg.Jobs {
		resolvedCfg.Jobs[i].Source = SubstituteVariables(job.Source, cfg.Variables)
		resolvedCfg.Jobs[i].Target = SubstituteVariables(job.Target, cfg.Variables)
	}

	return resolvedCfg
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

func ValidatePath(jobPath string, paths []Path, pathType string, jobName string) error {
	if slices.ContainsFunc(paths, func(p Path) bool { return strings.HasPrefix(jobPath, p.Path) }) {
		return nil
	}

	return fmt.Errorf("%w for job '%s': %s %s", ErrInvalidPath, jobName, pathType, jobPath)
}

func ValidatePaths(cfg Config) error {
	errs := make([]error, 0, len(cfg.Jobs)*2) //nolint:mnd // 2 validations per job: source + target

	for _, job := range cfg.Jobs {
		errs = append(errs, ValidatePath(job.Source, cfg.Sources, "source", job.Name))
		errs = append(errs, ValidatePath(job.Target, cfg.Targets, "target", job.Name))
	}

	joined := errors.Join(errs...)
	if joined != nil {
		return fmt.Errorf("%w: %w", ErrPathValidation, joined)
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

func LoadResolvedConfig(configPath string) (Config, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config: %w", err)
	}
	defer configFile.Close()

	cfg, err := LoadConfig(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	err = ValidateJobNames(cfg.Jobs)
	if err != nil {
		return Config{}, fmt.Errorf("job validation failed: %w", err)
	}

	resolvedCfg := ResolveConfig(cfg)

	err = ValidatePaths(resolvedCfg)
	if err != nil {
		return Config{}, fmt.Errorf("path validation failed: %w", err)
	}

	err = validateJobPaths(resolvedCfg.Jobs, "source", func(job Job) string { return job.Source })
	if err != nil {
		return Config{}, fmt.Errorf("job source path validation failed: %w", err)
	}

	err = validateJobPaths(resolvedCfg.Jobs, "target", func(job Job) string { return job.Target })
	if err != nil {
		return Config{}, fmt.Errorf("job target path validation failed: %w", err)
	}

	return resolvedCfg, nil
}
