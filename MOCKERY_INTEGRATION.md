# Mockery Integration Guide

This document explains how mockery is integrated for generating mocks from interfaces.

## Installation

```bash
go install github.com/vektra/mockery/v3@latest
```

## Configuration

The project uses `.mockery.yml` to control mock generation:

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

Key points:
- **Output directory**: `<InterfaceDir>/test/` (alongside other test files)
- **Filename**: `mock_<interface>_test.go`
- **Struct naming**: `Mock<Interface>` (e.g., `MockExec`, `MockJobCommand`)
- **Package**: `internal_test` (external test package)
- **Template**: `testify` for expectation-based mocking

## Generated Mocks

| Mock | Source Interface | File |
|---|---|---|
| `MockExec` | `Exec` | `backup/internal/test/mock_exec_test.go` |
| `MockJobCommand` | `JobCommand` | `backup/internal/test/mock_jobcommand_test.go` |

## Usage Examples

### MockJobCommand — Testing Config.Apply

```go
func TestConfigApply_VersionInfoSuccess(t *testing.T) {
    mockCmd := NewMockJobCommand(t)
    var output bytes.Buffer
    logger := log.New(&bytes.Buffer{}, "", 0)

    cfg := Config{
        Jobs: []Job{
            {Name: "job1", Source: "/src/", Target: "/dst/", Enabled: true},
            {Name: "job2", Source: "/src2/", Target: "/dst2/", Enabled: false},
        },
    }

    mockCmd.EXPECT().GetVersionInfo().Return("rsync version 3.2.3", "/usr/bin/rsync", nil).Once()
    mockCmd.EXPECT().Run(mock.AnythingOfType("internal.Job")).Return(Success).Once()

    err := cfg.Apply(mockCmd, logger, &output)
    require.NoError(t, err)
    assert.Contains(t, output.String(), "Status [job1]: SUCCESS")
}
```

### MockExec — Testing Command Execution

```go
func TestSyncCommand_Run_Success(t *testing.T) {
    mockExec := NewMockExec(t)
    var output bytes.Buffer

    cmd := NewSyncCommand("/usr/bin/rsync", "/tmp/logs", mockExec, &output)
    job := Job{Name: "docs", Source: "/src/", Target: "/dst/", Enabled: true, Delete: true}

    mockExec.EXPECT().Execute("/usr/bin/rsync", mock.Anything).Return([]byte("done"), nil).Once()

    status := cmd.Run(job)
    assert.Equal(t, Success, status)
}
```

### Testing Disabled Jobs (no mock expectations needed)

```go
func TestJobApply_DisabledJob(t *testing.T) {
    mockCmd := NewMockJobCommand(t)
    disabledJob := Job{Name: "skip_me", Enabled: false}

    // No expectations set — Run should NOT be called
    status := disabledJob.Apply(mockCmd)
    assert.Equal(t, Skipped, status)
    // MockJobCommand automatically verifies Run was not called
}
```

## Regenerating Mocks

When interfaces change, regenerate with:

```bash
mockery
```

This updates all mocks according to `.mockery.yml`. Generated files are committed to the repository.