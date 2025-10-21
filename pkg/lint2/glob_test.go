package lint2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlob_RecursiveDoublestar(t *testing.T) {
	// Test that ** matches recursively (zero or more directories)
	tmpDir := t.TempDir()

	// Create nested directory structure
	// manifests/
	// ├── app.yaml              (level 0 - no intermediate dir)
	// ├── base/
	// │   ├── deployment.yaml   (level 1)
	// │   └── service.yaml      (level 1)
	// └── overlays/
	//     └── prod/
	//         └── patch.yaml    (level 2)

	manifestsDir := filepath.Join(tmpDir, "manifests")
	baseDir := filepath.Join(manifestsDir, "base")
	prodDir := filepath.Join(manifestsDir, "overlays", "prod")

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(prodDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create YAML files at different depths
	files := []string{
		filepath.Join(manifestsDir, "app.yaml"),
		filepath.Join(baseDir, "deployment.yaml"),
		filepath.Join(baseDir, "service.yaml"),
		filepath.Join(prodDir, "patch.yaml"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("test: yaml"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test 1: ** should match ALL files recursively
	pattern := filepath.Join(manifestsDir, "**", "*.yaml")
	matches, err := Glob(pattern)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	// Should match all 4 files
	if len(matches) != 4 {
		t.Errorf("Glob(%q) returned %d files, want 4", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	// Verify all files are matched
	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}
	for _, expected := range files {
		if !matchMap[expected] {
			t.Errorf("Expected file %s not found in matches", expected)
		}
	}

	// Test 2: ** at end should match all directories
	pattern2 := filepath.Join(manifestsDir, "**")
	matches2, err := Glob(pattern2)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	// Should match the manifests dir itself, subdirs, and all files
	if len(matches2) < 4 {
		t.Errorf("Glob(%q) returned %d items, want at least 4", pattern2, len(matches2))
	}

	// Test 3: **/ in middle of pattern
	pattern3 := filepath.Join(tmpDir, "**", "*.yaml")
	matches3, err := Glob(pattern3)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	// Should still match all 4 yaml files
	if len(matches3) != 4 {
		t.Errorf("Glob(%q) returned %d files, want 4", pattern3, len(matches3))
	}
}

func TestGlob_DoublestarMatchesZeroLevels(t *testing.T) {
	// Test that ** matches ZERO directories (not just one+)
	tmpDir := t.TempDir()

	// Create files at root and in subdirectory
	rootFile := filepath.Join(tmpDir, "root.yaml")
	if err := os.WriteFile(rootFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	subFile := filepath.Join(subDir, "sub.yaml")
	if err := os.WriteFile(subFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// ** should match both zero levels (root.yaml) and one level (sub/sub.yaml)
	pattern := filepath.Join(tmpDir, "**", "*.yaml")
	matches, err := Glob(pattern)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Glob(%q) returned %d files, want 2 (root and subdirectory)", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}

	if !matchMap[rootFile] {
		t.Error("Expected root.yaml to be matched (** matches zero levels)")
	}
	if !matchMap[subFile] {
		t.Error("Expected sub/sub.yaml to be matched")
	}
}

func TestGlob_BraceExpansion(t *testing.T) {
	// Test {a,b,c} brace expansion
	tmpDir := t.TempDir()

	// Create directories: app, api, web
	dirs := []string{"app", "api", "web", "other"}
	for _, dir := range dirs {
		dirPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatal(err)
		}
		// Create a file in each
		filePath := filepath.Join(dirPath, "Chart.yaml")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test brace expansion: should match app, api, web but not other
	pattern := filepath.Join(tmpDir, "{app,api,web}", "Chart.yaml")
	matches, err := Glob(pattern)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(matches) != 3 {
		t.Errorf("Glob(%q) returned %d files, want 3", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	// Verify correct files matched
	expectedFiles := []string{
		filepath.Join(tmpDir, "app", "Chart.yaml"),
		filepath.Join(tmpDir, "api", "Chart.yaml"),
		filepath.Join(tmpDir, "web", "Chart.yaml"),
	}

	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}

	for _, expected := range expectedFiles {
		if !matchMap[expected] {
			t.Errorf("Expected file %s not found in matches", expected)
		}
	}

	// Verify "other" was NOT matched
	otherFile := filepath.Join(tmpDir, "other", "Chart.yaml")
	if matchMap[otherFile] {
		t.Error("File 'other/Chart.yaml' should NOT be matched by brace expansion")
	}
}

func TestGlob_CombinedDoublestarAndBraces(t *testing.T) {
	// Test combining ** and {} in same pattern
	tmpDir := t.TempDir()

	// Create structure:
	// dev/
	//   charts/app.yaml
	// prod/
	//   charts/app.yaml
	// staging/
	//   charts/app.yaml

	envs := []string{"dev", "prod", "staging"}
	for _, env := range envs {
		chartsDir := filepath.Join(tmpDir, env, "charts")
		if err := os.MkdirAll(chartsDir, 0755); err != nil {
			t.Fatal(err)
		}
		appFile := filepath.Join(chartsDir, "app.yaml")
		if err := os.WriteFile(appFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Pattern: {dev,prod}/**/*.yaml should match dev and prod, but not staging
	pattern := filepath.Join(tmpDir, "{dev,prod}", "**", "*.yaml")
	matches, err := Glob(pattern)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Glob(%q) returned %d files, want 2 (dev and prod only)", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}

	// Should match dev and prod
	if !matchMap[filepath.Join(tmpDir, "dev", "charts", "app.yaml")] {
		t.Error("Expected dev/charts/app.yaml to be matched")
	}
	if !matchMap[filepath.Join(tmpDir, "prod", "charts", "app.yaml")] {
		t.Error("Expected prod/charts/app.yaml to be matched")
	}

	// Should NOT match staging
	if matchMap[filepath.Join(tmpDir, "staging", "charts", "app.yaml")] {
		t.Error("staging/charts/app.yaml should NOT be matched")
	}
}

func TestGlob_BasicPatternsStillWork(t *testing.T) {
	// Verify that basic glob patterns (* ? []) still work correctly
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"app1.yaml",
		"app2.yaml",
		"api.yaml",
		"web.yml",
		"config.txt",
	}

	for _, file := range files {
		filePath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		pattern     string
		wantCount   int
		wantFiles   []string
		description string
	}{
		{
			name:        "star wildcard",
			pattern:     filepath.Join(tmpDir, "*.yaml"),
			wantCount:   3,
			wantFiles:   []string{"app1.yaml", "app2.yaml", "api.yaml"},
			description: "* should match all .yaml files",
		},
		{
			name:        "question mark",
			pattern:     filepath.Join(tmpDir, "app?.yaml"),
			wantCount:   2,
			wantFiles:   []string{"app1.yaml", "app2.yaml"},
			description: "? should match single character",
		},
		{
			name:        "character class",
			pattern:     filepath.Join(tmpDir, "app[12].yaml"),
			wantCount:   2,
			wantFiles:   []string{"app1.yaml", "app2.yaml"},
			description: "[12] should match 1 or 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := Glob(tt.pattern)
			if err != nil {
				t.Fatalf("Glob() error = %v", err)
			}

			if len(matches) != tt.wantCount {
				t.Errorf("%s: Glob(%q) returned %d files, want %d", tt.description, tt.pattern, len(matches), tt.wantCount)
				t.Logf("Matches: %v", matches)
			}

			// Check expected files are present
			matchMap := make(map[string]bool)
			for _, m := range matches {
				matchMap[filepath.Base(m)] = true
			}

			for _, wantFile := range tt.wantFiles {
				if !matchMap[wantFile] {
					t.Errorf("%s: Expected %s in matches", tt.description, wantFile)
				}
			}
		})
	}
}

func TestGlob_EmptyResult(t *testing.T) {
	// Test pattern that matches nothing
	tmpDir := t.TempDir()

	pattern := filepath.Join(tmpDir, "nonexistent", "*.yaml")
	matches, err := Glob(pattern)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Glob(%q) should return empty slice for non-matching pattern, got %d matches", pattern, len(matches))
	}
}

func TestGlob_InvalidPattern(t *testing.T) {
	// Test invalid glob pattern (unclosed bracket)
	pattern := "/tmp/invalid[pattern"
	_, err := Glob(pattern)
	if err == nil {
		t.Error("Glob() should return error for invalid pattern")
	}
}

func TestGlob_HiddenFiles(t *testing.T) {
	// Test that .hidden files can be matched with explicit pattern
	tmpDir := t.TempDir()

	// Create hidden file
	hiddenFile := filepath.Join(tmpDir, ".hidden.yaml")
	if err := os.WriteFile(hiddenFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create regular file
	regularFile := filepath.Join(tmpDir, "regular.yaml")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Pattern with .* should match hidden files explicitly
	pattern := filepath.Join(tmpDir, ".*.yaml")
	matches, err := Glob(pattern)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	// Should match the hidden file
	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}

	if !matchMap[hiddenFile] {
		t.Error("Expected .hidden.yaml to be matched with .*.yaml pattern")
	}

	// Test that regular * pattern matches both files
	// Note: doublestar matches hidden files with * (unlike shell behavior)
	pattern2 := filepath.Join(tmpDir, "*.yaml")
	matches2, err := Glob(pattern2)
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	// Both files should be matched
	if len(matches2) != 2 {
		t.Errorf("Expected 2 files to be matched, got %d", len(matches2))
	}

	matchMap2 := make(map[string]bool)
	for _, m := range matches2 {
		matchMap2[m] = true
	}

	if !matchMap2[hiddenFile] {
		t.Error("hidden file should be matched by *.yaml pattern (doublestar behavior)")
	}

	if !matchMap2[regularFile] {
		t.Error("regular.yaml should be matched by *.yaml pattern")
	}
}

func TestGlobFiles_OnlyReturnsFiles(t *testing.T) {
	// Test that GlobFiles() excludes directories and only returns files
	tmpDir := t.TempDir()

	// Create mixed content: files and directories
	// Structure:
	// ├── file1.yaml (file)
	// ├── file2.yaml (file)
	// ├── dir1/ (directory)
	// └── dir2.yaml/ (directory with .yaml in name)

	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2.yaml")
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2.yaml")

	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	// Test GlobFiles with pattern that would match both files and directories
	pattern := filepath.Join(tmpDir, "*")
	matches, err := GlobFiles(pattern)
	if err != nil {
		t.Fatalf("GlobFiles() error = %v", err)
	}

	// Should only match the 2 files, not the directories
	if len(matches) != 2 {
		t.Errorf("GlobFiles(%q) returned %d items, want 2 (files only)", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}

	if !matchMap[file1] {
		t.Error("Expected file1.yaml to be matched")
	}
	if !matchMap[file2] {
		t.Error("Expected file2.yaml to be matched")
	}
	if matchMap[dir1] {
		t.Error("dir1 should NOT be matched (is a directory)")
	}
	if matchMap[dir2] {
		t.Error("dir2.yaml should NOT be matched (is a directory)")
	}
}

func TestGlobDirs_OnlyReturnsDirectories(t *testing.T) {
	// Test that GlobDirs() excludes files and only returns directories
	tmpDir := t.TempDir()

	// Create mixed content: files and directories
	// Structure:
	// ├── file1.yaml (file)
	// ├── file2 (file with no extension)
	// ├── dir1/ (directory)
	// └── dir2/ (directory)

	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2")
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	// Test GlobDirs with pattern that would match both files and directories
	pattern := filepath.Join(tmpDir, "*")
	matches, err := GlobDirs(pattern)
	if err != nil {
		t.Fatalf("GlobDirs() error = %v", err)
	}

	// Should only match the 2 directories, not the files
	if len(matches) != 2 {
		t.Errorf("GlobDirs(%q) returned %d items, want 2 (directories only)", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
	}

	if matchMap[file1] {
		t.Error("file1.yaml should NOT be matched (is a file)")
	}
	if matchMap[file2] {
		t.Error("file2 should NOT be matched (is a file)")
	}
	if !matchMap[dir1] {
		t.Error("Expected dir1 to be matched")
	}
	if !matchMap[dir2] {
		t.Error("Expected dir2 to be matched")
	}
}

func TestGlobFiles_RecursiveMixedContent(t *testing.T) {
	// Test GlobFiles with recursive ** pattern in mixed content
	tmpDir := t.TempDir()

	// Create nested structure with files and directories at multiple levels
	// Structure:
	// ├── root.yaml (file)
	// ├── rootdir/ (directory)
	// ├── sub/
	// │   ├── sub.yaml (file)
	// │   └── subdir/ (directory)
	// └── deep/
	//     └── nested/
	//         └── deep.yaml (file)

	rootFile := filepath.Join(tmpDir, "root.yaml")
	rootDir := filepath.Join(tmpDir, "rootdir")
	subDir := filepath.Join(tmpDir, "sub")
	subFile := filepath.Join(subDir, "sub.yaml")
	subSubDir := filepath.Join(subDir, "subdir")
	deepDir := filepath.Join(tmpDir, "deep", "nested")
	deepFile := filepath.Join(deepDir, "deep.yaml")

	if err := os.WriteFile(rootFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(subFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subSubDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deepFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test recursive glob - should only return files, not directories
	pattern := filepath.Join(tmpDir, "**", "*.yaml")
	matches, err := GlobFiles(pattern)
	if err != nil {
		t.Fatalf("GlobFiles() error = %v", err)
	}

	// Should match the 3 .yaml files at different depths
	if len(matches) != 3 {
		t.Errorf("GlobFiles(%q) returned %d items, want 3 (files only)", pattern, len(matches))
		t.Logf("Matches: %v", matches)
	}

	matchMap := make(map[string]bool)
	for _, m := range matches {
		matchMap[m] = true
		// Verify all matches are files, not directories
		info, err := os.Stat(m)
		if err != nil {
			t.Errorf("Failed to stat matched path %s: %v", m, err)
		} else if info.IsDir() {
			t.Errorf("GlobFiles returned directory %s, should only return files", m)
		}
	}

	if !matchMap[rootFile] {
		t.Error("Expected root.yaml to be matched")
	}
	if !matchMap[subFile] {
		t.Error("Expected sub/sub.yaml to be matched")
	}
	if !matchMap[deepFile] {
		t.Error("Expected deep/nested/deep.yaml to be matched")
	}
}

func TestValidateGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"valid star pattern", "./charts/*", false},
		{"valid doublestar", "./charts/**/*.yaml", false},
		{"valid brace expansion", "./charts/{app,api}", false},
		{"valid question mark", "./charts/?", false},
		{"valid character class", "./charts/[abc]", false},
		{"unclosed bracket", "./charts/[invalid", true},
		{"unclosed brace", "./charts/{app,api", true},
		{"invalid escape", "./charts/\\", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGlobPattern(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGlobPattern(%q) error = %v, wantErr %v",
					tt.pattern, err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "invalid glob syntax") {
				t.Errorf("ValidateGlobPattern(%q) error message = %q, want to contain 'invalid glob syntax'",
					tt.pattern, err.Error())
			}
		})
	}
}

func TestContainsGlob(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"./charts/*", true},
		{"./charts/**/*.yaml", true},
		{"./charts/{app,api}", true},
		{"./charts/[abc]", true},
		{"./charts/foo?bar", true},
		{"./charts/simple", false},
		{"./charts/simple-path", false},
		{"simple", false},
		{"/absolute/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ContainsGlob(tt.path)
			if got != tt.want {
				t.Errorf("ContainsGlob(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// Defensive validation tests - ensure public API validates patterns even if caller didn't

func TestGlob_DefensiveValidation(t *testing.T) {
	// Test that public Glob() validates pattern even if caller didn't
	// This follows the lint2 pattern where all public functions validate defensively
	invalidPattern := "/tmp/[unclosed"

	_, err := Glob(invalidPattern)
	if err == nil {
		t.Error("Glob() should validate pattern and return error for invalid syntax")
	}

	if !strings.Contains(err.Error(), "invalid glob pattern") {
		t.Errorf("Error should mention invalid pattern, got: %v", err)
	}

	// Should include helpful details in error
	if !strings.Contains(err.Error(), "unclosed brackets") {
		t.Errorf("Error should include helpful details about what might be wrong, got: %v", err)
	}
}

func TestGlobFiles_DefensiveValidation(t *testing.T) {
	// Test that GlobFiles() validates pattern syntax before expansion
	invalidPattern := "/tmp/{unclosed"

	_, err := GlobFiles(invalidPattern)
	if err == nil {
		t.Error("GlobFiles() should validate pattern and return error for invalid syntax")
	}

	if !strings.Contains(err.Error(), "invalid glob pattern") {
		t.Errorf("Error should mention invalid pattern, got: %v", err)
	}
}

func TestGlobDirs_DefensiveValidation(t *testing.T) {
	// Test that GlobDirs() validates pattern syntax before expansion
	invalidPattern := "/tmp/[abc"

	_, err := GlobDirs(invalidPattern)
	if err == nil {
		t.Error("GlobDirs() should validate pattern and return error for invalid syntax")
	}

	if !strings.Contains(err.Error(), "invalid glob pattern") {
		t.Errorf("Error should mention invalid pattern, got: %v", err)
	}
}

func TestGlob_DefensiveValidation_ValidPattern(t *testing.T) {
	// Test that validation doesn't reject valid patterns
	// Using a pattern that won't match any files but is syntactically valid
	tmpDir := t.TempDir()
	validPattern := filepath.Join(tmpDir, "**", "*.nonexistent")

	_, err := Glob(validPattern)
	if err != nil {
		t.Fatalf("Glob() should accept valid pattern, got error: %v", err)
	}

	// No error means validation passed - we're testing validation, not matching
	// (doublestar may return nil or empty slice for no matches, both are fine)
}
