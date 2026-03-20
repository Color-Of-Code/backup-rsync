# Testing Guide

## Overview

All tests use dependency injection — no global state mutation. Key patterns:

- **`testify`** for assertions (`require` for fatal checks, `assert` for non-fatal)
- **`mockery`** for generated mocks (`MockExec`, `MockJobCommand`)
- **`afero`** for in-memory filesystem abstraction
- **`bytes.Buffer`** for capturing output
- **Table-driven tests** for multiple input scenarios
- Test files live in `<package>/test/` subdirectories

## Test Architecture

```
backup/
  cmd/test/
    commands_test.go      # CLI command tests (all commands, stubbed exec)
    integration_test.go   # Integration tests with real rsync (build tag: integration)
    root_test.go          # Root command help output
  internal/test/
    check_test.go      # CoverageChecker tests (afero-based)
    config_test.go     # Config loading, validation, Apply
    helper_test.go     # NormalizePath, CreateMainLogger
    job_test.go        # Job.Apply with MockJobCommand
    rsync_test.go      # Command constructors, Run methods, GetVersionInfo
    mock_exec_test.go          # Generated mock for Exec interface
    mock_jobcommand_test.go    # Generated mock for JobCommand interface
```

## Dependency Injection Points

| Dependency        | Interface/Type        | Real                                            | Test                     |
| ----------------- | --------------------- | ----------------------------------------------- | ------------------------ |
| Command execution | `internal.Exec`       | `OsExec`                                        | `MockExec` or `stubExec` |
| Job runner        | `internal.JobCommand` | `ListCommand`, `SyncCommand`, `SimulateCommand` | `MockJobCommand`         |
| Filesystem        | `afero.Fs`            | `afero.NewOsFs()`                               | `afero.NewMemMapFs()`    |
| Output            | `io.Writer`           | `os.Stdout` / `cmd.OutOrStdout()`               | `bytes.Buffer`           |
| Logging           | `*log.Logger`         | File-backed logger                              | `log.New(&buf, "", 0)`   |
| Time              | `time.Time`           | `time.Now()`                                    | Fixed `time.Date(...)`   |

## Command-Level Tests (cmd/test/)

Commands are tested through cobra's `Execute()` with captured stdout:

```go
// Stub for Exec interface — lightweight alternative to MockExec for cmd tests
type stubExec struct {
    output []byte
    err    error
}

func (s *stubExec) Execute(_ string, _ ...string) ([]byte, error) {
    return s.output, s.err
}

// Helper: run a command with full dependency injection
func executeCommandWithDeps(t *testing.T, fs afero.Fs, shell internal.Exec, args ...string) (string, error) {
    t.Helper()
    rootCmd := cmd.BuildRootCommandWithDeps(fs, shell)
    var stdout bytes.Buffer
    rootCmd.SetOut(&stdout)
    rootCmd.SetErr(&bytes.Buffer{})
    rootCmd.SetArgs(args)
    err := rootCmd.Execute()
    return stdout.String(), err
}
```

Usage:

```go
func TestRun_ValidConfig(t *testing.T) {
    cfgPath := writeConfigFile(t, `...yaml...`)
    shell := &stubExec{output: []byte("rsync version 3.2.7 protocol version 31\n")}

    stdout, err := executeCommandWithDeps(t, afero.NewMemMapFs(), shell, "run", "--config", cfgPath)

    require.NoError(t, err)
    assert.Contains(t, stdout, "Job: docs")
    assert.Contains(t, stdout, "Status [docs]: SUCCESS")
}
```

Three builder levels available:

- `BuildRootCommand()` — production defaults (real OS filesystem, real exec)
- `BuildRootCommandWithFs(fs)` — custom filesystem, real exec
- `BuildRootCommandWithDeps(fs, shell)` — full control for testing

## Internal Tests — Mockery Mocks

Generated mocks use the expectation pattern:

```go
func TestConfigApply_VersionInfoSuccess(t *testing.T) {
    mockCmd := NewMockJobCommand(t)
    var output bytes.Buffer
    logger := log.New(&bytes.Buffer{}, "", 0)

    cfg := Config{
        Jobs: []Job{
            {Name: "job1", Source: "/src/", Target: "/dst/", Enabled: true},
        },
    }

    mockCmd.EXPECT().GetVersionInfo().Return("rsync version 3.2.3", "/usr/bin/rsync", nil).Once()
    mockCmd.EXPECT().Run(mock.AnythingOfType("internal.Job")).Return(Success).Once()

    err := cfg.Apply(mockCmd, logger, &output)
    require.NoError(t, err)
}
```

## CoverageChecker Tests (afero)

The `CoverageChecker` uses `afero.Fs` so tests never hit the real filesystem:

```go
func newTestChecker() (*CoverageChecker, *bytes.Buffer) {
    var buf bytes.Buffer
    checker := &CoverageChecker{
        Logger: log.New(&buf, "", 0),
        Fs:     afero.NewMemMapFs(),
    }
    return checker, &buf
}
```

## Deterministic Time

`CreateMainLogger` accepts `time.Time` for predictable log paths:

```go
func fixedTime() time.Time {
    return time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
}

func TestCreateMainLogger_DeterministicLogPath(t *testing.T) {
    _, logPath, cleanup, err := CreateMainLogger("backup.yaml", true, fixedTime())
    require.NoError(t, err)
    defer cleanup()
    assert.Equal(t, "logs/sync-2025-06-15T14-30-45-backup-sim", logPath)
}
```

## Integration Tests

Integration tests live in `cmd/test/integration_test.go` behind the `//go:build integration` tag. They exercise the full CLI with **real rsync** against temp directories — no mocks or stubs.

### Build Tag

```go
//go:build integration
```

Tests are excluded from `make test` and `make check-coverage`. Run them separately:

```sh
make test-integration   # go test -race -tags=integration ./... -v
```

### Design Principles

- **Real rsync** — uses `/usr/bin/rsync` via `BuildRootCommand()` (production defaults)
- **Real filesystem** — creates temp directories via `t.TempDir()`, cleaned up automatically
- **Reproducible** — each test sets up its own isolated source/target directory pair
- **No mocks** — validates actual rsync behavior (file transfer, deletion, exclusions)

### Scenarios Covered

| Category             | Tests                                                                            | What's Verified                                                                 |
| -------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| **run — basic**      | `BasicSync`, `IdempotentSync`, `PartialChanges`, `EmptySource`, `DeepHierarchy`  | Files are synced correctly; re-sync is idempotent; only modified files transfer |
| **run — delete**     | `DeleteRemovesExtraFiles`, `NoDeletePreservesExtraFiles`                         | `--delete` flag removes stale files; omitting it preserves them                 |
| **run — exclusions** | `Exclusions`                                                                     | `--exclude` patterns prevent syncing of matching paths                          |
| **run — jobs**       | `DisabledJobSkipped`, `MultipleJobs`, `MixedJobsSummary`, `VariableSubstitution` | Multi-job orchestration, enabled/disabled, `${var}` resolution                  |
| **simulate**         | `NoChanges`, `ShowsChanges`, `SimulateThenRun`                                   | Dry-run produces no side effects; subsequent run works normally                 |
| **list**             | `ShowsCommands`                                                                  | Prints rsync commands without executing them                                    |
| **check-coverage**   | `FullCoverage`, `IncompleteCoverage`                                             | Coverage checker on real directory trees                                        |
| **config**           | `ConfigShow`, `ConfigValidate_Valid`, `ConfigValidate_OverlappingSources`        | End-to-end config parsing and validation                                        |
| **version**          | `Version`                                                                        | Real rsync version output                                                       |

### Example

```go
func TestIntegration_Run_BasicSync(t *testing.T) {
    src, dst := setupDirs(t)
    writeFile(t, filepath.Join(src, "hello.txt"), "hello world")

    cfgPath := writeIntegrationConfig(t, `
sources:
  - path: "`+src+`"
targets:
  - path: "`+dst+`"
jobs:
  - name: "basic"
    source: "`+src+`/"
    target: "`+dst+`/"
    delete: false
`)

    stdout, err := executeIntegrationCommand(t, "run", "--config", cfgPath)
    require.NoError(t, err)
    assert.Contains(t, stdout, "Status [basic]: SUCCESS")
    assert.Equal(t, "hello world", readFileContent(t, filepath.Join(dst, "hello.txt")))
}
```

## Running Tests

```sh
make test               # go test -race ./... -v (unit tests only)
make test-integration   # go test -race -tags=integration ./... -v (includes integration)
make check-coverage     # Fail if coverage < threshold (unit tests only)
make report-coverage    # Generate HTML coverage report
```

## Key Principles

1. **Inject, don't hardcode** — all external dependencies go through interfaces
2. **Never hit the real filesystem** in unit tests — use `afero.NewMemMapFs()`
3. **Use `require` for errors, `assert` for values** — `require` stops the test on failure
4. **Table-driven tests** for multiple input/output scenarios
5. **Scope mocks to individual tests** — each test creates its own mock instance
6. **Defer cleanup** — `CreateMainLogger` returns a cleanup function; always `defer` it
