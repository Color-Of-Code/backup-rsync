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

## Configuration File Format (`sync.yaml`)

The backup tool is configured using a YAML file, typically named `config.yaml`. This file defines the sources, targets, variables, and backup jobs. Below is a description of the structure, settings, and an example configuration.

### Top-Level Structure

```yaml
sources: # List of source paths to back up
targets: # List of target paths for backups
variables: # Key-value pairs for variable substitution
jobs: # List of backup jobs
```

### Sources and Targets

Each source and target is defined as a `Path` object:

```yaml
- path: "/path/to/source/or/target/"
  exclusions:
    - "/path/to/exclude/"
```

- `path`: The directory path to include as a source or target.
- `exclusions` (optional): List of subpaths to exclude from backup.

### Variables

Variables are key-value pairs that can be referenced in job definitions using `${varname}` syntax.

```yaml
variables:
  target_base: "/mnt/backup1"
```

### Jobs

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

#### Job Fields

- `name`: Unique identifier for the job.
- `source`: Path to the source directory.
- `target`: Path to the target directory. Variables can be used (e.g., `${target_base}/user/home`).
- `delete`: (Optional) If `true`, files deleted from the source are also deleted from the target. Defaults to `true` if omitted.
- `enabled`: (Optional) If `false`, the job is skipped. Defaults to `true` if omitted.
- `exclusions`: (Optional) List of subpaths to exclude from this job.

### Example Configuration

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

### Notes

- All paths should be absolute.
- Exclusions are relative to the specified source or target path.
- Jobs with `enabled: false` are ignored.
- If `delete` is omitted, it defaults to `true` (target files not present in source will be deleted from the destination).

## rsync and Logging

This tool uses `rsync` with the following key options:

- `-a` : Archive mode (preserves permissions, times, symbolic links, etc.)
- `-i` : Itemize changes (shows a change summary for each file)
- `-v` : Verbose output
- `--stats` : Print a detailed set of statistics on the file transfer
- `--delete` : Delete extraneous files from the destination dirs (if enabled in the job)
- `--exclude=PATTERN` : Exclude files matching PATTERN (from job or source/target exclusions)
- `--log-file=FILE` : Write rsync output to the specified log file
- `--dry-run` : Show what would be done, but make no changes (for simulation/dry-run mode)

### Understanding the `-i` (itemize changes) Output

The `-i` flag produces a change summary for each file, with a string of characters indicating what changed. For example:

```
>f.st...... somefile.txt
cd+++++++++ newdir/
```

The first character indicates the file type and action:

- `>` : File sent to the receiver
- `<` : File received from the sender
- `c` : Local change/creation of a directory

The next characters indicate what changed:

- `f` : File
- `d` : Directory
- `L` : Symlink
- `D` : Device
- `S` : Special file

The remaining characters show what changed (see `man rsync` for full details):

- `s` : Size
- `t` : Modification time
- `p` : Permissions
- `o` : Owner
- `g` : Group
- `a` : ACL
- `x` : Extended attributes
- `+` : Creation (for directories)

### Logging

Each job writes its rsync output to a dedicated log file, typically named `job-<jobname>.log` in a timestamped log directory (e.g., `logs/sync-YYYY-MM-DDTHH-MM-SS/`).

The log files contain the full rsync output, including the itemized changes and statistics. A `summary.log` file records the status (SUCCESS, FAILURE, SKIPPED) for each job in the run.

**Empty log files** may indicate that no changes were made during the backup for that job.

You can review these logs to audit what was copied, changed, or deleted during each backup run.
