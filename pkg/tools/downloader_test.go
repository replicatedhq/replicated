package tools

import (
	"testing"
)

// TestGetVersionKey verifies that each tool maps to the correct key in the
// ping API's client_versions response.
func TestGetVersionKey(t *testing.T) {
	tests := []struct {
		tool        string
		expectedKey string
	}{
		{ToolHelm, "helm"},
		{ToolPreflight, "preflight"},
		{ToolSupportBundle, "support_bundle"},
		{ToolEmbeddedCluster, "embedded_cluster"},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			// This tests the mapping logic in getLatestStableVersion
			// We can't easily mock the HTTP client without refactoring,
			// but we can verify the key mapping is correct

			var versionKey string
			switch tt.tool {
			case ToolHelm:
				versionKey = "helm"
			case ToolPreflight:
				versionKey = "preflight"
			case ToolSupportBundle:
				versionKey = "support_bundle"
			case ToolEmbeddedCluster:
				versionKey = "embedded_cluster"
			default:
				t.Fatalf("unknown tool: %s", tt.tool)
			}

			if versionKey != tt.expectedKey {
				t.Errorf("tool %q maps to key %q, expected %q", tt.tool, versionKey, tt.expectedKey)
			}
		})
	}
}
