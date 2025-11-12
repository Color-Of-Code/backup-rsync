package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"path/filepath"

	"backup-rsync/backup/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func loadConfig(reader io.Reader) (internal.Config, error) {
	var cfg internal.Config
	if err := yaml.NewDecoder(reader).Decode(&cfg); err != nil {
		return internal.Config{}, err
	}
	return cfg, nil
}

func substituteVariables(input string, variables map[string]string) string {
	for key, value := range variables {
		placeholder := fmt.Sprintf("${%s}", key)
		input = strings.ReplaceAll(input, placeholder, value)
	}
	return input
}

func resolveConfig(cfg internal.Config) internal.Config {
	resolvedCfg := cfg
	for i, job := range resolvedCfg.Jobs {
		resolvedCfg.Jobs[i].Source = substituteVariables(job.Source, cfg.Variables)
		resolvedCfg.Jobs[i].Target = substituteVariables(job.Target, cfg.Variables)
	}
	return resolvedCfg
}

func validateJobNames(jobs []internal.Job) error {
	invalidNames := []string{}
	nameSet := make(map[string]bool)

	for _, job := range jobs {
		if nameSet[job.Name] {
			invalidNames = append(invalidNames, fmt.Sprintf("duplicate job name: %s", job.Name))
		} else {
			nameSet[job.Name] = true
		}

		for _, r := range job.Name {
			if r > 127 || r == ' ' {
				invalidNames = append(invalidNames, fmt.Sprintf("invalid characters in job name: %s", job.Name))
				break
			}
		}
	}

	if len(invalidNames) > 0 {
		return fmt.Errorf("job validation errors: %v", invalidNames)
	}
	return nil
}

func validatePath(jobPath string, paths []internal.Path, pathType string, jobName string) error {
	for _, path := range paths {
		if strings.HasPrefix(jobPath, path.Path) {
			return nil
		}
	}
	return fmt.Errorf("invalid %s path for job '%s': %s", pathType, jobName, jobPath)
}

func validatePaths(cfg internal.Config) error {
	invalidPaths := []string{}

	for _, job := range cfg.Jobs {
		if err := validatePath(job.Source, cfg.Sources, "source", job.Name); err != nil {
			invalidPaths = append(invalidPaths, err.Error())
		}
		if err := validatePath(job.Target, cfg.Targets, "target", job.Name); err != nil {
			invalidPaths = append(invalidPaths, err.Error())
		}
	}

	if len(invalidPaths) > 0 {
		return fmt.Errorf("path validation errors: %v", invalidPaths)
	}
	return nil
}

func validateJobPaths(jobs []internal.Job, pathType string, getPath func(job internal.Job) string) error {
	for i, job1 := range jobs {
		for j, job2 := range jobs {
			if i != j {
				path1, path2 := internal.NormalizePath(getPath(job1)), internal.NormalizePath(getPath(job2))

				// Check if path2 is part of job1's exclusions
				excluded := false
				if pathType == "source" {
					for _, exclusion := range job2.Exclusions {
						exclusionPath := internal.NormalizePath(filepath.Join(job2.Source, exclusion))
						// log.Printf("job2: %s %s\n", job2.Name, exclusionPath)
						if strings.HasPrefix(path1, exclusionPath) {
							excluded = true
							break
						}
					}
				}

				if !excluded && strings.HasPrefix(path1, path2) {
					return fmt.Errorf("Job '%s' has a %s path overlapping with job '%s'", job1.Name, pathType, job2.Name)
				}
			}
		}
	}
	return nil
}

func loadResolvedConfig(configPath string) internal.Config {
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config: %v", err)
	}
	defer f.Close()

	cfg, err := loadConfig(f)
	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	if err := validateJobNames(cfg.Jobs); err != nil {
		log.Fatalf("Job validation failed: %v", err)
	}

	resolvedCfg := resolveConfig(cfg)

	if err := validatePaths(resolvedCfg); err != nil {
		log.Fatalf("Path validation failed: %v", err)
	}

	if err := validateJobPaths(resolvedCfg.Jobs, "source", func(job internal.Job) string { return job.Source }); err != nil {
		log.Fatalf("Job source path validation failed: %v", err)
	}

	if err := validateJobPaths(resolvedCfg.Jobs, "target", func(job internal.Job) string { return job.Target }); err != nil {
		log.Fatalf("Job target path validation failed: %v", err)
	}

	return resolvedCfg
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation for the config command
		fmt.Println("Config command executed")
	},
}

// Extend the config subcommand with the show verb
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show resolved configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadResolvedConfig(configPath)
		out, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("Failed to marshal resolved configuration: %v", err)
		}
		fmt.Printf("Resolved Configuration:\n%s\n", string(out))
	},
}

// Extend the config subcommand with the validate verb
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Run: func(cmd *cobra.Command, args []string) {
		loadResolvedConfig(configPath)
		fmt.Println("Configuration is valid.")
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(validateCmd)
}
