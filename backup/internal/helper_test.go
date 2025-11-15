package internal_test

import (
	"backup-rsync/backup/internal"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to//normalize/", "/path/to/normalize"},
		{"/path//with//double/slashes/", "/path/with/double/slashes"},
		{"/trailing/slash/", "/trailing/slash"},
		{"/no/trailing/slash", "/no/trailing/slash"},
	}

	for _, test := range tests {
		result := internal.NormalizePath(test.input)
		if result != test.expected {
			t.Errorf("NormalizePath(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}
