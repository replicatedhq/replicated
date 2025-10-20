package cmd

import (
	"bytes"
	"testing"
	"text/tabwriter"
)

func TestParseSetValues(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]interface{}
	}{
		{
			name:  "simple key-value",
			input: []string{"image=nginx"},
			expected: map[string]interface{}{
				"image": "nginx",
			},
		},
		{
			name:  "nested key-value",
			input: []string{"image.repository=nginx", "image.tag=1.19"},
			expected: map[string]interface{}{
				"image": map[string]interface{}{
					"repository": "nginx",
					"tag":        "1.19",
				},
			},
		},
		{
			name:  "deeply nested",
			input: []string{"a.b.c=value"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "value",
					},
				},
			},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: map[string]interface{}{},
		},
		{
			name:     "invalid format (no equals)",
			input:    []string{"invalid"},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSetValues(tt.input)

			// Simple comparison for string values
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d keys, got %d", len(tt.expected), len(result))
			}

			// For nested values, just check if keys exist
			for key := range tt.expected {
				if _, ok := result[key]; !ok {
					t.Errorf("expected key %q not found in result", key)
				}
			}
		})
	}
}

func TestReleaseExtractImages_Validation(t *testing.T) {
	tests := []struct {
		name        string
		yamlDir     string
		chart       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no input specified",
			yamlDir:     "",
			chart:       "",
			expectError: true,
			errorMsg:    "either --yaml-dir or --chart must be specified",
		},
		{
			name:        "both inputs specified",
			yamlDir:     "./manifests",
			chart:       "./chart.tgz",
			expectError: true,
			errorMsg:    "cannot specify both --yaml-dir and --chart",
		},
		{
			name:        "valid yaml-dir",
			yamlDir:     "../../pkg/imageextract/testdata/simple-deployment",
			chart:       "",
			expectError: false,
		},
		{
			name:        "valid chart",
			yamlDir:     "",
			chart:       "../../pkg/imageextract/testdata/helm-chart",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

			r := &runners{
				w:            w,
				outputFormat: "list",
				args: runnerArgs{
					extractImagesYamlDir: tt.yamlDir,
					extractImagesChart:   tt.chart,
				},
			}

			err := r.releaseExtractImages(nil, nil)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestReleaseExtractImages_OutputFormat(t *testing.T) {
	tests := []struct {
		name         string
		outputFormat string
		expectError  bool
	}{
		{
			name:         "valid table format",
			outputFormat: "table",
			expectError:  false,
		},
		{
			name:         "valid json format",
			outputFormat: "json",
			expectError:  false,
		},
		{
			name:         "valid list format",
			outputFormat: "list",
			expectError:  false,
		},
		{
			name:         "invalid format",
			outputFormat: "xml",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

			r := &runners{
				w:            w,
				outputFormat: tt.outputFormat,
				args: runnerArgs{
					extractImagesYamlDir: "../../pkg/imageextract/testdata/simple-deployment",
				},
			}

			err := r.releaseExtractImages(nil, nil)

			if tt.expectError {
				if err == nil {
					t.Error("expected error for invalid format")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
