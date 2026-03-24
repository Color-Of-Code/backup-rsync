package internal

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrUnresolvedMacro indicates a macro could not be resolved.
var ErrUnresolvedMacro = errors.New("unresolved macro")

// MacroFunc transforms an input string and returns the result.
type MacroFunc func(string) string

// GetMacroFunc returns the macro function for the given name, or false if not found.
func GetMacroFunc(name string) (MacroFunc, bool) {
	registry := map[string]MacroFunc{
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"title":      toTitleCase,
		"capitalize": capitalize,
		"camelcase":  toCamelCase,
		"pascalcase": toPascalCase,
		"snakecase":  toSnakeCase,
		"kebabcase":  toKebabCase,
		"trim":       strings.TrimSpace,
	}

	fn, ok := registry[name]

	return fn, ok
}

func toTitleCase(input string) string {
	runes := []rune(input)
	capitalizeNext := true

	for i, r := range runes {
		if unicode.IsSpace(r) || r == '_' || r == '-' {
			capitalizeNext = true
		} else if capitalizeNext {
			runes[i] = unicode.ToUpper(r)
			capitalizeNext = false
		}
	}

	return string(runes)
}

func capitalize(input string) string {
	if input == "" {
		return ""
	}

	runes := []rune(input)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// splitWords splits a string into words by recognizing boundaries at
// underscores, hyphens, spaces, and camelCase transitions.
func isSeparator(r rune) bool {
	return r == '_' || r == '-' || unicode.IsSpace(r)
}

func isCamelBoundary(prev, current rune) bool {
	return unicode.IsUpper(current) && unicode.IsLower(prev)
}

func splitWords(input string) []string {
	var words []string

	var current []rune

	runes := []rune(input)

	for idx, char := range runes {
		if isSeparator(char) {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}

			continue
		}

		if idx > 0 && len(current) > 0 && isCamelBoundary(runes[idx-1], char) {
			words = append(words, string(current))
			current = nil
		}

		current = append(current, char)
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}

func toCamelCase(input string) string {
	words := splitWords(input)

	for i, word := range words {
		lower := strings.ToLower(word)
		if i == 0 {
			words[i] = lower
		} else {
			words[i] = capitalize(lower)
		}
	}

	return strings.Join(words, "")
}

func toPascalCase(input string) string {
	words := splitWords(input)

	for i, word := range words {
		words[i] = capitalize(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

func toSnakeCase(input string) string {
	words := splitWords(input)

	for i, word := range words {
		words[i] = strings.ToLower(word)
	}

	return strings.Join(words, "_")
}

func toKebabCase(input string) string {
	words := splitWords(input)

	for i, word := range words {
		words[i] = strings.ToLower(word)
	}

	return strings.Join(words, "-")
}

const macroPrefix = "@{"
const macroSuffix = "}"

// ResolveMacros evaluates all @{function:argument} expressions in the input string.
// Macros are resolved from the innermost outward to support nesting.
func ResolveMacros(input string) (string, error) {
	// Iteratively resolve innermost macros until none remain.
	for {
		start, end, funcName, arg, found := findInnermostMacro(input)
		if !found {
			break
		}

		fn, ok := GetMacroFunc(funcName)
		if !ok {
			return "", fmt.Errorf("%w: unknown function %q in @{%s:%s}", ErrUnresolvedMacro, funcName, funcName, arg)
		}

		result := fn(arg)
		input = input[:start] + result + input[end:]
	}

	return input, nil
}

// findInnermostMacro finds the innermost (deepest nested) @{func:arg} in the string.
// Returns the start/end indices of the full macro expression, the function name,
// the argument, and whether a macro was found.
func findInnermostMacro(input string) (int, int, string, string, bool) {
	// Find the last occurrence of "@{" before any "}" — that's the innermost.
	lastStart := -1

	for i := range len(input) - 1 {
		if input[i] == '@' && input[i+1] == '{' {
			lastStart = i
		}
	}

	if lastStart < 0 {
		return 0, 0, "", "", false
	}

	// Find the matching closing brace.
	closeIdx := strings.Index(input[lastStart:], macroSuffix)
	if closeIdx < 0 {
		return 0, 0, "", "", false
	}

	closeIdx += lastStart

	inner := input[lastStart+len(macroPrefix) : closeIdx]

	funcName, arg, found := strings.Cut(inner, ":")
	if !found {
		return 0, 0, "", "", false
	}

	return lastStart, closeIdx + len(macroSuffix), funcName, arg, true
}

// ValidateNoUnresolvedMacros checks that no @{...} patterns remain in config fields.
func ValidateNoUnresolvedMacros(cfg Config) error {
	var errs []error

	for _, mapping := range cfg.Mappings {
		for _, field := range []struct {
			name, value string
		}{
			{"mapping source", mapping.Source},
			{"mapping target", mapping.Target},
			{"mapping name", mapping.Name},
		} {
			if strings.Contains(field.value, macroPrefix) {
				errs = append(errs, fmt.Errorf(
					"%w in mapping %q field %q: %s", ErrUnresolvedMacro, mapping.Name, field.name, field.value))
			}
		}

		for _, job := range mapping.Jobs {
			for _, field := range []struct {
				name, value string
			}{
				{"source", job.Source},
				{"target", job.Target},
				{"name", job.Name},
			} {
				if strings.Contains(field.value, macroPrefix) {
					errs = append(errs, fmt.Errorf(
						"%w in job %q field %q: %s", ErrUnresolvedMacro, job.Name, field.name, field.value))
				}
			}
		}
	}

	return errors.Join(errs...)
}
