package lint2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatKotsMessage(t *testing.T) {
	tests := []struct {
		name     string
		expr     KotsLintExpression
		expected string
	}{
		{
			name: "message with line number and rule",
			expr: KotsLintExpression{
				Rule:    "config-option-invalid-type",
				Message: "Config option \"test\" has an invalid type",
				Positions: []KotsLintPosition{
					{Start: KotsLintLinePosition{Line: 10}},
				},
			},
			expected: "[config-option-invalid-type] line 10: Config option \"test\" has an invalid type",
		},
		{
			name: "message with rule only",
			expr: KotsLintExpression{
				Rule:    "config-group-missing-title",
				Message: "Config group is missing a title",
			},
			expected: "[config-group-missing-title] Config group is missing a title",
		},
		{
			name: "message with line number only",
			expr: KotsLintExpression{
				Message: "Syntax error detected",
				Positions: []KotsLintPosition{
					{Start: KotsLintLinePosition{Line: 25}},
				},
			},
			expected: "line 25: Syntax error detected",
		},
		{
			name: "plain message",
			expr: KotsLintExpression{
				Message: "General validation warning",
			},
			expected: "General validation warning",
		},
		{
			name: "line number zero is ignored",
			expr: KotsLintExpression{
				Rule:    "some-rule",
				Message: "Some message",
				Positions: []KotsLintPosition{
					{Start: KotsLintLinePosition{Line: 0}},
				},
			},
			expected: "[some-rule] Some message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatKotsMessage(tt.expr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeKotsSeverity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "error maps to ERROR",
			input:    "error",
			expected: "ERROR",
		},
		{
			name:     "warn maps to WARNING",
			input:    "warn",
			expected: "WARNING",
		},
		{
			name:     "info maps to INFO",
			input:    "info",
			expected: "INFO",
		},
		{
			name:     "unknown maps to INFO",
			input:    "unknown",
			expected: "INFO",
		},
		{
			name:     "empty maps to INFO",
			input:    "",
			expected: "INFO",
		},
		{
			name:     "uppercase ERROR maps to ERROR (case-insensitive)",
			input:    "ERROR",
			expected: "ERROR",
		},
		{
			name:     "mixed case Error maps to ERROR (case-insensitive)",
			input:    "Error",
			expected: "ERROR",
		},
		{
			name:     "uppercase WARN maps to WARNING (case-insensitive)",
			input:    "WARN",
			expected: "WARNING",
		},
		{
			name:     "mixed case Warn maps to WARNING (case-insensitive)",
			input:    "Warn",
			expected: "WARNING",
		},
		{
			name:     "uppercase INFO maps to INFO (case-insensitive)",
			input:    "INFO",
			expected: "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeKotsSeverity(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseKotsOutput_Valid(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		expectedCount int
		expectedFirst LintMessage
		expectError   bool
		errorContains string
	}{
		{
			name: "single error",
			output: `{
				"lintExpressions": [
					{
						"rule": "config-option-invalid-type",
						"type": "error",
						"message": "Config option \"test\" has an invalid type",
						"path": "kots-config.yaml",
						"positions": [{"start": {"line": 10}}]
					}
				],
				"isLintingComplete": true
			}`,
			expectedCount: 1,
			expectedFirst: LintMessage{
				Severity: "ERROR",
				Message:  "[config-option-invalid-type] line 10: Config option \"test\" has an invalid type",
				Path:     "kots-config.yaml",
			},
			expectError: false,
		},
		{
			name: "multiple mixed severities",
			output: `{
				"lintExpressions": [
					{
						"rule": "error-rule",
						"type": "error",
						"message": "Critical error",
						"path": "config.yaml",
						"positions": [{"start": {"line": 5}}]
					},
					{
						"rule": "warning-rule",
						"type": "warn",
						"message": "Warning message",
						"path": "config.yaml",
						"positions": [{"start": {"line": 10}}]
					},
					{
						"rule": "info-rule",
						"type": "info",
						"message": "Info message",
						"path": "config.yaml"
					}
				],
				"isLintingComplete": true
			}`,
			expectedCount: 3,
			expectedFirst: LintMessage{
				Severity: "ERROR",
				Message:  "[error-rule] line 5: Critical error",
				Path:     "config.yaml",
			},
			expectError: false,
		},
		{
			name: "empty lint expressions",
			output: `{
				"lintExpressions": [],
				"isLintingComplete": true
			}`,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "empty output",
			output:        "",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "json with surrounding text",
			output: `
Initializing KOTS linter...
  • Loading configuration...
{
	"lintExpressions": [
		{
			"rule": "test-rule",
			"type": "warn",
			"message": "Test warning",
			"path": "test.yaml"
		}
	],
	"isLintingComplete": true
}
  ✓ Linting complete!
`,
			expectedCount: 1,
			expectedFirst: LintMessage{
				Severity: "WARNING",
				Message:  "[test-rule] Test warning",
				Path:     "test.yaml",
			},
			expectError: false,
		},
		{
			name:          "malformed json",
			output:        `{"lintExpressions": [}`,
			expectedCount: 0,
			expectError:   true,
			errorContains: "failed to extract JSON",
		},
		{
			name:          "no json in output",
			output:        "Error: file not found",
			expectedCount: 0,
			expectError:   true,
			errorContains: "failed to extract JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := parseKotsOutput(tt.output)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Len(t, messages, tt.expectedCount)

			if tt.expectedCount > 0 {
				assert.Equal(t, tt.expectedFirst.Severity, messages[0].Severity)
				assert.Equal(t, tt.expectedFirst.Message, messages[0].Message)
				assert.Equal(t, tt.expectedFirst.Path, messages[0].Path)
			}
		})
	}
}

func TestParseKotsOutput_EdgeCases(t *testing.T) {
	t.Run("message without positions", func(t *testing.T) {
		output := `{
			"lintExpressions": [
				{
					"rule": "test-rule",
					"type": "error",
					"message": "Test error",
					"path": "test.yaml"
				}
			],
			"isLintingComplete": true
		}`

		messages, err := parseKotsOutput(output)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		// Should include rule but not line number
		assert.Equal(t, "[test-rule] Test error", messages[0].Message)
	})

	t.Run("message without rule", func(t *testing.T) {
		output := `{
			"lintExpressions": [
				{
					"type": "warn",
					"message": "Warning without rule",
					"path": "test.yaml",
					"positions": [{"start": {"line": 15}}]
				}
			],
			"isLintingComplete": true
		}`

		messages, err := parseKotsOutput(output)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		// Should include line number but not rule
		assert.Equal(t, "line 15: Warning without rule", messages[0].Message)
	})

	t.Run("message without path", func(t *testing.T) {
		output := `{
			"lintExpressions": [
				{
					"rule": "global-rule",
					"type": "info",
					"message": "Global info"
				}
			],
			"isLintingComplete": true
		}`

		messages, err := parseKotsOutput(output)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		assert.Equal(t, "", messages[0].Path)
		assert.Equal(t, "[global-rule] Global info", messages[0].Message)
	})
}

func TestParseKotsOutput_SeverityCounts(t *testing.T) {
	output := `{
		"lintExpressions": [
			{"type": "error", "message": "Error 1", "path": "test.yaml"},
			{"type": "error", "message": "Error 2", "path": "test.yaml"},
			{"type": "warn", "message": "Warning 1", "path": "test.yaml"},
			{"type": "warn", "message": "Warning 2", "path": "test.yaml"},
			{"type": "warn", "message": "Warning 3", "path": "test.yaml"},
			{"type": "info", "message": "Info 1", "path": "test.yaml"},
			{"type": "info", "message": "Info 2", "path": "test.yaml"},
			{"type": "info", "message": "Info 3", "path": "test.yaml"},
			{"type": "info", "message": "Info 4", "path": "test.yaml"}
		],
		"isLintingComplete": true
	}`

	messages, err := parseKotsOutput(output)
	require.NoError(t, err)
	require.Len(t, messages, 9)

	// Count by severity
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, msg := range messages {
		switch msg.Severity {
		case "ERROR":
			errorCount++
		case "WARNING":
			warningCount++
		case "INFO":
			infoCount++
		}
	}

	assert.Equal(t, 2, errorCount, "expected 2 errors")
	assert.Equal(t, 3, warningCount, "expected 3 warnings")
	assert.Equal(t, 4, infoCount, "expected 4 infos")
}
