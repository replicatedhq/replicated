package lint2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitignoreChecker_WithValidDirectory(t *testing.T) {
	// Create temp directory with .git and .gitignore
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "vendor/\n*.log\nnode_modules/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)
	assert.Equal(t, tmpDir, checker.baseDir)
	assert.NotEmpty(t, checker.matchers)
	assert.NotEmpty(t, checker.ignoredPatterns)
}

func TestNewGitignoreChecker_NoGitignore(t *testing.T) {
	// Create temp directory with .git but no .gitignore
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	assert.Nil(t, checker) // Should return nil when no gitignore files found
}

func TestNewGitignoreChecker_NotGitRepository(t *testing.T) {
	// Create temp directory without .git
	tmpDir := t.TempDir()

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	assert.Nil(t, checker) // Should return nil when not a git repository
}

func TestNewGitignoreChecker_InvalidDirectory(t *testing.T) {
	checker, err := NewGitignoreChecker("/nonexistent/directory")
	require.Error(t, err)
	assert.Nil(t, checker)
	assert.Contains(t, err.Error(), "invalid base directory")
}

func TestNewGitignoreChecker_NestedGitignore(t *testing.T) {
	// Create directory structure with nested .gitignore files
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	// Root .gitignore
	rootGitignore := "vendor/\n*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(rootGitignore), 0644))

	// Subdir .gitignore
	subGitignore := "temp/\n*.tmp\n"
	require.NoError(t, os.WriteFile(filepath.Join(subDir, ".gitignore"), []byte(subGitignore), 0644))

	checker, err := NewGitignoreChecker(subDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Should have patterns from both gitignore files
	assert.Contains(t, checker.ignoredPatterns, "vendor/")
	assert.Contains(t, checker.ignoredPatterns, "*.log")
	assert.Contains(t, checker.ignoredPatterns, "temp/")
	assert.Contains(t, checker.ignoredPatterns, "*.tmp")
}

func TestShouldIgnore_IgnoredFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create a .log file
	logFile := filepath.Join(tmpDir, "test.log")
	require.NoError(t, os.WriteFile(logFile, []byte("log"), 0644))

	assert.True(t, checker.ShouldIgnore(logFile))
	assert.True(t, checker.ShouldIgnore("test.log")) // Relative path
}

func TestShouldIgnore_NonIgnoredFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create a .txt file
	txtFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(txtFile, []byte("text"), 0644))

	assert.False(t, checker.ShouldIgnore(txtFile))
	assert.False(t, checker.ShouldIgnore("test.txt")) // Relative path
}

func TestShouldIgnore_IgnoredDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "node_modules/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create node_modules directory and file inside it
	nodeModules := filepath.Join(tmpDir, "node_modules")
	require.NoError(t, os.Mkdir(nodeModules, 0755))
	packageFile := filepath.Join(nodeModules, "package.json")
	require.NoError(t, os.WriteFile(packageFile, []byte("{}"), 0644))

	assert.True(t, checker.ShouldIgnore(nodeModules))
	assert.True(t, checker.ShouldIgnore(packageFile))
	assert.True(t, checker.ShouldIgnore("node_modules/package.json")) // Relative path
}

func TestShouldIgnore_NegationPattern(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "*.log\n!important.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create log files
	normalLog := filepath.Join(tmpDir, "test.log")
	require.NoError(t, os.WriteFile(normalLog, []byte("log"), 0644))
	importantLog := filepath.Join(tmpDir, "important.log")
	require.NoError(t, os.WriteFile(importantLog, []byte("important"), 0644))

	assert.True(t, checker.ShouldIgnore(normalLog))
	assert.False(t, checker.ShouldIgnore(importantLog)) // Negated pattern
}

func TestShouldIgnore_NilChecker(t *testing.T) {
	var checker *GitignoreChecker
	assert.False(t, checker.ShouldIgnore("any/path"))
}

func TestPathMatchesIgnoredPattern_ExplicitVendorPath(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "vendor/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	assert.True(t, checker.PathMatchesIgnoredPattern("./vendor/**"))
	assert.True(t, checker.PathMatchesIgnoredPattern("vendor/**"))
	assert.True(t, checker.PathMatchesIgnoredPattern("./vendor/chart"))
}

func TestPathMatchesIgnoredPattern_NonMatchingPath(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "vendor/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	assert.False(t, checker.PathMatchesIgnoredPattern("./charts/**"))
	assert.False(t, checker.PathMatchesIgnoredPattern("./app/**"))
}

func TestPathMatchesIgnoredPattern_LiteralIgnoredPath(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "dist/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	assert.True(t, checker.PathMatchesIgnoredPattern("./dist/output.txt"))
	assert.True(t, checker.PathMatchesIgnoredPattern("dist/helm-chart"))
}

func TestPathMatchesIgnoredPattern_FilenameSimilarity(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "vendor/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// "vendor" in filename should NOT match "vendor/" directory pattern
	assert.False(t, checker.PathMatchesIgnoredPattern("./app/vendor.txt"))
	assert.False(t, checker.PathMatchesIgnoredPattern("./charts/vendored/**"))
}

func TestPathMatchesIgnoredPattern_MultiplePatterns(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "vendor/\ndist/\nnode_modules/\n*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	assert.True(t, checker.PathMatchesIgnoredPattern("./vendor/**"))
	assert.True(t, checker.PathMatchesIgnoredPattern("./dist/charts"))
	assert.True(t, checker.PathMatchesIgnoredPattern("./node_modules/package"))
	assert.False(t, checker.PathMatchesIgnoredPattern("./charts/**"))
}

func TestPathMatchesIgnoredPattern_NilChecker(t *testing.T) {
	var checker *GitignoreChecker
	assert.False(t, checker.PathMatchesIgnoredPattern("any/path"))
}

func TestFindGitRoot_ValidRepository(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	subDir := filepath.Join(tmpDir, "sub", "dir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	root, err := findGitRoot(subDir)
	require.NoError(t, err)
	assert.Equal(t, tmpDir, root)
}

func TestFindGitRoot_NotRepository(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := findGitRoot(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestReadGitignoreFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	content := "vendor/\n# This is a comment\n*.log\n\ndist/\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(content), 0644))

	patterns, err := readGitignoreFile(gitignorePath)
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/", "*.log", "dist/"}, patterns)
}

func TestReadGitignoreFile_NonexistentFile(t *testing.T) {
	patterns, err := readGitignoreFile("/nonexistent/file")
	require.NoError(t, err) // Should not error on missing file
	assert.Nil(t, patterns)
}

func TestReadGitignoreFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte(""), 0644))

	patterns, err := readGitignoreFile(gitignorePath)
	require.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestReadGitignoreFile_OnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	content := "# Comment 1\n# Comment 2\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(content), 0644))

	patterns, err := readGitignoreFile(gitignorePath)
	require.NoError(t, err)
	assert.Empty(t, patterns)
}

// Benchmarks

func BenchmarkNewGitignoreChecker(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	os.Mkdir(gitDir, 0755)

	gitignoreContent := "vendor/\nnode_modules/\n*.log\n*.tmp\ndist/\nbuild/\n"
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewGitignoreChecker(tmpDir)
	}
}

func BenchmarkShouldIgnore_File(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	os.Mkdir(gitDir, 0755)

	gitignoreContent := "*.log\n"
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)

	checker, _ := NewGitignoreChecker(tmpDir)
	testFile := filepath.Join(tmpDir, "test.log")
	os.WriteFile(testFile, []byte("data"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.ShouldIgnore(testFile)
	}
}

func BenchmarkShouldIgnore_Directory(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	os.Mkdir(gitDir, 0755)

	gitignoreContent := "node_modules/\n"
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)

	checker, _ := NewGitignoreChecker(tmpDir)
	nodeModules := filepath.Join(tmpDir, "node_modules")
	os.Mkdir(nodeModules, 0755)
	testFile := filepath.Join(nodeModules, "package.json")
	os.WriteFile(testFile, []byte("{}"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.ShouldIgnore(testFile)
	}
}

func BenchmarkPathMatchesIgnoredPattern(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	os.Mkdir(gitDir, 0755)

	gitignoreContent := "vendor/\nnode_modules/\ndist/\n"
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)

	checker, _ := NewGitignoreChecker(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.PathMatchesIgnoredPattern("./vendor/**")
	}
}

// Edge case tests

func TestGitignoreChecker_EmptyGitignoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	// Create empty .gitignore
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(""), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	// Note: Library creates matcher even for empty file, but ignoredPatterns is empty
	// This is fine - ShouldIgnore will return false for everything
	if checker != nil {
		assert.Empty(t, checker.ignoredPatterns)
		// Verify it doesn't ignore anything
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("data"), 0644)
		assert.False(t, checker.ShouldIgnore(testFile))
	}
}

func TestGitignoreChecker_GitignoreWithOnlyCommentsAndWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "# Just comments\n\n  \n# More comments\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	// Library may create matcher, but ignoredPatterns should be empty
	if checker != nil {
		assert.Empty(t, checker.ignoredPatterns)
	}
}

func TestGitignoreChecker_PathsWithSpaces(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	gitignoreContent := "my dir/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create directory with spaces
	dirWithSpaces := filepath.Join(tmpDir, "my dir")
	require.NoError(t, os.Mkdir(dirWithSpaces, 0755))
	testFile := filepath.Join(dirWithSpaces, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("data"), 0644))

	assert.True(t, checker.ShouldIgnore(testFile))
	assert.True(t, checker.PathMatchesIgnoredPattern("./my dir/**"))
}

func TestGitignoreChecker_PathsWithSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	// Use a pattern with hyphens and underscores (common special characters)
	gitignoreContent := "test-dir_name/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create directory
	specialDir := filepath.Join(tmpDir, "test-dir_name")
	require.NoError(t, os.Mkdir(specialDir, 0755))
	testFile := filepath.Join(specialDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("data"), 0644))

	assert.True(t, checker.ShouldIgnore(testFile))
	assert.True(t, checker.PathMatchesIgnoredPattern("./test-dir_name/**"))
}

func TestGitignoreChecker_NotGitRepository(t *testing.T) {
	// Already tested in TestNewGitignoreChecker_NotGitRepository
	// This verifies the behavior is correct
	tmpDir := t.TempDir()

	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	assert.Nil(t, checker) // Not a git repo = nil checker

	// Should handle nil checker gracefully
	assert.False(t, checker.ShouldIgnore("any/path"))
	assert.False(t, checker.PathMatchesIgnoredPattern("any/pattern"))
}

func TestGitignoreChecker_LargeGitignoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	// Create a large .gitignore with many patterns
	var gitignoreContent strings.Builder
	for i := 0; i < 1000; i++ {
		gitignoreContent.WriteString(fmt.Sprintf("pattern%d/\n", i))
	}
	gitignoreContent.WriteString("*.log\n")

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent.String()), 0644))

	// Should handle large files without issues
	checker, err := NewGitignoreChecker(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, checker)

	// Create a .log file
	logFile := filepath.Join(tmpDir, "test.log")
	require.NoError(t, os.WriteFile(logFile, []byte("data"), 0644))

	assert.True(t, checker.ShouldIgnore(logFile))
}
