# Configuration Templating

## Overview

Configuration files support a variable substitution system that enables a single
template to serve multiple use cases. Variables defined in the YAML `variables`
section can be overridden from the command line using the `--set` flag, turning
any config file into a reusable template.

A second mechanism, **includes**, allows a main config to instantiate a template
multiple times with different variable bindings — similar to GitHub Actions'
`uses`/`with` pattern. This replaces per-user (or per-host) config files with a
single template plus a lightweight orchestration file.

## Variable Substitution

Variables are referenced with `${variable_name}` syntax and can appear in:

- **Mapping fields**: `name`, `source`, `target`
- **Job fields**: `name`, `source`, `target`
- **Other variables** (variable-to-variable references)

### Variable Resolution Order

1. Variables defined in the YAML `variables` section are loaded
2. CLI `--set` overrides are merged in (overwriting any matching keys)
3. Variable self-references are resolved (multi-pass, up to 10 iterations)
4. All mapping and job fields are substituted using the fully resolved variables
5. Job relative paths are joined with their mapping's base paths to produce absolute paths

This means variables can reference other variables:

```yaml
variables:
  user: alice

mappings:
  - name: "home"
    source: "/home/${user}"
    target: "/mnt/backup1/${user}"
    jobs:
      - name: "${user}_documents"
        source: "Documents"
        target: "documents"
```

When invoked with `--set user=alice`, the resolution chain is:

1. `user` = `alice` (from CLI)
2. Mapping source = `/home/${user}` → `/home/alice`
3. Mapping target = `/mnt/backup1/${user}` → `/mnt/backup1/alice`
4. Job name = `${user}_documents` → `alice_documents`
5. Job source = `Documents` joined with `/home/alice` → `/home/alice/Documents/`

## Declaring Required Variables (`template:`)

A config file can declare which variables it **requires** using a `template:`
section. When present, the tool validates that every listed variable has a value
before resolving the config — either from the YAML `variables:` section, a
`--set` flag, or an `include:` `with:` block.

```yaml
template:
  variables:
    - user
    - user_cap
```

If any declared variable is missing, the tool exits with an error listing the
unset variables. This makes template requirements explicit and catches typos or
forgotten flags early.

## Includes (`include:`)

A main config can instantiate one or more templates using the `include:` section.
Each entry specifies:

- **`uses`**: path to the template config file (relative to the main config's
  directory, or absolute)
- **`with`**: map of variable values to inject into the template

```yaml
include:
  - uses: user_template.yaml
    with:
      user: alice
      user_cap: Alice

  - uses: user_template.yaml
    with:
      user: bob
      user_cap: Bob
```

### How includes work

1. Each `include` entry loads the referenced template file
2. The `with` values are merged into the template's `variables` map
3. Template variable validation runs (all `template.variables` must be set)
4. The template is resolved (variable substitution)
5. The resolved mappings (with their jobs) are appended to the main config
6. After all includes are expanded, the main config goes through standard
   validation (job names, paths, overlaps)

### Constraints

- **No nested includes**: a template referenced via `include` cannot itself
  contain `include` entries. This keeps the system simple and predictable.
- **Include paths are relative** to the directory containing the main config
  file, unless an absolute path is specified.
- A main config can have both its own `sources`/`targets`/`jobs` and `include`
  entries — they are merged together.

### Example: multi-user orchestration

**Template** (`user_template.yaml`):

```yaml
template:
  variables:
    - user
    - user_cap

sources:
  - path: "/home/${user}/"
  - path: "/home/data/family/${user_cap}/"

targets:
  - path: "/mnt/backup1/${user}"

variables:
  source_home: "/home/${user}"
  target_base: "/mnt/backup1/${user}"

jobs:
  - name: "${user}_mail"
    source: "${source_home}/.thunderbird/"
    target: "${target_base}/mail"
  - name: "${user}_documents"
    source: "${source_home}/Documents/"
    target: "${target_base}/documents"
```

**Main config** (`users.yaml`):

```yaml
include:
  - uses: user_template.yaml
    with:
      user: alice
      user_cap: Alice

  - uses: user_template.yaml
    with:
      user: bob
      user_cap: Bob
```

Running `backup run --config users.yaml` expands both includes and executes all
jobs for both users in a single invocation.

## CLI Usage

The `--set` flag can be used with any command (`list`, `run`, `simulate`,
`config show`, `config validate`, `check-coverage`):

```sh
# Show resolved config for user "alice"
backup config show --config user_template.yaml --set user=alice --set user_cap=Alice

# Simulate backup for user "bob"
backup simulate --config user_template.yaml --set user=bob --set user_cap=Bob

# Run backup for user "alice"
backup run --config user_template.yaml --set user=alice --set user_cap=Alice

# Run all users via the orchestration config (no --set needed)
backup run --config users.yaml
```

Multiple `--set` flags can be specified. Later values override earlier ones for
the same key.

## Choosing Between `--set` and `include:`

| Approach             | Best for                                    | Example                                      |
| -------------------- | ------------------------------------------- | -------------------------------------------- |
| `--set` flags        | Ad-hoc CLI usage, CI scripts, single user   | `backup run --config tpl.yaml --set user=bob` |
| `include:` in config | Multi-user/multi-host, declarative setups   | `backup run --config users.yaml`              |
| Combined             | Includes with a shared override             | `backup run --config users.yaml --set base=/mnt/nfs` |

Both approaches can be combined: `--set` overrides apply to the main config's
variables before includes are expanded. Variables defined inside a template's
`with:` block are scoped to that template instance only.

## Design Notes

- **Backward compatible**: Configs without `${…}` placeholders work unchanged.
  The `--set` flag and `template:`/`include:` sections are all optional.
- **Override semantics**: `--set` values take precedence over values defined in
  the YAML `variables` section.
- **Multi-pass resolution**: Variable self-references are resolved iteratively
  (up to 10 passes). Circular references are left unresolved rather than
  causing an error.
- **Validation**: Template variable validation runs before resolution to catch
  missing variables early. Job name validation (uniqueness, character checks)
  and path validation run on fully resolved configs.
- **No nested includes**: Keeping the include depth to one level avoids
  complexity and makes configs easy to reason about.
