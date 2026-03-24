# Configuration File Format (`sync.yaml`)

The backup tool is configured using a YAML file, typically named `sync.yaml`. This file defines the sources, targets, variables, and backup jobs. Below is a description of the structure, settings, and an example configuration.

## Top-Level Structure

```yaml
sources: # List of source paths to back up
targets: # List of target paths for backups
variables: # Key-value pairs for variable substitution
jobs: # List of backup jobs
```

## Sources and Targets

Each source and target is defined as a `Path` object:

```yaml
- path: "/path/to/source/or/target/"
  exclusions:
    - "/path/to/exclude/"
```

- `path`: The directory path to include as a source or target.
- `exclusions` (optional): List of subpaths to exclude from backup.

## Variables

Variables are key-value pairs that can be referenced in job definitions using `${varname}` syntax.

```yaml
variables:
  target_base: "/mnt/backup1"
```

## Macros

Macros apply string transformation functions to values using `@{function:argument}` syntax. Variables are resolved before macros, so they compose naturally. See [macros.md](macros.md) for the full list of available functions and detailed usage.

```yaml
variables:
  user: alice

jobs:
  - name: "${user}_docs"
    target: "/backup/@{capitalize:${user}}/docs"  # resolves to /backup/Alice/docs
```

## Jobs

Each job defines a backup operation:

```yaml
- name: "job_name" # Unique name for the job
  source: "/path/to/source/" # Source directory
  target: "/path/to/target/" # Target directory (can use variables)
  delete: true # (Optional) Delete files in target not in source (default: true)
  enabled: true # (Optional) Enable/disable the job (default: true)
  exclusions: # (Optional) List of subpaths to exclude
    - "/subpath/to/exclude/"
```

### Job Fields

- `name`: Unique identifier for the job.
- `source`: Path to the source directory.
- `target`: Path to the target directory. Variables can be used (e.g., `${target_base}/user/home`).
- `delete`: (Optional) If `true`, files deleted from the source are also deleted from the target. Defaults to `true` if omitted.
- `enabled`: (Optional) If `false`, the job is skipped. Defaults to `true` if omitted.
- `exclusions`: (Optional) List of subpaths to exclude from this job.

## Example Configuration

```yaml
sources:
  - path: "/home/user/"
    exclusions:
      - "/Downloads/"
      - "/.cache/"
targets:
  - path: "/mnt/backup1/"
variables:
  target_base: "/mnt/backup1"
jobs:
  - name: "user_home"
    source: "/home/user/"
    target: "${target_base}/user/home"
    exclusions:
      - "/Downloads/"
      - "/.cache/"
    delete: true
    enabled: true
  - name: "user_documents"
    source: "/home/user/Documents/"
    target: "${target_base}/user/documents"
```

## Notes

- All paths should be absolute.
- Exclusions are relative to the specified source or target path.
- Jobs with `enabled: false` are ignored.
- If `delete` is omitted, it defaults to `true` (target files not present in source will be deleted from the destination).
