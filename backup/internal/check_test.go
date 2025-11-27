package internal_test

import (
	"bytes"
	"log"
	"path/filepath"
	"sort"
	"testing"

	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsExcludedGlobally_PathGloballyExcluded(t *testing.T) {
	sources := []internal.Path{
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
	log.SetOutput(&logBuffer)

	path := "/home/data/projects/P1"
	expectsError := true
	expectedLog := "Path '/home/data/projects/P1' is globally excluded by '/projects/P1/' in source '/home/data/'"

	result := internal.IsExcludedGlobally(path, sources)
	assert.Equal(t, expectsError, result)
	assert.Contains(t, logBuffer.String(), expectedLog)
}

func TestIsExcludedGlobally_PathNotExcluded(t *testing.T) {
	sources := []internal.Path{
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
	log.SetOutput(&logBuffer)

	path := "/home/data/projects/Other"
	expectsError := false

	result := internal.IsExcludedGlobally(path, sources)
	assert.Equal(t, expectsError, result)
}

func TestIsExcludedGlobally_PathExcludedInAnotherSource(t *testing.T) {
	sources := []internal.Path{
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
	log.SetOutput(&logBuffer)

	path := "/home/user/cache"
	expectsError := true
	expectedLog := "Path '/home/user/cache' is globally excluded by '/cache/' in source '/home/user/'"

	result := internal.IsExcludedGlobally(path, sources)
	assert.Equal(t, expectsError, result)
	assert.Contains(t, logBuffer.String(), expectedLog)
}

func runListUncoveredPathsTest(
	t *testing.T,
	fakeFS map[string][]string,
	cfg internal.Config,
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

	// Call the function
	uncoveredPaths := internal.ListUncoveredPaths(fs, cfg)

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
		internal.Config{
			Sources: []internal.Path{
				{Path: "/var/log"},
				{Path: "/tmp"},
			},
			Jobs: []internal.Job{
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
		internal.Config{
			Sources: []internal.Path{
				{Path: "/home/data"},
				{Path: "/home/user"},
			},
			Jobs: []internal.Job{
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
		internal.Config{
			Sources: []internal.Path{
				{Path: "/home/data", Exclusions: []string{"media"}},
			},
			Jobs: []internal.Job{
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
		internal.Config{
			Sources: []internal.Path{
				{Path: "/home/data"},
			},
			Jobs: []internal.Job{
				{Name: "JobMe", Source: "/home/data/family/me"},
				{Name: "JobYou", Source: "/home/data/family/you"},
			},
		},
		[]string{},
	)
}

func TestListUncoveredPathsVariationsSubfoldersPartiallyCovered(t *testing.T) {
	t.Skip("Skipping test for partially covered subfolders")
	// // Variation: one source covered, one uncovered subfolder
	// runListUncoveredPathsTest(t,
	// 	map[string][]string{
	// 		"/home/data":            {"family"},
	// 		"/home/data/family":     {"me", "you"},
	// 		"/home/data/family/me":  {"a"},
	// 		"/home/data/family/you": {"a"},
	// 	},
	// 	Config{
	// 		Sources: []Path{
	// 			{Path: "/home/data"},
	// 		},
	// 		Jobs: []Job{
	// 			{Name: "JobMe", Source: "/home/data/family/me"},
	// 		},
	// 	},
	// 	[]string{"/home/data/family/you"},
	// )
}
