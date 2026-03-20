package internal_test

import (
	"bytes"
	"io"
	"log"
	"path/filepath"
	"sort"
	"testing"

	. "backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func newTestChecker(fs afero.Fs, logBuf *bytes.Buffer) *CoverageChecker {
	return &CoverageChecker{
		Logger: log.New(logBuf, "", 0),
		Fs:     fs,
	}
}

func newSilentChecker(fs afero.Fs) *CoverageChecker {
	return &CoverageChecker{
		Logger: log.New(io.Discard, "", 0),
		Fs:     fs,
	}
}

func TestIsExcludedGlobally_PathGloballyExcluded(t *testing.T) {
	sources := []Path{
		{
			Path:       "/home/data/",
			Exclusions: []string{"/projects/P1/", "/media/"},
		},
		{
			Path:       "/home/user/",
			Exclusions: []string{"/cache/", "/npm/"},
		},
	}

	var logBuffer bytes.Buffer

	checker := newTestChecker(nil, &logBuffer)

	path := "/home/data/projects/P1"
	expectedLog := "Path '/home/data/projects/P1' is globally excluded by '/projects/P1/' in source '/home/data/'"

	result := checker.IsExcludedGlobally(path, sources)
	assert.True(t, result)
	assert.Contains(t, logBuffer.String(), expectedLog)
}

func TestIsExcludedGlobally_PathNotExcluded(t *testing.T) {
	sources := []Path{
		{
			Path:       "/home/data/",
			Exclusions: []string{"/projects/P1/", "/media/"},
		},
		{
			Path:       "/home/user/",
			Exclusions: []string{"/cache/", "/npm/"},
		},
	}

	checker := newSilentChecker(nil)

	path := "/home/data/projects/Other"

	result := checker.IsExcludedGlobally(path, sources)
	assert.False(t, result)
}

func TestIsExcludedGlobally_PathExcludedInAnotherSource(t *testing.T) {
	sources := []Path{
		{
			Path:       "/home/data/",
			Exclusions: []string{"/projects/P1/", "/media/"},
		},
		{
			Path:       "/home/user/",
			Exclusions: []string{"/cache/", "/npm/"},
		},
	}

	var logBuffer bytes.Buffer

	checker := newTestChecker(nil, &logBuffer)

	path := "/home/user/cache"
	expectedLog := "Path '/home/user/cache' is globally excluded by '/cache/' in source '/home/user/'"

	result := checker.IsExcludedGlobally(path, sources)
	assert.True(t, result)
	assert.Contains(t, logBuffer.String(), expectedLog)
}

func runListUncoveredPathsTest(
	t *testing.T,
	fakeFS map[string][]string,
	cfg Config,
	expectedUncoveredPaths []string,
) {
	t.Helper()
	// Create an in-memory filesystem using Afero
	fs := afero.NewMemMapFs()

	// Populate the in-memory filesystem with the fakeFS structure
	for path, entries := range fakeFS {
		_ = fs.MkdirAll(path, 0755)
		for _, entry := range entries {
			entryPath := filepath.Join(path, entry)
			_ = fs.MkdirAll(entryPath, 0755)
		}
	}

	checker := newSilentChecker(fs)

	// Call the function
	uncoveredPaths := checker.ListUncoveredPaths(cfg)

	// Assertions
	sort.Strings(uncoveredPaths)
	sort.Strings(expectedUncoveredPaths)

	assert.Len(t, uncoveredPaths, len(expectedUncoveredPaths))
	assert.ElementsMatch(t, expectedUncoveredPaths, uncoveredPaths)
}

// Variation: all paths used.
func TestListUncoveredPathsVariationsAllCovered(t *testing.T) {
	runListUncoveredPathsTest(t,
		map[string][]string{
			"/var/log": {"app1", "app2"},
			"/tmp":     {"cache", "temp"},
		},
		Config{
			Sources: []Path{
				{Path: "/var/log"},
				{Path: "/tmp"},
			},
			Jobs: []Job{
				{Name: "Job1", Source: "/var/log"},
				{Name: "Job2", Source: "/tmp"},
			},
		},
		[]string{},
	)
}

// Variation: one source covered, one uncovered.
func TestListUncoveredPathsVariationsOneCoveredOneUncovered(t *testing.T) {
	runListUncoveredPathsTest(t,
		map[string][]string{
			"/home/data":       {"projects", "media"},
			"/home/user":       {"cache", "npm"},
			"/home/user/cache": {},
			"/home/user/npm":   {},
		},
		Config{
			Sources: []Path{
				{Path: "/home/data"},
				{Path: "/home/user"},
			},
			Jobs: []Job{
				{Name: "Job1", Source: "/home/data"},
			},
		},
		[]string{"/home/user"},
	)
}

// Variation: one source covered, one uncovered but excluded.
func TestListUncoveredPathsVariationsUncoveredExcluded(t *testing.T) {
	runListUncoveredPathsTest(t,
		map[string][]string{
			"/home/data": {"projects", "media"},
		},
		Config{
			Sources: []Path{
				{Path: "/home/data", Exclusions: []string{"media"}},
			},
			Jobs: []Job{
				{Name: "Job1", Source: "/home/data/projects"},
			},
		},
		[]string{},
	)
}

// Variation: one source covered, subfolders covered.
func TestListUncoveredPathsVariationsSubfoldersCovered(t *testing.T) {
	runListUncoveredPathsTest(t,
		map[string][]string{
			"/home/data":            {"family"},
			"/home/data/family":     {"me", "you"},
			"/home/data/family/me":  {"a"},
			"/home/data/family/you": {"a"},
		},
		Config{
			Sources: []Path{
				{Path: "/home/data"},
			},
			Jobs: []Job{
				{Name: "JobMe", Source: "/home/data/family/me"},
				{Name: "JobYou", Source: "/home/data/family/you"},
			},
		},
		[]string{},
	)
}

func TestListUncoveredPathsVariationsSubfoldersPartiallyCovered(t *testing.T) {
	t.Skip("Skipping test for partially covered subfolders")
}

// Test that a job with exclusions properly marks child paths as excluded.
func TestListUncoveredPaths_JobExclusion(t *testing.T) {
	runListUncoveredPathsTest(t,
		map[string][]string{
			"/data":       {"docs", "cache"},
			"/data/docs":  {},
			"/data/cache": {},
		},
		Config{
			Sources: []Path{
				{Path: "/data"},
			},
			Jobs: []Job{
				{Name: "backup", Source: "/data/", Exclusions: []string{"cache"}},
			},
		},
		[]string{},
	)
}

// Test that duplicate source paths are processed only once.
func TestListUncoveredPaths_DuplicateSourcesSkipped(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/data", 0755)

	var logBuf bytes.Buffer

	checker := newTestChecker(fs, &logBuf)

	cfg := Config{
		Sources: []Path{
			{Path: "/data"},
			{Path: "/data"},
		},
		Jobs: []Job{
			{Name: "backup", Source: "/data"},
		},
	}

	result := checker.ListUncoveredPaths(cfg)

	assert.Empty(t, result)
	assert.Contains(t, logBuf.String(), "SKIP: Path '/data' already seen")
}

// Test getChildDirectories error path (unreadable directory).
func TestListUncoveredPaths_UnreadableDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Don't create /data, so ReadDir will fail

	var logBuf bytes.Buffer

	checker := newTestChecker(fs, &logBuf)

	cfg := Config{
		Sources: []Path{
			{Path: "/data"},
		},
		Jobs: []Job{},
	}

	result := checker.ListUncoveredPaths(cfg)

	assert.Equal(t, []string{"/data"}, result)
	assert.Contains(t, logBuf.String(), "ADD: Path '/data' is uncovered")
}
