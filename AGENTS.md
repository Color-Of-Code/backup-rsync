# Project Guidelines

## Overview

CLI tool for managing local backups using `rsync` as the engine.
Built in Go with `cobra` for CLI, `afero` for filesystem abstraction, and YAML for configuration. Local-only — no remote rsync support.

## Code Style

- Follow idiomatic Go conventions (see `go.mod` for the required Go version)
- Format with `go fmt`; lint with `golangci-lint` (config in `.golangci.yml`)
- All linters enabled by default — check `.golangci.yml` for disabled ones
- Keep packages focused: `cmd/` for CLI wiring, `internal/` for core logic
- Prefer dependency injection over global state for testability
- Use interfaces at consumption boundaries (see `internal/exec.go`, `internal/job_command.go`)
- All output in commands routed through `cmd.OutOrStdout()` or injected `io.Writer` — never raw `fmt.Printf`
- All commands use `RunE` with wrapped errors

## Architecture

```
backup/
  main.go                # Entrypoint — calls cmd.BuildRootCommand().Execute()
  cmd/                   # Cobra commands: list, run, simulate, config (show/validate), check-coverage, version
    root.go              # BuildRootCommand() / BuildRootCommandWithFs(fs) — injects afero.Fs
    test/                # Cobra command integration tests
  internal/              # Core logic: config, job execution, rsync wrapper, coverage checker
    test/                # Unit tests + mockery-generated mocks
```

### Key Types & Interfaces

- **`Exec`** (interface): Command execution abstraction (`OsExec` for real, `MockExec` for tests)
- **`JobCommand`** (interface): `Run(job Job) JobStatus` + `GetVersionInfo()` — implemented by `ListCommand`, `SimulateCommand`, `SyncCommand`
- **`SharedCommand`** (struct): Base for all commands — holds `BinPath`, `BaseLogPath`, `Shell Exec`, `Output io.Writer`
- **`CoverageChecker`** (struct): Analyzes path coverage with injected `*log.Logger` and `afero.Fs`
- **`Config`**: YAML-based (`Config`, `Job`, `Path`, variables with `${var}` substitution)

### Dependency Injection

- `Exec` injected into all command constructors (`NewListCommand`, `NewSimulateCommand`, `NewSyncCommand`)
- `io.Writer` injected into `SharedCommand` for output capture
- `afero.Fs` injected into `BuildRootCommandWithFs()` → `buildCheckCoverageCommand(fs)`
- `*log.Logger` injected into `CoverageChecker` and `Config.Apply()`
- Commands use `cmd.OutOrStdout()` for testable output

## Build and Test

```sh
make build              # Build to dist/backup
make test               # go test -race ./... -v
make test-integration   # go test -race -tags=integration ./... -v
make lint               # golangci-lint run ./...
make lint-fix           # Auto-fix lint issues
make format             # go fmt ./...
make tidy               # gofmt -s + go mod tidy
make sanity-check       # format + clean + tidy
make check-coverage     # Fail if coverage < 98%
make report-coverage    # Generate HTML coverage report
```

## Testing Conventions

- See `docs/testing-guide.md` for patterns and examples
- Use **dependency injection** — inject interfaces, not concrete types
- **Mocks**: Generated with [mockery](https://github.com/vektra/mockery) (config: `.mockery.yml`)
  - Mock files live in `internal/test/` as `mock_<interface>_test.go`
  - Mock structs named `Mock<Interface>` (e.g., `MockExec`, `MockJobCommand`)
  - See `docs/mockery-integration.md` for setup details
- Use `testify` for assertions (`require` / `assert`)
- Test files live in `<package>/test/` subdirectories
- Prefer table-driven tests for multiple input scenarios
- Use `afero.NewMemMapFs()` in tests — never hit the real filesystem
- Use `bytes.Buffer` or `io.Discard` for output capture in tests
- Integration tests use `//go:build integration` tag and run real rsync on temp directories
- CI enforces coverage threshold via `make check-coverage`

## CI Pipeline

CI runs on every push/PR to `main` (`.github/workflows/go.yml`):

1. Sanity check (format + clean + mod tidy)
2. Lint (golangci-lint)
3. Build
4. Test (with `-race` flag)
5. Integration test (with real rsync, `-tags=integration`)
6. Coverage threshold enforcement (98%)

## Conventions

- No remote rsync — only locally mounted paths
- Job-level granularity: each backup job can be listed, simulated, or run independently
- Dry-run/simulate mode available for all operations
- Logging goes to both an injected `io.Writer` (user output) and `*log.Logger` (file logging) under `logs/`
- Custom YAML unmarshaling handles job defaults (see `internal/job.go`)
- CI runs sanity checks, lint, and build on every push/PR (`.github/workflows/go.yml`)
