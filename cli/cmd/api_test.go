package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNormalizeAPIPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         string
		wantPath    string
		wantVersion string
		wantErr     bool
	}{
		{name: "v3 with leading slash", raw: "/v3/apps", wantPath: "/v3/apps", wantVersion: "v3"},
		{name: "v1 without leading slash", raw: "v1/apps", wantPath: "/v1/apps", wantVersion: "v1"},
		{name: "v2", raw: "/v2/license", wantPath: "/v2/license", wantVersion: "v2"},
		{name: "unsupported version", raw: "/v4/apps", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "no version", raw: "/apps", wantErr: true},
		{name: "slash only", raw: "/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path, version, err := normalizeAPIPath(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got path=%q version=%q", path, version)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != tt.wantPath {
				t.Fatalf("path = %q, want %q", path, tt.wantPath)
			}
			if version != tt.wantVersion {
				t.Fatalf("version = %q, want %q", version, tt.wantVersion)
			}
		})
	}
}

func TestDoAPIRequestRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	r := &runners{}
	err := r.doAPIRequest(context.Background(), "GET", "/v9/apps", "")
	if err == nil {
		t.Fatal("expected error for unsupported API version")
	}
	if !strings.Contains(err.Error(), "unsupported API version") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoAPIRequestRequiresClient(t *testing.T) {
	t.Parallel()

	r := &runners{}
	err := r.doAPIRequest(context.Background(), "GET", "/v1/apps", "")
	if err == nil {
		t.Fatal("expected error when platformAPI is nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}
