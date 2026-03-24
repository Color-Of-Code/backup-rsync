# backup-rsync

Backup using `rsync` as an engine.

NOTE: Using rsync in remote mode is not a use case considered for this tool.
Both the source and destination are local mounted drives, ensuring efficient and direct data transfer.

Go tool used for my own private purposes.

**Use at your own risk!**

## Features

- The tool checks that all specified source paths are covered, ensuring completeness of backups.
- Each data copy job is defined and documented in the configuration file.
- Individual jobs can be executed directly from the command line.
- All backup operations are extensively logged, including detailed rsync output and job summaries.
- A dry run mode is available to preview actions without making changes.

## Quick Start

```sh
make build              # Build to dist/backup
./dist/backup --help    # Show available commands
```

## Documentation

- [Configuration File Format](docs/configuration.md) — YAML structure, job definitions, variables, and examples
- [rsync Options and Logging](docs/rsync.md) — rsync flags, itemize-changes output, and log file layout
- [Testing Guide](docs/testing-guide.md) — testing patterns, dependency injection, mocks, and integration tests
- [Mockery Integration](docs/mockery-integration.md) — mock generation setup and usage examples
- [Contributing](CONTRIBUTING.md) — how to set up, develop, and submit changes
