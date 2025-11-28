package internal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Static errors for wrapping..
var (
	ErrJobValidation   = errors.New("job validation failed")
	ErrInvalidPath     = errors.New("invalid path")
	ErrPathValidation  = errors.New("path validation failed")
	ErrOverlappingPath = errors.New("overlapping path detected")
)

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
	for key, value := range variables {
		placeholder := fmt.Sprintf("${%s}", key)
		input = strings.ReplaceAll(input, placeholder, value)
	}

	return input
}

func resolveConfig(cfg Config) Config {
	resolvedCfg := cfg
	for i, job := range resolvedCfg.Jobs {
		resolvedCfg.Jobs[i].Source = SubstituteVariables(job.Source, cfg.Variables)
		resolvedCfg.Jobs[i].Target = SubstituteVariables(job.Target, cfg.Variables)
	}

	return resolvedCfg
}

func ValidateJobNames(jobs []Job) error {
	invalidNames := []string{}
	nameSet := make(map[string]bool)

	for _, job := range jobs {
		if nameSet[job.Name] {
			invalidNames = append(invalidNames, "duplicate job name: "+job.Name)
		} else {
			nameSet[job.Name] = true
		}

		for _, r := range job.Name {
			if r > 127 || r == ' ' {
				invalidNames = append(invalidNames, "invalid characters in job name: "+job.Name)

				break
			}
		}
	}

	if len(invalidNames) > 0 {
		return fmt.Errorf("%w: %v", ErrJobValidation, invalidNames)
	}

	return nil
}

func ValidatePath(jobPath string, paths []Path, pathType string, jobName string) error {
	for _, path := range paths {
		if strings.HasPrefix(jobPath, path.Path) {
			return nil
		}
	}

	return fmt.Errorf("%w for job '%s': %s %s", ErrInvalidPath, jobName, pathType, jobPath)
}

func ValidatePaths(cfg Config) error {
	invalidPaths := []string{}

	for _, job := range cfg.Jobs {
		err := ValidatePath(job.Source, cfg.Sources, "source", job.Name)
		if err != nil {
			invalidPaths = append(invalidPaths, err.Error())
		}

		err = ValidatePath(job.Target, cfg.Targets, "target", job.Name)
		if err != nil {
			invalidPaths = append(invalidPaths, err.Error())
		}
	}

	if len(invalidPaths) > 0 {
		return fmt.Errorf("%w: %v", ErrPathValidation, invalidPaths)
	}

	return nil
}

func validateJobPaths(jobs []Job, pathType string, getPath func(job Job) string) error {
	for i, job1 := range jobs {
		for j, job2 := range jobs {
			if i != j {
				path1, path2 := NormalizePath(getPath(job1)), NormalizePath(getPath(job2))

				// Check if path2 is part of job1's exclusions
				excluded := false

				if pathType == "source" {
					for _, exclusion := range job2.Exclusions {
						exclusionPath := NormalizePath(filepath.Join(job2.Source, exclusion))
						// log.Printf("job2: %s %s\n", job2.Name, exclusionPath)
						if strings.HasPrefix(path1, exclusionPath) {
							excluded = true

							break
						}
					}
				}

				if !excluded && strings.HasPrefix(path1, path2) {
					return fmt.Errorf("%w: job '%s' has a %s path overlapping with job '%s'",
						ErrOverlappingPath, job1.Name, pathType, job2.Name)
				}
			}
		}
	}

	return nil
}

func LoadResolvedConfig(configPath string) Config {
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config: %v", err)
	}

	cfg, err := LoadConfig(configFile)
	_ = configFile.Close()

	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	err = ValidateJobNames(cfg.Jobs)
	if err != nil {
		log.Fatalf("Job validation failed: %v", err)
	}

	resolvedCfg := resolveConfig(cfg)

	err = ValidatePaths(resolvedCfg)
	if err != nil {
		log.Fatalf("Path validation failed: %v", err)
	}

	err = validateJobPaths(resolvedCfg.Jobs, "source", func(job Job) string { return job.Source })
	if err != nil {
		log.Fatalf("Job source path validation failed: %v", err)
	}

	err = validateJobPaths(resolvedCfg.Jobs, "target", func(job Job) string { return job.Target })
	if err != nil {
		log.Fatalf("Job target path validation failed: %v", err)
	}

	return resolvedCfg
}
