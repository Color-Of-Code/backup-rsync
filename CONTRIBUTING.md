# Contributing

Thanks for your interest in contributing to backup-rsync!

## Getting Started

1. Fork the repo and clone it
2. Install Go (see `go.mod` for minimum version) and `rsync`
3. Run `make build` to verify the setup

## Development Workflow

```sh
make format            # Format code
make lint              # Run linter
make test              # Run unit tests
make test-integration  # Run integration tests (requires rsync)
make check-coverage    # Verify coverage threshold (98%)
```

## Submitting Changes

1. Create a branch from `main`
2. Make your changes — keep commits focused
3. Ensure `make lint` and `make test` pass
4. Open a pull request against `main`

## Guidelines

- Follow idiomatic Go conventions
- Add tests for new functionality
- Use dependency injection over global state
- Keep the CI green — all checks must pass before merge

## License

By contributing, you agree that your contributions will be licensed under the [GPL-3.0 License](LICENSE).
