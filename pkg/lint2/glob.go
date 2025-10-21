package lint2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Glob expands glob patterns using doublestar library, which supports:
// - * : matches any sequence of non-separator characters
// - ** : matches zero or more directories (recursive)
// - ? : matches any single character
// - [abc] : matches any character in the brackets
// - {alt1,alt2} : matches any of the alternatives
//
// This is a wrapper around doublestar.FilepathGlob that provides:
// - Drop-in replacement for filepath.Glob
// - Recursive ** globbing (unlike stdlib filepath.Glob)
// - Brace expansion {a,b,c}
func Glob(pattern string) ([]string, error) {
	// Defensive check: validate pattern syntax
	// Note: patterns are validated during config parsing, but we check again here
	// since Glob is a public function that could be called directly
	if err := ValidateGlobPattern(pattern); err != nil {
		return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
	}

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("expanding glob pattern %s: %w", pattern, err)
	}
	return matches, nil
}

// GlobFiles expands glob patterns returning only files (not directories).
// Uses WithFilesOnly() option for efficient library-level filtering.
// This is useful for preflight specs and manifest discovery where only
// files should be processed.
func GlobFiles(pattern string) ([]string, error) {
	// Defensive check: validate pattern syntax
	// Note: patterns are validated during config parsing, but we check again here
	// since GlobFiles is a public function that could be called directly
	if err := ValidateGlobPattern(pattern); err != nil {
		return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
	}

	matches, err := doublestar.FilepathGlob(pattern, doublestar.WithFilesOnly())
	if err != nil {
		return nil, fmt.Errorf("expanding glob pattern %s: %w", pattern, err)
	}
	return matches, nil
}

// ValidateGlobPattern checks if a pattern is valid doublestar glob syntax and
// does not contain path traversal attempts.
// This is useful for validating user input early before attempting to expand patterns.
// Returns an error if the pattern syntax is invalid or attempts path traversal.
func ValidateGlobPattern(pattern string) error {
	// Check glob syntax
	if !doublestar.ValidatePattern(pattern) {
		return fmt.Errorf("invalid glob syntax (check for unclosed brackets, braces, or invalid escape sequences)")
	}

	// Security: prevent path traversal outside repository
	// Clean the pattern to normalize .. sequences
	cleanPath := filepath.Clean(pattern)

	// If the cleaned path starts with .., it's trying to escape
	// Note: We allow relative paths within the repo (./foo, foo/bar)
	// but reject paths that go up and out (../../../etc)
	if strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || cleanPath == ".." {
		return fmt.Errorf("pattern cannot traverse outside repository (contains path traversal)")
	}

	return nil
}

// ContainsGlob checks if a path contains glob wildcards (* ? [ {).
// Exported for use by config parsing to detect patterns that need validation.
func ContainsGlob(path string) bool {
	return strings.ContainsAny(path, "*?[{")
}
