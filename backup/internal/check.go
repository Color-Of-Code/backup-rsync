package internal

import (
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

func isExcluded(path string, job Job) bool {
	for _, exclusion := range job.Exclusions {
		exclusionPath := filepath.Join(job.Source, exclusion)
		if strings.HasPrefix(NormalizePath(path), exclusionPath) {
			return true
		}
	}
	return false
}

func isExcludedGlobally(path string, sources []Path) bool {
	for _, source := range sources {
		for _, exclusion := range source.Exclusions {
			exclusionPath := filepath.Join(source.Path, exclusion)
			if strings.HasPrefix(NormalizePath(path), exclusionPath) {
				log.Printf("EXCLUDED: Path '%s' is globally excluded by '%s' in source '%s'", path, exclusion, source.Path)
				return true
			}
		}
	}
	return false
}

func isCoveredByJob(path string, job Job) bool {
	if NormalizePath(job.Source) == NormalizePath(path) {
		log.Printf("COVERED: Path '%s' is covered by job '%s'", path, job.Name)
		return true
	}
	if isExcluded(path, job) {
		log.Printf("EXCLUDED: Path '%s' is excluded by job '%s'", path, job.Name)
		return true
	}
	return false
}

func isCovered(path string, jobs []Job) bool {
	for _, job := range jobs {
		if isCoveredByJob(path, job) {
			return true
		}
	}
	return false
}

func listUncoveredPaths(fs afero.Fs, cfg Config) []string {
	var result []string
	seen := make(map[string]bool)

	for _, source := range cfg.Sources {
		checkPath(fs, source.Path, cfg, &result, seen)
	}

	sort.Strings(result) // Ensure consistent ordering for test comparison
	return result
}

func checkPath(fs afero.Fs, path string, cfg Config, result *[]string, seen map[string]bool) {
	if seen[path] {
		log.Printf("SKIP: Path '%s' already seen", path)
		return
	}
	seen[path] = true

	// Skip if globally excluded
	if isExcludedGlobally(path, cfg.Sources) {
		log.Printf("SKIP: Path '%s' is globally excluded", path)
		return
	}

	// Skip if covered by a job
	if isCovered(path, cfg.Jobs) {
		log.Printf("SKIP: Path '%s' is covered by a job", path)
		return
	}

	// Check if it's effectively covered through descendants
	if isEffectivelyCovered(fs, path, cfg) {
		log.Printf("SKIP: Path '%s' is effectively covered", path)
		return
	}

	// Add uncovered path
	log.Printf("ADD: Path '%s' is uncovered", path)
	*result = append(*result, path)
}

// Check if a directory is effectively covered (all its descendants are covered or excluded)
func isEffectivelyCovered(fs afero.Fs, path string, cfg Config) bool {
	children, err := getChildDirectories(fs, path)
	if err != nil {
		log.Printf("ERROR: could not get child directories of '%s': %v", path, err)
		return false
	}

	if len(children) == 0 {
		log.Printf("NOT COVERED: Path '%s' has no children", path)
		return false // Leaf directories are not effectively covered unless directly covered
	}

	allCovered := true
	for _, child := range children {
		if !isExcludedGlobally(child, cfg.Sources) && !isCovered(child, cfg.Jobs) && !isEffectivelyCovered(fs, child, cfg) {
			log.Printf("UNCOVERED CHILD: Path '%s' has uncovered child '%s'", path, child)
			allCovered = false
		}
	}

	if allCovered {
		log.Printf("COVERED: Path '%s' is effectively covered", path)
	}
	return allCovered
}

func getChildDirectories(fs afero.Fs, path string) ([]string, error) {
	var children []string
	fileInfos, err := afero.ReadDir(fs, path)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			children = append(children, filepath.Join(path, info.Name()))
		}
	}
	return children, nil
}
