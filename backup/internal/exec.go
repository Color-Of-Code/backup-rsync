package internal

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Exec interface {
	Execute(name string, args ...string) ([]byte, error)
}

// OsExec implements Exec using actual os/exec.
type OsExec struct{}

// Execute runs the actual command.
func (r *OsExec) Execute(name string, args ...string) ([]byte, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command '%s %s': %w", name, strings.Join(args, " "), err)
	}

	return output, nil
}
