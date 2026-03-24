# rsync and Logging

## rsync Options

This tool uses `rsync` with the following key options:

- `-a` : Archive mode (preserves permissions, times, symbolic links, etc.)
- `-i` : Itemize changes (shows a change summary for each file)
- `-v` : Verbose output
- `--stats` : Print a detailed set of statistics on the file transfer
- `--delete` : Delete extraneous files from the destination dirs (if enabled in the job)
- `--exclude=PATTERN` : Exclude files matching PATTERN (from job or source/target exclusions)
- `--log-file=FILE` : Write rsync output to the specified log file
- `--dry-run` : Show what would be done, but make no changes (for simulation/dry-run mode)

## Understanding the `-i` (itemize changes) Output

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

## Logging

Each job writes its rsync output to a dedicated log file, typically named `job-<jobname>.log` in a timestamped log directory (e.g., `logs/sync-YYYY-MM-DDTHH-MM-SS/`).

The log files contain the full rsync output, including the itemized changes and statistics. A `summary.log` file records the status (SUCCESS, FAILURE, SKIPPED) for each job in the run.

You can review these logs to audit what was copied, changed, or deleted during each backup run.
