package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAppSlugOrID(t *testing.T) {
	// Save and restore cache.DefaultApp to avoid mutating global state
	originalDefaultApp := cache.DefaultApp
	defer func() { cache.DefaultApp = originalDefaultApp }()

	tests := []struct {
		name           string
		flagValue      string
		envValue       string
		cacheDefault   string
		expectedResult string
	}{
		{
			name:           "explicit flag wins over everything",
			flagValue:      "flag-app",
			envValue:       "env-app",
			cacheDefault:   "cache-app",
			expectedResult: "flag-app",
		},
		{
			name:           "env var wins over cache default",
			flagValue:      "",
			envValue:       "env-app",
			cacheDefault:   "cache-app",
			expectedResult: "env-app",
		},
		{
			name:           "cache default is last resort",
			flagValue:      "",
			envValue:       "",
			cacheDefault:   "cache-app",
			expectedResult: "cache-app",
		},
		{
			name:           "nothing set returns empty",
			flagValue:      "",
			envValue:       "",
			cacheDefault:   "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.DefaultApp = tt.cacheDefault
			if tt.envValue != "" {
				t.Setenv("REPLICATED_APP", tt.envValue)
			} else {
				os.Unsetenv("REPLICATED_APP")
			}

			// Create a temp dir without .replicated to avoid cwd interference
			tmpDir := t.TempDir()
			origWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(origWd)

			result := resolveAppSlugOrID(tt.flagValue)
			if result != tt.expectedResult {
				t.Errorf("resolveAppSlugOrID() = %q, want %q", result, tt.expectedResult)
			}
		})
	}
}

func TestResolveAppSlugOrID_ReplicatedFilePrecedence(t *testing.T) {
	// Save and restore cache.DefaultApp
	originalDefaultApp := cache.DefaultApp
	defer func() { cache.DefaultApp = originalDefaultApp }()

	t.Run(".replicated wins over cache default", func(t *testing.T) {
		cache.DefaultApp = "cache-app"
		os.Unsetenv("REPLICATED_APP")

		// Create temp dir with .replicated config
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		if err := os.WriteFile(configPath, []byte("appSlug: file-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppSlugOrID("")
		if result != "file-app" {
			t.Errorf("resolveAppSlugOrID() = %q, want %q", result, "file-app")
		}
	})

	t.Run("env var wins over .replicated", func(t *testing.T) {
		cache.DefaultApp = "cache-app"
		t.Setenv("REPLICATED_APP", "env-app")

		// Create temp dir with .replicated config
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		if err := os.WriteFile(configPath, []byte("appSlug: file-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppSlugOrID("")
		if result != "env-app" {
			t.Errorf("resolveAppSlugOrID() = %q, want %q", result, "env-app")
		}
	})
}

func TestResolveAppFromConfig(t *testing.T) {
	t.Run("finds .replicated in cwd", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		if err := os.WriteFile(configPath, []byte("appSlug: my-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppFromConfig()
		if result != "my-app" {
			t.Errorf("resolveAppFromConfig() = %q, want %q", result, "my-app")
		}
	})

	t.Run("finds .replicated.yaml in cwd", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated.yaml")
		if err := os.WriteFile(configPath, []byte("appSlug: yaml-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppFromConfig()
		if result != "yaml-app" {
			t.Errorf("resolveAppFromConfig() = %q, want %q", result, "yaml-app")
		}
	})

	t.Run("finds .replicated in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		if err := os.WriteFile(configPath, []byte("appSlug: parent-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create child dir with no config
		childDir := filepath.Join(tmpDir, "child")
		if err := os.MkdirAll(childDir, 0755); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(childDir)
		defer os.Chdir(origWd)

		result := resolveAppFromConfig()
		if result != "parent-app" {
			t.Errorf("resolveAppFromConfig() = %q, want %q", result, "parent-app")
		}
	})

	t.Run("returns empty when no config found", func(t *testing.T) {
		tmpDir := t.TempDir()

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppFromConfig()
		if result != "" {
			t.Errorf("resolveAppFromConfig() = %q, want empty string", result)
		}
	})

	t.Run("prefers AppSlug over AppId", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		if err := os.WriteFile(configPath, []byte("appSlug: slug-app\nappId: id-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppFromConfig()
		if result != "slug-app" {
			t.Errorf("resolveAppFromConfig() = %q, want %q", result, "slug-app")
		}
	})

	t.Run("falls back to AppId when no AppSlug", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		if err := os.WriteFile(configPath, []byte("appId: id-app\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origWd)

		result := resolveAppFromConfig()
		if result != "id-app" {
			t.Errorf("resolveAppFromConfig() = %q, want %q", result, "id-app")
		}
	})
}
