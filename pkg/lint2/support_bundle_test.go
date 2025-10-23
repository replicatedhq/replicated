package lint2

import (
	"testing"
)

func TestParseSupportBundleOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []LintMessage
		wantErr  bool
	}{
		{
			name: "valid spec with warning",
			output: `{
  "results": [
    {
      "filePath": "/tmp/support-bundle-test/valid-spec.yaml",
      "errors": [],
      "warnings": [
        {
          "line": 5,
          "column": 0,
          "message": "Some collectors are missing docString (recommended for v1beta3)",
          "field": "spec"
        }
      ]
    }
  ]
}`,
			expected: []LintMessage{
				{
					Severity: "WARNING",
					Path:     "/tmp/support-bundle-test/valid-spec.yaml",
					Message:  "line 5: Some collectors are missing docString (recommended for v1beta3) (field: spec)",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid yaml with error",
			output: `{
  "results": [
    {
      "filePath": "/tmp/support-bundle-test/invalid-yaml.yaml",
      "errors": [
        {
          "line": 15,
          "column": 0,
          "message": "YAML syntax error: error converting YAML to JSON: yaml: line 15: mapping values are not allowed in this context",
          "field": ""
        }
      ],
      "warnings": []
    }
  ]
}`,
			expected: []LintMessage{
				{
					Severity: "ERROR",
					Path:     "/tmp/support-bundle-test/invalid-yaml.yaml",
					Message:  "line 15: YAML syntax error: error converting YAML to JSON: yaml: line 15: mapping values are not allowed in this context",
				},
			},
			wantErr: false,
		},
		{
			name: "multiple errors and warnings",
			output: `{
  "results": [
    {
      "filePath": "/tmp/support-bundle-test/missing-fields.yaml",
      "errors": [
        {
          "line": 8,
          "column": 0,
          "message": "Support bundle spec must have at least one collector",
          "field": "spec.collectors"
        }
      ],
      "warnings": [
        {
          "line": 6,
          "column": 0,
          "message": "Some collectors are missing docString (recommended for v1beta3)",
          "field": "spec.collectors"
        }
      ]
    }
  ]
}`,
			expected: []LintMessage{
				{
					Severity: "ERROR",
					Path:     "/tmp/support-bundle-test/missing-fields.yaml",
					Message:  "line 8: Support bundle spec must have at least one collector (field: spec.collectors)",
				},
				{
					Severity: "WARNING",
					Path:     "/tmp/support-bundle-test/missing-fields.yaml",
					Message:  "line 6: Some collectors are missing docString (recommended for v1beta3) (field: spec.collectors)",
				},
			},
			wantErr: false,
		},
		{
			name: "multiple files",
			output: `{
  "results": [
    {
      "filePath": "/tmp/spec1.yaml",
      "errors": [
        {
          "line": 10,
          "column": 0,
          "message": "Missing required field",
          "field": "spec.collectors"
        }
      ],
      "warnings": []
    },
    {
      "filePath": "/tmp/spec2.yaml",
      "errors": [],
      "warnings": [
        {
          "line": 5,
          "column": 0,
          "message": "Deprecated field usage",
          "field": "spec.hostCollectors"
        }
      ]
    }
  ]
}`,
			expected: []LintMessage{
				{
					Severity: "ERROR",
					Path:     "/tmp/spec1.yaml",
					Message:  "line 10: Missing required field (field: spec.collectors)",
				},
				{
					Severity: "WARNING",
					Path:     "/tmp/spec2.yaml",
					Message:  "line 5: Deprecated field usage (field: spec.hostCollectors)",
				},
			},
			wantErr: false,
		},
		{
			name: "no issues",
			output: `{
  "results": [
    {
      "filePath": "/tmp/valid.yaml",
      "errors": [],
      "warnings": []
    }
  ]
}`,
			expected: []LintMessage{},
			wantErr:  false,
		},
		{
			name:     "empty results",
			output:   `{"results": []}`,
			expected: []LintMessage{},
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			output:   `not valid json`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "empty output",
			output:   ``,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "info severity support",
			output: `{
  "results": [
    {
      "filePath": "/tmp/spec-with-info.yaml",
      "errors": [],
      "warnings": [],
      "infos": [
        {
          "line": 3,
          "column": 0,
          "message": "Consider adding description field",
          "field": "metadata"
        }
      ]
    }
  ]
}`,
			expected: []LintMessage{
				{
					Severity: "INFO",
					Path:     "/tmp/spec-with-info.yaml",
					Message:  "line 3: Consider adding description field (field: metadata)",
				},
			},
			wantErr: false,
		},
		{
			name: "error message with braces before JSON",
			output: `Error: failed to parse {invalid} syntax
{
  "results": [
    {
      "filePath": "/tmp/spec.yaml",
      "errors": [
        {
          "line": 10,
          "column": 0,
          "message": "Validation failed",
          "field": "spec"
        }
      ],
      "warnings": []
    }
  ]
}`,
			expected: []LintMessage{
				{
					Severity: "ERROR",
					Path:     "/tmp/spec.yaml",
					Message:  "line 10: Validation failed (field: spec)",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSupportBundleOutput(tt.output)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSupportBundleOutput() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseSupportBundleOutput() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("parseSupportBundleOutput() returned %d messages, want %d", len(result), len(tt.expected))
				return
			}

			for i, msg := range result {
				expected := tt.expected[i]
				if msg.Severity != expected.Severity {
					t.Errorf("Message %d: Severity = %q, want %q", i, msg.Severity, expected.Severity)
				}
				if msg.Path != expected.Path {
					t.Errorf("Message %d: Path = %q, want %q", i, msg.Path, expected.Path)
				}
				if msg.Message != expected.Message {
					t.Errorf("Message %d: Message = %q, want %q", i, msg.Message, expected.Message)
				}
			}
		})
	}
}

func TestFormatSupportBundleMessage(t *testing.T) {
	tests := []struct {
		name     string
		issue    SupportBundleLintIssue
		expected string
	}{
		{
			name: "full issue with line and field",
			issue: SupportBundleLintIssue{
				Line:    10,
				Column:  0,
				Message: "Missing required field",
				Field:   "spec.collectors",
			},
			expected: "line 10: Missing required field (field: spec.collectors)",
		},
		{
			name: "issue with line only",
			issue: SupportBundleLintIssue{
				Line:    5,
				Column:  0,
				Message: "YAML syntax error",
				Field:   "",
			},
			expected: "line 5: YAML syntax error",
		},
		{
			name: "issue with field only",
			issue: SupportBundleLintIssue{
				Line:    0,
				Column:  0,
				Message: "Deprecated usage",
				Field:   "spec.hostCollectors",
			},
			expected: "Deprecated usage (field: spec.hostCollectors)",
		},
		{
			name: "issue with message only",
			issue: SupportBundleLintIssue{
				Line:    0,
				Column:  0,
				Message: "General warning",
				Field:   "",
			},
			expected: "General warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTroubleshootMessage(tt.issue)
			if result != tt.expected {
				t.Errorf("formatTroubleshootMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}
