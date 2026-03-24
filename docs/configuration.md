# Configuration File Format (`sync.yaml`)

The backup tool is configured using a YAML file, typically named `sync.yaml`. This file defines mappings (source-to-target directory pairs), variables, and backup jobs. Below is a description of the structure, settings, and an example configuration.

## Top-Level Structure

```yaml
template:   # (Optional) Declares required variables for this template
include:    # (Optional) List of template configs to instantiate
variables:  # (Optional) Key-value pairs for variable substitution
mappings:   # List of source-to-target directory mappings, each with its own jobs
```

## Mappings

Each mapping defines a source-to-target directory pair and owns a list of backup jobs. Job paths within a mapping are relative to the mapping's source and target.

```yaml
mappings:
  - name: "home"
    source: "/home/user"
    target: "/mnt/backup1/user"
    exclusions:        # (Optional) Source-level exclusions
      - "/Downloads/"
    jobs:
      - name: "documents"
        source: "Documents"
        target: "documents"
```

- `name`: A label for identifying the mapping.
- `source`: Absolute path to the source directory for this mapping.
- `target`: Absolute path to the target directory for this mapping.
- `exclusions` (optional): List of subpaths to exclude at the source level.
- `jobs`: List of backup jobs (see below).

During resolution, each job's relative source and target paths are joined with the mapping's base paths to produce absolute paths for rsync. For example, a job with `source: "Documents"` under a mapping with `source: "/home/user"` resolves to `/home/user/Documents/`.

## Variables

Variables are key-value pairs that can be referenced in mapping and job fields using `${varname}` syntax.

```yaml
variables:
  user: alice
```

## Macros

Macros apply string transformation functions to values using `@{function:argument}` syntax. Variables are resolved before macros, so they compose naturally. See [macros.md](macros.md) for the full list of available functions and detailed usage.

```yaml
variables:
  user: alice

mappings:
  - name: "home"
    source: "/home/${user}"
    target: "/backup/@{capitalize:${user}}"
    jobs:
      - name: "${user}_docs"
        source: "Documents"
        target: "docs"
```

## Template (Optional)

Declares which variables a config file requires. When present, the tool validates
that every listed variable has a value before resolving. See
[templating.md](templating.md) for details.

```yaml
template:
  variables:
    - user
    - user_cap
```

## Include (Optional)

Instantiate one or more template configs with specific variable bindings. Each
entry references a template file and provides the required variables. See
[templating.md](templating.md) for details.

```yaml
include:
  - uses: user_template.yaml
    with:
      user: alice
      user_cap: Alice
```

## Jobs

Each job defines a backup operation within a mapping. Job paths are relative to the mapping's source and target:

```yaml
- name: "job_name"        # Unique name for the job
  source: "relative/src"  # Relative to mapping source (use "" for root)
  target: "relative/tgt"  # Relative to mapping target (use "" for root)
  delete: true            # (Optional) Delete files in target not in source (default: true)
  enabled: true           # (Optional) Enable/disable the job (default: true)
  exclusions:             # (Optional) List of subpaths to exclude
    - "/subpath/to/exclude/"
```

### Job Fields

- `name`: Unique identifier for the job.
- `source`: Path to the source directory, relative to the mapping's source. Use `""` to sync the entire mapping source.
- `target`: Path to the target directory, relative to the mapping's target. Use `""` to sync to the mapping target root.
- `delete`: (Optional) If `true`, files deleted from the source are also deleted from the target. Defaults to `true` if omitted.
- `enabled`: (Optional) If `false`, the job is skipped. Defaults to `true` if omitted.
- `exclusions`: (Optional) List of subpaths to exclude from this job.

## Example Configuration

```yaml
mappings:
  - name: "home"
    source: "/home/user"
    target: "/mnt/backup1/user"
    exclusions:
      - "/Downloads/"
      - "/.cache/"
    jobs:
      - name: "user_home"
        source: ""
        target: "home"
        exclusions:
          - "/Downloads/"
          - "/.cache/"
      - name: "user_documents"
        source: "Documents"
        target: "documents"
```

## Notes

- Mapping-level source and target paths should be absolute.
- Job-level source and target paths are relative to the mapping and are joined during resolution.
- Exclusions are relative to the specified source path.
- Jobs with `enabled: false` are ignored.
- If `delete` is omitted, it defaults to `true` (target files not present in source will be deleted from the destination).
- For templating features (`template:`, `include:`, `--set` flags), see [templating.md](templating.md).
