package internal_test

import (
	"errors"
	"strings"
)

// Static error for testing.
var ErrExitStatus23 = errors.New("exit status 23")

// MockCommandExecutor implements CommandExecutor for testing.
type MockCommandExecutor struct {
	CapturedCommands []MockCommand
	Output           string
	Error            error
}

// MockCommand represents a captured command execution.
type MockCommand struct {
	Name string
	Args []string
}

// Execute captures the command and simulates execution.
func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	m.CapturedCommands = append(m.CapturedCommands, MockCommand{
		Name: name,
		Args: append([]string{}, args...), // Make a copy of args
	})

	// If Error is set, return it.
	if m.Error != nil {
		return nil, m.Error
	}

	// If Output is set, return it.
	if m.Output != "" {
		return []byte(m.Output), nil
	}

	// Simulate specific scenarios for rsync.
	if name == "rsync" {
		argsStr := strings.Join(args, " ")

		if strings.Contains(argsStr, "/invalid/source/path") {
			errMsg := "rsync: link_stat \"/invalid/source/path\" failed: No such file or directory"

			return []byte(errMsg), ErrExitStatus23
		}

		return []byte("mocked rsync success"), nil
	}

	return []byte("command not mocked"), nil
}
