package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "backup-rsync/backup/internal"
	"backup-rsync/backup/internal/testutil"
)

func TestResolveMacros_StringFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// upper
		{"UpperSimple", "@{upper:hello}", "HELLO"},
		{"UpperMixed", "@{upper:Hello World}", "HELLO WORLD"},
		{"UpperEmpty", "@{upper:}", ""},

		// lower
		{"LowerSimple", "@{lower:HELLO}", "hello"},
		{"LowerMixed", "@{lower:Hello World}", "hello world"},

		// title
		{"TitleSimple", "@{title:hello world}", "Hello World"},
		{"TitleUnderscores", "@{title:hello_world}", "Hello_World"},
		{"TitleHyphens", "@{title:hello-world}", "Hello-World"},

		// capitalize
		{"CapitalizeSimple", "@{capitalize:hello}", "Hello"},
		{"CapitalizeOneChar", "@{capitalize:h}", "H"},
		{"CapitalizeEmpty", "@{capitalize:}", ""},
		{"CapitalizeSentence", "@{capitalize:hello world}", "Hello world"},

		// camelcase
		{"CamelFromSnake", "@{camelcase:hello_world}", "helloWorld"},
		{"CamelFromKebab", "@{camelcase:hello-world}", "helloWorld"},
		{"CamelFromSpace", "@{camelcase:hello world}", "helloWorld"},
		{"CamelFromPascal", "@{camelcase:HelloWorld}", "helloWorld"},
		{"CamelSingleWord", "@{camelcase:hello}", "hello"},

		// pascalcase
		{"PascalFromSnake", "@{pascalcase:hello_world}", "HelloWorld"},
		{"PascalFromKebab", "@{pascalcase:hello-world}", "HelloWorld"},
		{"PascalFromSpace", "@{pascalcase:hello world}", "HelloWorld"},
		{"PascalFromCamel", "@{pascalcase:helloWorld}", "HelloWorld"},
		{"PascalSingleWord", "@{pascalcase:hello}", "Hello"},

		// snakecase
		{"SnakeFromCamel", "@{snakecase:helloWorld}", "hello_world"},
		{"SnakeFromPascal", "@{snakecase:HelloWorld}", "hello_world"},
		{"SnakeFromKebab", "@{snakecase:hello-world}", "hello_world"},
		{"SnakeFromSpace", "@{snakecase:hello world}", "hello_world"},
		{"SnakeSingleWord", "@{snakecase:hello}", "hello"},

		// kebabcase
		{"KebabFromCamel", "@{kebabcase:helloWorld}", "hello-world"},
		{"KebabFromPascal", "@{kebabcase:HelloWorld}", "hello-world"},
		{"KebabFromSnake", "@{kebabcase:hello_world}", "hello-world"},
		{"KebabFromSpace", "@{kebabcase:hello world}", "hello-world"},

		// trim
		{"TrimSpaces", "@{trim:  hello  }", "hello"},
		{"TrimTabs", "@{trim:\thello\t}", "hello"},
		{"TrimNone", "@{trim:hello}", "hello"},

		// no macros
		{"NoMacros", "/home/user/docs", "/home/user/docs"},
		{"EmptyString", "", ""},
		{"VariableSyntax", "${var}", "${var}"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ResolveMacros(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestResolveMacros_InContext(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"MacroInPath", "/home/@{lower:USER}/docs", "/home/user/docs"},
		{"MacroAtStart", "@{upper:hello}/world", "HELLO/world"},
		{"MacroAtEnd", "/path/@{capitalize:test}", "/path/Test"},
		{"MultipleMacros", "@{upper:hello}-@{lower:WORLD}", "HELLO-world"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ResolveMacros(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestResolveMacros_Nested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"NestedUpperTrim", "@{upper:@{trim:  hello  }}", "HELLO"},
		{"NestedCapitalizeLower", "@{capitalize:@{lower:HELLO}}", "Hello"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ResolveMacros(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestResolveMacros_UnknownFunction(t *testing.T) {
	_, err := ResolveMacros("@{unknown:hello}")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnresolvedMacro)
	assert.Contains(t, err.Error(), "unknown function")
}

func TestResolveMacros_MissingColon(t *testing.T) {
	// @{nocolon} has no colon separator, so it's not a valid macro — left unchanged.
	result, err := ResolveMacros("@{nocolon}")
	require.NoError(t, err)
	assert.Equal(t, "@{nocolon}", result)
}

func TestResolveConfig_WithMacros(t *testing.T) {
	cfg := Config{
		Variables: map[string]string{
			"user": "jaap",
		},
		Jobs: []Job{
			{
				Name:   "${user}_mail",
				Source: "/home/${user}/",
				Target: "/backup/@{capitalize:${user}}/mail",
			},
		},
	}

	resolved, err := ResolveConfig(cfg)
	require.NoError(t, err)

	assert.Equal(t, "jaap_mail", resolved.Jobs[0].Name)
	assert.Equal(t, "/home/jaap/", resolved.Jobs[0].Source)
	assert.Equal(t, "/backup/Jaap/mail", resolved.Jobs[0].Target)
}

func TestResolveConfig_MacroError(t *testing.T) {
	cfg := Config{
		Jobs: []Job{
			{
				Name:   "job1",
				Source: "/home/@{bogus:val}/",
				Target: "/backup/",
			},
		},
	}

	_, err := ResolveConfig(cfg)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnresolvedMacro)
}

func TestValidateNoUnresolvedMacros(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "AllResolved",
			cfg: Config{
				Jobs: []Job{{Name: "job1", Source: "/home/user/", Target: "/backup/user/"}},
			},
			wantErr: false,
		},
		{
			name: "UnresolvedInSource",
			cfg: Config{
				Jobs: []Job{{Name: "job1", Source: "/home/@{upper:user}/", Target: "/backup/user/"}},
			},
			wantErr: true,
		},
		{
			name: "UnresolvedInTarget",
			cfg: Config{
				Jobs: []Job{{Name: "job1", Source: "/home/user/", Target: "/backup/@{lower:user}/"}},
			},
			wantErr: true,
		},
		{
			name: "UnresolvedInName",
			cfg: Config{
				Jobs: []Job{{Name: "@{upper:job}", Source: "/home/user/", Target: "/backup/user/"}},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateNoUnresolvedMacros(test.cfg)
			if test.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrUnresolvedMacro)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadResolvedConfigWithMacros(t *testing.T) {
	yamlContent := testutil.NewConfigBuilder().
		Source("/home/jaap").Target("/backup").
		Variable("user", "jaap").
		AddJob("jaap_docs", "/home/jaap/docs", "/backup/@{capitalize:${user}}/docs").
		Build()

	path := testutil.WriteConfigFile(t, yamlContent)

	cfg, err := LoadResolvedConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "/backup/Jaap/docs", cfg.Jobs[0].Target)
}

func TestLoadResolvedConfigWithMacros_Error(t *testing.T) {
	yamlContent := testutil.NewConfigBuilder().
		Source("/home/jaap").Target("/backup").
		AddJob("job1", "/home/jaap/docs", "/backup/@{nonexistent:val}/docs").
		Build()

	path := testutil.WriteConfigFile(t, yamlContent)

	_, err := LoadResolvedConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config resolution failed")
}
