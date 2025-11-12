// Correct the package declaration
package internal

import "strings"

func NormalizePath(path string) string {
	return strings.TrimSuffix(strings.ReplaceAll(path, "//", "/"), "/")
}
