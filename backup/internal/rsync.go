package internal

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var ErrInvalidRsyncVersion = errors.New("invalid rsync version output")
var ErrInvalidRsyncPath = errors.New("rsync path must be an absolute path")

func FetchRsyncVersion(executor CommandExecutor, rsyncPath string) (string, error) {
	if !filepath.IsAbs(rsyncPath) {
		return "", fmt.Errorf("%w: \"%s\"", ErrInvalidRsyncPath, rsyncPath)
	}

	cmdArgs := BuildRsyncVersionCmd()

	output, err := executor.Execute(rsyncPath, cmdArgs...)
	if err != nil {
		return "", fmt.Errorf("error fetching rsync version: %w", err)
	}

	// Validate output
	if !strings.Contains(string(output), "rsync") || !strings.Contains(string(output), "protocol version") {
		return "", fmt.Errorf("%w: %s", ErrInvalidRsyncVersion, output)
	}

	return string(output), nil
}
