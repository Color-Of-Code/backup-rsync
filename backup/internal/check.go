package internal

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
)

// CoverageChecker analyzes path coverage against a configuration.
type CoverageChecker struct {
	Logger *log.Logger
	Fs     afero.Fs
}

func (c *CoverageChecker) IsExcludedGlobally(path string, sources []Path) bool {
	for _, source := range sources {
		for _, exclusion := range source.Exclusions {
			exclusionPath := filepath.Join(source.Path, exclusion)
			if strings.HasPrefix(NormalizePath(path), exclusionPath) {
				c.Logger.Printf("EXCLUDED: Path '%s' is globally excluded by '%s' in source '%s'", path, exclusion, source.Path)

				return true
			}
		}
	}

	return false
}

func (c *CoverageChecker) ListUncoveredPaths(cfg Config) []string {
	var result []string

	seen := make(map[string]bool)
	sources := cfg.AllSources()
	allJobs := cfg.AllJobs()

	for _, source := range sources {
		c.checkPath(source.Path, sources, allJobs, &result, seen)
	}

	slices.Sort(result) // Ensure consistent ordering for test comparison

	return result
}

func (c *CoverageChecker) isExcluded(path string, job Job) bool {
	normalized := NormalizePath(path)

	return slices.ContainsFunc(job.Exclusions, func(exclusion string) bool {
		return strings.HasPrefix(normalized, filepath.Join(job.Source, exclusion))
	})
}

func (c *CoverageChecker) isCoveredByJob(path string, job Job) bool {
	if NormalizePath(job.Source) == NormalizePath(path) {
		c.Logger.Printf("COVERED: Path '%s' is covered by job '%s'", path, job.Name)

		return true
	}

	if c.isExcluded(path, job) {
		c.Logger.Printf("EXCLUDED: Path '%s' is excluded by job '%s'", path, job.Name)

		return true
	}

	return false
}

func (c *CoverageChecker) isCovered(path string, jobs []Job) bool {
	return slices.ContainsFunc(jobs, func(job Job) bool {
		return c.isCoveredByJob(path, job)
	})
}

func (c *CoverageChecker) checkPath(path string, sources []Path, jobs []Job, result *[]string, seen map[string]bool) {
	if seen[path] {
		c.Logger.Printf("SKIP: Path '%s' already seen", path)

		return
	}

	seen[path] = true

	// Skip if globally excluded
	if c.IsExcludedGlobally(path, sources) {
		c.Logger.Printf("SKIP: Path '%s' is globally excluded", path)

		return
	}

	// Skip if covered by a job
	if c.isCovered(path, jobs) {
		c.Logger.Printf("SKIP: Path '%s' is covered by a job", path)

		return
	}

	// Check if it's effectively covered through descendants
	if c.isEffectivelyCovered(path, sources, jobs) {
		c.Logger.Printf("SKIP: Path '%s' is effectively covered", path)

		return
	}

	// Add uncovered path
	c.Logger.Printf("ADD: Path '%s' is uncovered", path)
	*result = append(*result, path)
}

// isEffectivelyCovered checks if a directory is effectively covered
// (all its descendants are covered or excluded).
func (c *CoverageChecker) isEffectivelyCovered(path string, sources []Path, jobs []Job) bool {
	children, err := getChildDirectories(c.Fs, path)
	if err != nil {
		c.Logger.Printf("ERROR: could not get child directories of '%s': %v", path, err)

		return false
	}

	if len(children) == 0 {
		c.Logger.Printf("NOT COVERED: Path '%s' has no children", path)

		return false // Leaf directories are not effectively covered unless directly covered
	}

	allCovered := true

	for _, child := range children {
		covered := c.IsExcludedGlobally(child, sources) || c.isCovered(child, jobs) ||
			c.isEffectivelyCovered(child, sources, jobs)
		if !covered {
			c.Logger.Printf("UNCOVERED CHILD: Path '%s' has uncovered child '%s'", path, child)

			allCovered = false
		}
	}

	if allCovered {
		c.Logger.Printf("COVERED: Path '%s' is effectively covered", path)
	}

	return allCovered
}

func getChildDirectories(fs afero.Fs, path string) ([]string, error) {
	var children []string

	fileInfos, err := afero.ReadDir(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", path, err)
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			children = append(children, filepath.Join(path, info.Name()))
		}
	}

	return children, nil
}
