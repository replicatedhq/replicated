package lint2

import (
	"testing"
)

func TestParseHelmOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []LintMessage
	}{
		{
			name: "single INFO message with path",
			output: `[INFO] Chart.yaml: icon is recommended
`,
			expected: []LintMessage{
				{Severity: "INFO", Path: "Chart.yaml", Message: "icon is recommended"},
			},
		},
		{
			name: "multiple messages with different severities",
			output: `[INFO] Chart.yaml: icon is recommended
[WARNING] templates/deployment.yaml: image tag should be specified
[ERROR] values.yaml: missing required field 'replicaCount'
`,
			expected: []LintMessage{
				{Severity: "INFO", Path: "Chart.yaml", Message: "icon is recommended"},
				{Severity: "WARNING", Path: "templates/deployment.yaml", Message: "image tag should be specified"},
				{Severity: "ERROR", Path: "values.yaml", Message: "missing required field 'replicaCount'"},
			},
		},
		{
			name: "message without path",
			output: `[WARNING] chart is deprecated
`,
			expected: []LintMessage{
				{Severity: "WARNING", Path: "", Message: "chart is deprecated"},
			},
		},
		{
			name: "mixed messages with and without paths",
			output: `==> Linting ./test-chart
[INFO] Chart.yaml: icon is recommended
[WARNING] chart directory not found

1 chart(s) linted, 0 chart(s) failed
`,
			expected: []LintMessage{
				{Severity: "INFO", Path: "Chart.yaml", Message: "icon is recommended"},
				{Severity: "WARNING", Path: "", Message: "chart directory not found"},
			},
		},
		{
			name:     "empty output",
			output:   "",
			expected: []LintMessage{},
		},
		{
			name: "output with only headers",
			output: `==> Linting ./test-chart

1 chart(s) linted, 0 chart(s) failed
`,
			expected: []LintMessage{},
		},
		{
			name: "message with colon in message text",
			output: `[ERROR] values.yaml: key 'foo': value must be a string
`,
			expected: []LintMessage{
				{Severity: "ERROR", Path: "values.yaml", Message: "key 'foo': value must be a string"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseHelmOutput(tt.output)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseHelmOutput() returned %d messages, want %d", len(result), len(tt.expected))
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

func TestParseHelmOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		wantLen      int
		desc         string
		wantSeverity string
		wantPath     string
		wantMessage  string
	}{
		{
			name:    "whitespace only",
			output:  "   \n  \t  \n   ",
			wantLen: 0,
			desc:    "should ignore whitespace-only lines",
		},
		{
			name:    "malformed severity",
			output:  "[INVALID] Chart.yaml: some message\n",
			wantLen: 0,
			desc:    "should ignore messages with invalid severity",
		},
		{
			name: "message with multiple colons",
			output: `[INFO] templates/service.yaml: port: should be number: got string
`,
			wantLen:      1,
			desc:         "should handle multiple colons in message",
			wantSeverity: "INFO",
			wantPath:     "templates/service.yaml",
			wantMessage:  "port: should be number: got string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseHelmOutput(tt.output)
			if len(result) != tt.wantLen {
				t.Errorf("%s: got %d messages, want %d", tt.desc, len(result), tt.wantLen)
				return
			}
			// Validate parsed structure for tests that expect messages
			if tt.wantLen > 0 && tt.wantSeverity != "" {
				if result[0].Severity != tt.wantSeverity {
					t.Errorf("%s: got Severity=%q, want %q", tt.desc, result[0].Severity, tt.wantSeverity)
				}
				if result[0].Path != tt.wantPath {
					t.Errorf("%s: got Path=%q, want %q", tt.desc, result[0].Path, tt.wantPath)
				}
				if result[0].Message != tt.wantMessage {
					t.Errorf("%s: got Message=%q, want %q", tt.desc, result[0].Message, tt.wantMessage)
				}
			}
		})
	}
}
