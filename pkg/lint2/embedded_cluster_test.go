package lint2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseECVersionFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "version with k8s suffix",
			content: `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "3.0.0+k8s-1.31"
`,
			want: "3.0.0",
		},
		{
			name: "pre-release version with k8s suffix",
			content: `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "3.0.0-beta.2+k8s-1.34"
`,
			want: "3.0.0-beta.2",
		},
		{
			name: "complex pre-release with k8s suffix",
			content: `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "3.0.0-rc0-35-g9b94-v3+k8s-1.33"
`,
			want: "3.0.0-rc0-35-g9b94-v3",
		},
		{
			name: "wrong kind returns empty",
			content: `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: NotConfig
spec:
  version: "3.0.0+k8s-1.31"
`,
			want: "",
		},
		{
			name: "multi-doc file returns ec config version",
			content: `apiVersion: v1
kind: ConfigMap
---
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "3.0.0-beta.2+k8s-1.34"
`,
			want: "3.0.0-beta.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "ec-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			f.Close()

			got, err := parseECVersionFromFile(f.Name())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiscoverECVersion(t *testing.T) {
	dir := t.TempDir()
	content := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "3.0.0-beta.2+k8s-1.34"
`
	if err := os.WriteFile(filepath.Join(dir, "ec.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := DiscoverECVersion([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "3.0.0-beta.2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
