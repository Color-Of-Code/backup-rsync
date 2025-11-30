package internal

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// JobRunner interface for executing commands.
type JobRunner interface {
	Execute(name string, args ...string) ([]byte, error)
}

// RealSync implements JobRunner using actual os/exec.
type RealSync struct{}

// Execute runs the actual command.
func (r *RealSync) Execute(name string, args ...string) ([]byte, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command '%s %s': %w", name, strings.Join(args, " "), err)
	}

	return output, nil
}
