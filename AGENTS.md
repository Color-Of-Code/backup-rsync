# Project Guidelines

## Overview

CLI tool for managing local backups using `rsync` as the engine. Built in Go with `cobra` for CLI, `afero` for filesystem abstraction, and YAML for configuration. Local-only — no remote rsync support.

Design principles: simple human-readable configuration, minimal interdependencies, flexible variable/macro substitution (`${var}`). Each backup job can be listed, simulated (dry-run), or run independently.

## Code Style

- Idiomatic Go (see `go.mod` for version); format with `go fmt`, lint with `golangci-lint` (`.golangci.yml`)
- All linters enabled by default — check `.golangci.yml` for disabled ones
- Packages: `cmd/` for CLI wiring, `internal/` for core logic
- Prefer dependency injection over global state; use interfaces at consumption boundaries
- Route all output through `cmd.OutOrStdout()` or injected `io.Writer` — never raw `fmt.Printf`
- Dual logging: user-facing output via `io.Writer`, file logging via `*log.Logger` under `logs/`
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

### Key Types, Interfaces & Dependency Injection

| Type              | Kind      | Description                                                                                                    | Injected Dependencies      |
| ----------------- | --------- | -------------------------------------------------------------------------------------------------------------- | -------------------------- |
| `Exec`            | interface | Command execution abstraction (`OsExec` real, `MockExec` tests)                                                | —                          |
| `JobCommand`      | interface | `Run(job Job) JobStatus` + `GetVersionInfo()` — implemented by `ListCommand`, `SimulateCommand`, `SyncCommand` | `Exec` via constructors    |
| `SharedCommand`   | struct    | Base for all commands — holds `BinPath`, `BaseLogPath`, `Shell Exec`, `Output io.Writer`                       | `io.Writer`, `Exec`        |
| `CoverageChecker` | struct    | Analyzes path coverage                                                                                         | `*log.Logger`, `afero.Fs`  |
| `Config`          | struct    | YAML config (`Config`, `Mapping`, `Job`, `Path`, `${var}` substitution); custom `UnmarshalYAML` for job defaults | `*log.Logger` in `Apply()` |

Additional injection points: `afero.Fs` into `BuildRootCommandWithFs()` → `buildCheckCoverageCommand(fs)`; commands use `cmd.OutOrStdout()` for testable output.

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

- See `docs/testing-guide.md` for patterns; `docs/mockery-integration.md` for mock setup
- **Mocks**: Generated with [mockery](https://github.com/vektra/mockery) (config: `.mockery.yml`)
  - Files: `internal/test/mock_<interface>_test.go`; structs: `Mock<Interface>`
- Assertions via `testify` (`require` / `assert`); prefer table-driven tests
- Test files live in `<package>/test/` subdirectories
- Use `afero.NewMemMapFs()` — never hit the real filesystem
- Use `bytes.Buffer` or `io.Discard` for output capture
- Integration tests: `//go:build integration` tag, real rsync on temp directories

## CI Pipeline

Runs on every push/PR to `main` (`.github/workflows/go.yml`):

1. Sanity check (format + clean + mod tidy)
2. Lint (`golangci-lint`)
3. Build
4. Test (with `-race` flag)
5. Integration test (real rsync, `-tags=integration`)
6. Coverage threshold enforcement (98%)
