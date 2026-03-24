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

func TestIsExcludedGlobally(t *testing.T) {
	sources := []Path{
		{Path: "/home/data/", Exclusions: []string{"/projects/P1/", "/media/"}},
		{Path: "/home/user/", Exclusions: []string{"/cache/", "/npm/"}},
	}

	tests := []struct {
		name, path, wantLog string
		want                bool
	}{
		{"PathGloballyExcluded", "/home/data/projects/P1",
			"globally excluded by '/projects/P1/' in source '/home/data/'", true},
		{"PathNotExcluded", "/home/data/projects/Other", "", false},
		{"PathExcludedInAnotherSource", "/home/user/cache",
			"globally excluded by '/cache/' in source '/home/user/'", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logBuf bytes.Buffer

			checker := newTestChecker(nil, &logBuf)

			result := checker.IsExcludedGlobally(test.path, sources)

			assert.Equal(t, test.want, result)

			if test.wantLog != "" {
				assert.Contains(t, logBuf.String(), test.wantLog)
			}
		})
	}
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

func TestListUncoveredPathsVariations(t *testing.T) {
	tests := []struct {
		name      string
		fakeFS    map[string][]string
		cfg       Config
		wantPaths []string
	}{
		{name: "AllCovered",
			fakeFS: map[string][]string{"/var/log": {"app1", "app2"}, "/tmp": {"cache", "temp"}},
			cfg: Config{Mappings: []Mapping{
				{Name: "logs", Source: "/var/log", Target: "/bak/log", Jobs: []Job{{Name: "Job1", Source: "/var/log"}}},
				{Name: "tmp", Source: "/tmp", Target: "/bak/tmp", Jobs: []Job{{Name: "Job2", Source: "/tmp"}}},
			}}},
		{name: "OneCoveredOneUncovered",
			fakeFS: map[string][]string{
				"/home/data": {"projects", "media"}, "/home/user": {"cache", "npm"},
				"/home/user/cache": {}, "/home/user/npm": {},
			},
			cfg: Config{Mappings: []Mapping{
				{Name: "data", Source: "/home/data", Target: "/bak/data", Jobs: []Job{{Name: "Job1", Source: "/home/data"}}},
				{Name: "user", Source: "/home/user", Target: "/bak/user", Jobs: []Job{}},
			}},
			wantPaths: []string{"/home/user"}},
		{name: "UncoveredExcluded",
			fakeFS: map[string][]string{"/home/data": {"projects", "media"}},
			cfg: Config{Mappings: []Mapping{
				{Name: "data", Source: "/home/data", Target: "/bak/data", Exclusions: []string{"media"},
					Jobs: []Job{{Name: "Job1", Source: "/home/data/projects"}}},
			}}},
		{name: "SubfoldersCovered",
			fakeFS: map[string][]string{
				"/home/data": {"family"}, "/home/data/family": {"me", "you"},
				"/home/data/family/me": {"a"}, "/home/data/family/you": {"a"},
			},
			cfg: Config{Mappings: []Mapping{
				{Name: "data", Source: "/home/data", Target: "/bak/data", Jobs: []Job{
					{Name: "JobMe", Source: "/home/data/family/me"},
					{Name: "JobYou", Source: "/home/data/family/you"},
				}},
			}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runListUncoveredPathsTest(t, test.fakeFS, test.cfg, test.wantPaths)
		})
	}
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
			Mappings: []Mapping{
				{Name: "data", Source: "/data", Target: "/bak/data", Jobs: []Job{
					{Name: "backup", Source: "/data/", Exclusions: []string{"cache"}},
				}},
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
		Mappings: []Mapping{
			{Name: "m1", Source: "/data", Target: "/bak1", Jobs: []Job{{Name: "backup", Source: "/data"}}},
			{Name: "m2", Source: "/data", Target: "/bak2", Jobs: []Job{}},
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
		Mappings: []Mapping{
			{Name: "data", Source: "/data", Target: "/bak", Jobs: []Job{}},
		},
	}

	result := checker.ListUncoveredPaths(cfg)

	assert.Equal(t, []string{"/data"}, result)
	assert.Contains(t, logBuf.String(), "ADD: Path '/data' is uncovered")
}

// Test that a child path matching a job exclusion is marked as excluded
// (covers isExcluded true + isCoveredByJob excluded log).
func TestListUncoveredPaths_ChildPathExcludedByJob(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/data/stuff/docs", 0755)
	_ = fs.MkdirAll("/data/stuff/cache", 0755)

	var logBuf bytes.Buffer

	checker := newTestChecker(fs, &logBuf)

	cfg := Config{
		Mappings: []Mapping{
			{Name: "stuff", Source: "/data/stuff", Target: "/bak/stuff", Jobs: []Job{
				// Source "/data" with exclusion "stuff/cache" so exclusionPath = "/data/stuff/cache"
				{Name: "data-backup", Source: "/data", Exclusions: []string{"stuff/cache"}},
				// Covers the /data/stuff/docs child directly
				{Name: "docs", Source: "/data/stuff/docs"},
			}},
		},
	}

	result := checker.ListUncoveredPaths(cfg)

	assert.Empty(t, result)
	assert.Contains(t, logBuf.String(), "EXCLUDED: Path '/data/stuff/cache' is excluded by job 'data-backup'")
}

// Test that a source path that is globally excluded is skipped in checkPath.
func TestListUncoveredPaths_GloballyExcludedSourceSkipped(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/data/cache", 0755)

	var logBuf bytes.Buffer

	checker := newTestChecker(fs, &logBuf)

	cfg := Config{
		Mappings: []Mapping{
			{Name: "data", Source: "/data", Target: "/bak/data", Exclusions: []string{"cache"},
				Jobs: []Job{{Name: "backup", Source: "/data"}}},
			{Name: "cache", Source: "/data/cache", Target: "/bak/cache", Jobs: []Job{}},
		},
	}

	result := checker.ListUncoveredPaths(cfg)

	assert.Empty(t, result)
	assert.Contains(t, logBuf.String(), "SKIP: Path '/data/cache' is globally excluded")
}
