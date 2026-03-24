# Testing Guide

## Approach

- **Dependency injection** over global state — every external dependency is injected via an interface or value
- **`testify`** for assertions — `require` (fatal) and `assert` (non-fatal)
- **`mockery`** for generated interface mocks
- **`afero`** for in-memory filesystem abstraction in unit tests
- **Table-driven tests** — canonical Go pattern; use whenever 2+ cases share the same structure
- **Declarative test data builders** in a shared `testutil` package — avoid inline YAML and repeated struct literals
- Test files live in `<package>/test/` subdirectories

## Dependency Injection

All external dependencies are abstracted behind interfaces or injected types. In tests, swap real implementations for mocks, stubs, or in-memory alternatives:

| What              | Abstraction      | In Tests                                       |
| ----------------- | ---------------- | ---------------------------------------------- |
| Command execution | Interface        | Generated mock or lightweight stub             |
| Job runner        | Interface        | Generated mock                                 |
| Filesystem        | `afero.Fs`       | `afero.NewMemMapFs()`                          |
| Output            | `io.Writer`      | `bytes.Buffer` or `io.Discard`                 |
| Logging           | `*log.Logger`    | Logger writing to a `bytes.Buffer`             |
| Time              | `time.Time`      | Fixed value via `time.Date(...)`               |

See `internal/exec.go` and `internal/job_command.go` for the interface definitions. See `cmd/root.go` for the builder functions that wire dependencies.

## Test Data Builders

A shared `internal/testutil/` package provides declarative helpers to reduce boilerplate:

- **Config builder** — fluent API to generate YAML config strings without raw string literals
- **Config file writer** — writes content to a temp file and returns the path
- **Job builder** — creates a `Job` struct with sensible defaults; override individual fields via functional options

See `internal/testutil/*.go` for the full API and available options.

## Table-Driven Tests

Use table-driven tests whenever multiple cases share the same test structure and differ only in inputs and expectations. Define a slice of test structs, iterate with `t.Run()`.

**When NOT to use**: when cases need fundamentally different mock wiring, different assertion logic, or complex per-case setup. If you'd need a `func(...)` closure field in the table struct, keep tests separate.

Browse the test files for examples — most validation, path-checking, and argument-building tests follow this pattern.

## Command-Level Tests

CLI commands are tested through cobra's `Execute()` with captured stdout/stderr. Helper functions in the test files wrap the root command builder at different injection levels (default deps, custom filesystem, or full control).

A lightweight exec stub (implementing the `Exec` interface inline) is used instead of full mocks for command-level tests where only the output matters.

## Generated Mocks

Generated mocks (via `mockery`) use the `.EXPECT()` pattern for setting expectations. Each test creates its own mock instance — no shared state between tests.

Mock configuration: `.mockery.yml`. See `MOCKERY_INTEGRATION.md` for regeneration instructions.

## Integration Tests

Integration tests are gated behind a build tag (`//go:build integration`). They exercise the full CLI with real rsync against temp directories — no mocks or stubs.

```sh
make test-integration
```

Design principles:
- Real filesystem via `t.TempDir()`, real rsync via production command builder
- Each test sets up its own isolated directory pair
- Config built using the shared `testutil` builder

Scenarios covered: sync (basic, idempotent, partial, empty, deep), delete/preserve, exclusions, disabled/multiple jobs, variable substitution, simulate, list, check-coverage, config show/validate, version.

## Running Tests

```sh
make test               # Unit tests
make test-integration   # Integration tests (requires rsync)
make check-coverage     # Fail if below threshold
make report-coverage    # HTML coverage report
```

## Key Principles

1. **Inject, don't hardcode** — all external dependencies go through interfaces
2. **Never hit the real filesystem** in unit tests — use in-memory filesystem
3. **`require` for errors, `assert` for values** — `require` stops the test on failure
4. **Table-driven tests** for 2+ cases with same structure
5. **Use shared builders** — avoid inline YAML and repeated struct literals
6. **Scope mocks per test** — no shared mock state
7. **Defer cleanup** for resources that return a cleanup function
8. **Keep functions short** — use compact table entries and data-driven fields over closures
