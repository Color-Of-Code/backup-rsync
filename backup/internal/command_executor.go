package internal

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandExecutor interface for executing commands.
type CommandExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using actual os/exec.
type RealCommandExecutor struct{}

// Execute runs the actual command.
func (r *RealCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command '%s %s': %w", name, strings.Join(args, " "), err)
	}

	return output, nil
}
