package internal

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
)

// CoverageChecker analyzes path coverage against a configuration.
type CoverageChecker struct {
	Logger *slog.Logger
	Fs     afero.Fs
}

func (c *CoverageChecker) IsExcludedGlobally(path string, mappings []Mapping) bool {
	for _, mapping := range mappings {
		for _, exclusion := range mapping.Exclusions {
			exclusionPath := filepath.Join(mapping.Source, exclusion)
			if strings.HasPrefix(NormalizePath(path), exclusionPath) {
				c.Logger.Info(fmt.Sprintf("EXCLUDED: Path '%s' is globally excluded by '%s' in source '%s'",
					path, exclusion, mapping.Source))

				return true
			}
		}
	}

	return false
}

func (c *CoverageChecker) ListUncoveredPaths(cfg Config) []string {
	var result []string

	seen := make(map[string]bool)

	for _, mapping := range cfg.Mappings {
		c.checkPath(mapping.Source, cfg.Mappings, &result, seen)
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
		c.Logger.Info(fmt.Sprintf("COVERED: Path '%s' is covered by job '%s'", path, job.Name))

		return true
	}

	if c.isExcluded(path, job) {
		c.Logger.Info(fmt.Sprintf("EXCLUDED: Path '%s' is excluded by job '%s'", path, job.Name))

		return true
	}

	return false
}

func (c *CoverageChecker) isCovered(path string, mappings []Mapping) bool {
	for _, mapping := range mappings {
		if slices.ContainsFunc(mapping.Jobs, func(job Job) bool {
			return c.isCoveredByJob(path, job)
		}) {
			return true
		}
	}

	return false
}

func (c *CoverageChecker) checkPath(
	path string, mappings []Mapping, result *[]string, seen map[string]bool,
) {
	if seen[path] {
		c.Logger.Info(fmt.Sprintf("SKIP: Path '%s' already seen", path))

		return
	}

	seen[path] = true

	// Skip if globally excluded
	if c.IsExcludedGlobally(path, mappings) {
		c.Logger.Info(fmt.Sprintf("SKIP: Path '%s' is globally excluded", path))

		return
	}

	// Skip if covered by a job
	if c.isCovered(path, mappings) {
		c.Logger.Info(fmt.Sprintf("SKIP: Path '%s' is covered by a job", path))

		return
	}

	// Check if it's effectively covered through descendants
	if c.isEffectivelyCovered(path, mappings) {
		c.Logger.Info(fmt.Sprintf("SKIP: Path '%s' is effectively covered", path))

		return
	}

	// Add uncovered path
	c.Logger.Info(fmt.Sprintf("ADD: Path '%s' is uncovered", path))
	*result = append(*result, path)
}

// isEffectivelyCovered checks if a directory is effectively covered
// (all its descendants are covered or excluded).
func (c *CoverageChecker) isEffectivelyCovered(path string, mappings []Mapping) bool {
	children, err := getChildDirectories(c.Fs, path)
	if err != nil {
		c.Logger.Info(fmt.Sprintf("ERROR: could not get child directories of '%s': %v", path, err))

		return false
	}

	if len(children) == 0 {
		c.Logger.Info(fmt.Sprintf("NOT COVERED: Path '%s' has no children", path))

		return false // Leaf directories are not effectively covered unless directly covered
	}

	allCovered := true

	for _, child := range children {
		covered := c.IsExcludedGlobally(child, mappings) || c.isCovered(child, mappings) ||
			c.isEffectivelyCovered(child, mappings)
		if !covered {
			c.Logger.Info(fmt.Sprintf("UNCOVERED CHILD: Path '%s' has uncovered child '%s'", path, child))

			allCovered = false
		}
	}

	if allCovered {
		c.Logger.Info(fmt.Sprintf("COVERED: Path '%s' is effectively covered", path))
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
