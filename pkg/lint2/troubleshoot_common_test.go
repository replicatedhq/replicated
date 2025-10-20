package lint2

import (
	"testing"
)

func TestParseTroubleshootJSON_Preflight(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{
			name: "valid preflight JSON",
			output: `{
  "results": [
    {
      "filePath": "/tmp/test.yaml",
      "errors": [
        {
          "line": 10,
          "column": 0,
          "message": "Test error",
          "field": "spec"
        }
      ],
      "warnings": [],
      "infos": []
    }
  ]
}`,
			wantErr: false,
		},
		{
			name: "error message with braces before JSON",
			output: `Error: failed to parse {invalid} syntax
{
  "results": [
    {
      "filePath": "/tmp/test.yaml",
      "errors": [],
      "warnings": [],
      "infos": []
    }
  ]
}`,
			wantErr: false,
		},
		{
			name:    "no JSON in output",
			output:  "Error: no JSON here",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			output:  "{not valid json}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTroubleshootJSON[PreflightLintIssue](tt.output)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTroubleshootJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseTroubleshootJSON() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("parseTroubleshootJSON() returned nil result")
			}
		})
	}
}

func TestParseTroubleshootJSON_SupportBundle(t *testing.T) {
	output := `{
  "results": [
    {
      "filePath": "/tmp/support-bundle.yaml",
      "errors": [
        {
          "line": 5,
          "column": 0,
          "message": "Missing collectors",
          "field": "spec.collectors"
        }
      ],
      "warnings": [],
      "infos": []
    }
  ]
}`

	result, err := parseTroubleshootJSON[SupportBundleLintIssue](output)
	if err != nil {
		t.Fatalf("parseTroubleshootJSON() unexpected error: %v", err)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}

	if len(result.Results[0].Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Results[0].Errors))
	}
}

func TestFormatTroubleshootMessage_Preflight(t *testing.T) {
	tests := []struct {
		name     string
		issue    PreflightLintIssue
		expected string
	}{
		{
			name: "full issue with line and field",
			issue: PreflightLintIssue{
				Line:    10,
				Column:  5,
				Message: "Test message",
				Field:   "spec.collectors",
			},
			expected: "line 10: Test message (field: spec.collectors)",
		},
		{
			name: "issue with line only",
			issue: PreflightLintIssue{
				Line:    5,
				Message: "Line message",
			},
			expected: "line 5: Line message",
		},
		{
			name: "issue with field only",
			issue: PreflightLintIssue{
				Message: "Field message",
				Field:   "metadata",
			},
			expected: "Field message (field: metadata)",
		},
		{
			name: "issue with message only",
			issue: PreflightLintIssue{
				Message: "Simple message",
			},
			expected: "Simple message",
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

func TestFormatTroubleshootMessage_SupportBundle(t *testing.T) {
	issue := SupportBundleLintIssue{
		Line:    15,
		Column:  0,
		Message: "Support bundle error",
		Field:   "spec",
	}

	expected := "line 15: Support bundle error (field: spec)"
	result := formatTroubleshootMessage(issue)

	if result != expected {
		t.Errorf("formatTroubleshootMessage() = %q, want %q", result, expected)
	}
}

func TestConvertTroubleshootResultToMessages_Preflight(t *testing.T) {
	result := &TroubleshootLintResult[PreflightLintIssue]{
		Results: []TroubleshootFileResult[PreflightLintIssue]{
			{
				FilePath: "/tmp/test.yaml",
				Errors: []PreflightLintIssue{
					{Line: 10, Message: "Error message", Field: "spec"},
				},
				Warnings: []PreflightLintIssue{
					{Line: 5, Message: "Warning message"},
				},
				Infos: []PreflightLintIssue{
					{Message: "Info message"},
				},
			},
		},
	}

	messages := convertTroubleshootResultToMessages(result)

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check error
	if messages[0].Severity != "ERROR" {
		t.Errorf("Expected first message severity ERROR, got %s", messages[0].Severity)
	}
	if messages[0].Path != "/tmp/test.yaml" {
		t.Errorf("Expected path /tmp/test.yaml, got %s", messages[0].Path)
	}

	// Check warning
	if messages[1].Severity != "WARNING" {
		t.Errorf("Expected second message severity WARNING, got %s", messages[1].Severity)
	}

	// Check info
	if messages[2].Severity != "INFO" {
		t.Errorf("Expected third message severity INFO, got %s", messages[2].Severity)
	}
}

func TestConvertTroubleshootResultToMessages_SupportBundle(t *testing.T) {
	result := &TroubleshootLintResult[SupportBundleLintIssue]{
		Results: []TroubleshootFileResult[SupportBundleLintIssue]{
			{
				FilePath: "/tmp/support-bundle.yaml",
				Errors: []SupportBundleLintIssue{
					{Line: 8, Message: "Missing collectors", Field: "spec.collectors"},
				},
				Warnings: []SupportBundleLintIssue{},
				Infos:    []SupportBundleLintIssue{},
			},
		},
	}

	messages := convertTroubleshootResultToMessages(result)

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Severity != "ERROR" {
		t.Errorf("Expected severity ERROR, got %s", messages[0].Severity)
	}

	expectedMsg := "line 8: Missing collectors (field: spec.collectors)"
	if messages[0].Message != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, messages[0].Message)
	}
}

func TestTroubleshootIssueInterface_Preflight(t *testing.T) {
	issue := PreflightLintIssue{
		Line:    10,
		Column:  5,
		Message: "Test",
		Field:   "spec",
	}

	// Test interface implementation
	var _ TroubleshootIssue = issue

	if issue.GetLine() != 10 {
		t.Errorf("GetLine() = %d, want 10", issue.GetLine())
	}
	if issue.GetColumn() != 5 {
		t.Errorf("GetColumn() = %d, want 5", issue.GetColumn())
	}
	if issue.GetMessage() != "Test" {
		t.Errorf("GetMessage() = %q, want %q", issue.GetMessage(), "Test")
	}
	if issue.GetField() != "spec" {
		t.Errorf("GetField() = %q, want %q", issue.GetField(), "spec")
	}
}

func TestTroubleshootIssueInterface_SupportBundle(t *testing.T) {
	issue := SupportBundleLintIssue{
		Line:    15,
		Column:  2,
		Message: "Bundle test",
		Field:   "metadata",
	}

	// Test interface implementation
	var _ TroubleshootIssue = issue

	if issue.GetLine() != 15 {
		t.Errorf("GetLine() = %d, want 15", issue.GetLine())
	}
	if issue.GetColumn() != 2 {
		t.Errorf("GetColumn() = %d, want 2", issue.GetColumn())
	}
	if issue.GetMessage() != "Bundle test" {
		t.Errorf("GetMessage() = %q, want %q", issue.GetMessage(), "Bundle test")
	}
	if issue.GetField() != "metadata" {
		t.Errorf("GetField() = %q, want %q", issue.GetField(), "metadata")
	}
}
