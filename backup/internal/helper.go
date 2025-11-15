// Package internal provides helper functions for internal use within the application.
package internal

import "strings"

func NormalizePath(path string) string {
	return strings.TrimSuffix(strings.ReplaceAll(path, "//", "/"), "/")
}
