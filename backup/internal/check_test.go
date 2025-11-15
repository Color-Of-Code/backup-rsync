package internal_test

import (
	"bytes"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"backup-rsync/backup/internal"

	"github.com/spf13/afero"
)

func TestIsExcludedGlobally(t *testing.T) {
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

	tests := []struct {
		name         string
		path         string
		expectsError bool
		expectedLog  string
	}{
		{
			name:         "Path is globally excluded",
			path:         "/home/data/projects/P1",
			expectsError: true,
			expectedLog:  "Path '/home/data/projects/P1' is globally excluded by '/projects/P1/' in source '/home/data/'",
		},
		{
			name:         "Path is not excluded",
			path:         "/home/data/projects/Other",
			expectsError: false,
		},
		{
			name:         "Path is excluded in another source",
			path:         "/home/user/cache",
			expectsError: true,
			expectedLog:  "Path '/home/user/cache' is globally excluded by '/cache/' in source '/home/user/'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			log.SetOutput(&logBuffer)

			result := internal.IsExcludedGlobally(test.path, sources)
			if result != test.expectsError {
				t.Errorf("Expected exclusion result %v, got %v", test.expectsError, result)
			}

			if test.expectsError {
				if !strings.Contains(logBuffer.String(), test.expectedLog) {
					t.Errorf("Expected log message '%s', but got '%s'", test.expectedLog, logBuffer.String())
				}
			}
		})
	}
}

func runListUncoveredPathsTest(t *testing.T, fakeFS map[string][]string, cfg internal.Config, expectedUncoveredPaths []string) {
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

	if len(uncoveredPaths) != len(expectedUncoveredPaths) {
		t.Errorf("Expected uncovered paths length %d, got %d. Expected: %v, Got: %v",
			len(expectedUncoveredPaths), len(uncoveredPaths), expectedUncoveredPaths, uncoveredPaths)

		return
	}

	for i, path := range uncoveredPaths {
		if i >= len(expectedUncoveredPaths) {
			t.Errorf("Got more uncovered paths than expected. Got: %v", uncoveredPaths)

			return
		}

		if path != expectedUncoveredPaths[i] {
			t.Errorf("Expected uncovered path '%s', got '%s'", expectedUncoveredPaths[i], path)
		}
	}
}

func TestListUncoveredPathsVariations(t *testing.T) {
	// Variation: all paths used
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

	// Variation: one source covered, one uncovered
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

	// Variation: one source covered, one uncovered but excluded
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

	// Variation: one source covered, subfolders covered
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
