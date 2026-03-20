# Mockery Integration Guide

This document explains how mockery has been integrated into this Go project for generating mocks from interfaces.

## Installation

Mockery v3.6.1 has been installed in the project:

```bash
go install github.com/vektra/mockery/v3@v3.6.1
```

## Configuration

The project uses a `.mockery.yml` configuration file to control mock generation:

```yaml
all: false
dir: '{{.InterfaceDir}}/test'
filename: mock_{{.InterfaceName | lower}}_test.go
force-file-write: true
formatter: goimports
generate: true
include-auto-generated: false
log-level: info
structname: 'Mock{{.InterfaceName}}'
pkgname: 'internal_test'
recursive: false
template: testify
packages:
  backup-rsync/backup/internal:
    interfaces:
      Exec:
      JobCommand:
```

Key configuration points:
- **Filename pattern**: `mock_{{.InterfaceName | lower}}_test.go` for standard naming
- **Struct naming**: `Mock{{.InterfaceName}}` for standard mock naming
- **Package**: `internal_test` to match the test package
- **Template**: Uses the `testify` template for rich mock functionality

## Generated Mocks

Running `mockery` generates the following mock files:

- `backup/internal/test/mock_exec_test.go` - Mock for the `Exec` interface
- `backup/internal/test/mock_jobcommand_test.go` - Mock for the `JobCommand` interface

These mocks provide:
- **Type-safe mocking** with compile-time verification
- **Expectation-based testing** with automatic verification
- **Fluent interface** for setting up complex expectations
- **Automatic cleanup** via `t.Cleanup()`

## Usage Examples

### Basic Usage

```go
func TestJobApply_WithMockeryJobCommand_Success(t *testing.T) {
    // Create mock using mockery's generated constructor
    mockJobCommand := NewMockJobCommand(t)
    
    // Create an enabled job
    enabledJob := NewJob(
        WithName("success_job"),
        WithSource("/home/success/"),
        WithTarget("/mnt/backup1/success/"),
        WithEnabled(true),
    )

    // Set expectation that Run will be called and return Success
    mockJobCommand.EXPECT().Run(mock.MatchedBy(func(job Job) bool {
        return job.Name == "success_job" && 
               job.Source == "/home/success/" &&
               job.Target == "/mnt/backup1/success/" &&
               job.Enabled == true
    })).Return(Success).Once()
    
    // Apply the job
    status := enabledJob.Apply(mockJobCommand)

    // Assert that the status is Success
    assert.Equal(t, Success, status)
    
    // mockery automatically verifies expectations via t.Cleanup()
}
```

### Testing Disabled Jobs

```go
func TestJobApply_WithMockeryJobCommand_DisabledJob(t *testing.T) {
    // Create mock using mockery's generated constructor
    mockJobCommand := NewMockJobCommand(t)
    
    // Create a disabled job
    disabledJob := NewJob(
        WithName("disabled_job"),
        WithEnabled(false),
    )

    // No expectations set - the Run method should NOT be called for disabled jobs
    
    // Apply the job
    status := disabledJob.Apply(mockJobCommand)

    // Assert that the status is Skipped
    assert.Equal(t, Skipped, status)
    
    // The mock will automatically verify that Run was NOT called
}
```

## Replacement of Simple Mocks

The project has fully migrated from simple mocks to mockery-generated mocks:

- **Old simple mocks**: Removed (used field-based APIs like `CapturedCommands`, `Output`, `Error`)
- **New mockery mocks** (`MockExec`, `MockJobCommand`): Used for all testing scenarios with rich expectations and automatic verification

## Benefits of Mockery Integration

1. **Type Safety**: Compile-time verification of mock method signatures
2. **Rich Expectations**: Support for complex argument matching and call counting
3. **Automatic Verification**: Expectations are automatically verified at test completion
4. **Fluent Interface**: Easy-to-read test setup with method chaining
5. **Maintenance**: Mocks stay in sync with interface changes automatically

## Regenerating Mocks

When interfaces change, regenerate mocks with:

```bash
mockery
```

This will update all configured mocks according to the `.mockery.yml` configuration.

## Dependencies

The project now includes the testify mock dependency:

```go
github.com/stretchr/testify/mock v1.11.1
```

This enables the full mockery feature set including expectations, matchers, and automatic verification.