# Macros

Macros allow you to apply string transformation functions inside configuration values. They complement [variable substitution](configuration.md) by transforming strings at resolution time.

## Syntax

```
@{function_name:argument}
```

- `@{` opens a macro call.
- `function_name` is the name of the transformation function.
- `:` separates the function name from its argument.
- `argument` is the input string to transform.
- `}` closes the macro call.

## Resolution Order

1. **Variable substitution** — all `${variable}` references are replaced with their values.
2. **Macro evaluation** — all `@{function:argument}` expressions are evaluated.

This means macros can operate on resolved variable values:

```yaml
variables:
  user: alice

jobs:
  - name: "${user}_docs"
    source: "/home/${user}/Documents/"
    target: "/backup/@{capitalize:${user}}/Documents/"
```

After resolution, the target becomes `/backup/Alice/Documents/`.

## Nesting

Macros can be nested. Inner macros are resolved first:

```yaml
target: "/backup/@{upper:@{trim:  ${user}  }}/"
```

Given `user: alice`, this resolves as:
1. Variable substitution: `/backup/@{upper:@{trim:  alice  }}/`
2. Inner macro `@{trim:  alice  }` → `alice`
3. Outer macro `@{upper:alice}` → `ALICE`
4. Result: `/backup/ALICE/`

## Available Functions

| Function | Description | Example Input | Example Output |
|---|---|---|---|
| `upper` | Convert to uppercase | `hello world` | `HELLO WORLD` |
| `lower` | Convert to lowercase | `HELLO WORLD` | `hello world` |
| `title` | Capitalize first letter of each word | `hello world` | `Hello World` |
| `capitalize` | Capitalize first character only | `hello world` | `Hello world` |
| `camelcase` | Convert to camelCase | `hello_world` | `helloWorld` |
| `pascalcase` | Convert to PascalCase | `hello_world` | `HelloWorld` |
| `snakecase` | Convert to snake_case | `helloWorld` | `hello_world` |
| `kebabcase` | Convert to kebab-case | `helloWorld` | `hello-world` |
| `trim` | Remove leading/trailing whitespace | `  hello  ` | `hello` |

### Case conversion details

The `camelcase`, `pascalcase`, `snakecase`, and `kebabcase` functions detect word boundaries at:

- Underscores (`_`)
- Hyphens (`-`)
- Spaces
- camelCase transitions (a lowercase letter followed by an uppercase letter)

Examples:

| Input | `camelcase` | `pascalcase` | `snakecase` | `kebabcase` |
|---|---|---|---|---|
| `hello_world` | `helloWorld` | `HelloWorld` | `hello_world` | `hello-world` |
| `hello-world` | `helloWorld` | `HelloWorld` | `hello_world` | `hello-world` |
| `HelloWorld` | `helloWorld` | `HelloWorld` | `hello_world` | `hello-world` |
| `helloWorld` | `helloWorld` | `HelloWorld` | `hello_world` | `hello-world` |

## Validation

After all variables and macros are resolved, the configuration is validated to ensure no unresolved `@{...}` expressions remain. If any macro could not be resolved (e.g., an unknown function name), the configuration is rejected with an error.

The `config show` command always displays the fully resolved configuration with all variables substituted and all macros evaluated.

## Practical Example

Instead of maintaining separate `user` and `user_cap` variables:

```yaml
# Before: redundant variables
variables:
  user: alice
  user_cap: Alice

jobs:
  - name: "${user}_docs"
    source: "/home/${user}/Documents/"
    target: "/backup/${user_cap}/Documents/"
```

Use a macro to derive the capitalized form:

```yaml
# After: single variable with macro
variables:
  user: alice

jobs:
  - name: "${user}_docs"
    source: "/home/${user}/Documents/"
    target: "/backup/@{capitalize:${user}}/Documents/"
```
