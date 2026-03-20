# Project Guidelines

## Overview

CLI tool for managing local backups using `rsync` as the engine. Built in Go with `cobra` for CLI, `afero` for filesystem abstraction, and YAML for configuration. Local-only — no remote rsync support.

## Code Style

- Go 1.24+; follow idiomatic Go conventions
- Format with `go fmt`; lint with `golangci-lint` (config in `.golangci.yml`)
- All linters enabled by default — check `.golangci.yml` for disabled ones
- Keep packages focused: `cmd/` for CLI wiring, `internal/` for core logic
- Prefer dependency injection over global state for testability
- Use interfaces at consumption boundaries (see `internal/exec.go`, `internal/job_command.go`)

## Architecture

```
backup/
  main.go              # Entrypoint — calls cmd.BuildRootCommand().Execute()
  cmd/                  # Cobra commands: list, run, simulate, config, check-coverage, version
  internal/             # Core logic: config loading, job execution, rsync wrapper
```

- **Config**: YAML-based (`Config`, `Job`, sources, targets, variables with `${var}` substitution)
- **Jobs**: Each job maps a source to a target with optional exclusions; jobs run independently
- **Rsync**: Wrapped via `SharedCommand` struct in `internal/rsync.go`
- YAML config files at repo root (e.g., `sync.yaml`) define backup configurations

## Build and Test

```sh
make build          # Build to dist/backup
make test           # go test ./... -v
make lint           # golangci-lint run ./...
make lint-fix       # Auto-fix lint issues
make format         # go fmt ./...
make tidy           # gofmt -s + go mod tidy
make sanity-check   # format + clean + tidy
```

## Testing Conventions

- See `TESTING_GUIDE.md` for patterns and examples
- Use **dependency injection** — inject interfaces, not concrete types
- **Mocks**: Generated with [mockery](https://github.com/vektra/mockery) (config: `.mockery.yml`)
  - Mock files live in `internal/test/` as `mock_<interface>_test.go`
  - Mock structs named `Mock<Interface>`
  - See `MOCKERY_INTEGRATION.md` for setup details
- Use `testify` for assertions (`require` / `assert`)
- Test files live in `<package>/test/` subdirectories
- Prefer table-driven tests for multiple input scenarios

## Conventions

- No remote rsync — only locally mounted paths
- Job-level granularity: each backup job can be listed, simulated, or run independently
- Dry-run/simulate mode available for all operations
- Logging goes to both stdout and timestamped log directories under `logs/`
- Custom YAML unmarshaling handles job defaults (see `internal/job.go`)
- CI runs sanity checks, lint, and build on every push/PR (`.github/workflows/go.yml`)
