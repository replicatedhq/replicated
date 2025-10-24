package lint2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/monochromegane/go-gitignore"
)

// GitignoreChecker checks if paths should be ignored based on .gitignore rules.
// It supports repository .gitignore files, .git/info/exclude, and global gitignore.
type GitignoreChecker struct {
	baseDir         string
	repoRoot        string
	matchers        []gitignore.IgnoreMatcher
	ignoredPatterns []string
}

// NewGitignoreChecker creates a new GitignoreChecker for the given base directory.
// It loads all gitignore sources (repository .gitignore files, .git/info/exclude, and global gitignore).
// Returns nil, nil if no gitignore files are found (not an error).
// Errors are only returned for critical failures like invalid baseDir.
func NewGitignoreChecker(baseDir string) (*GitignoreChecker, error) {
	// Validate baseDir exists and is a directory
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("invalid base directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base directory is not a directory: %s", baseDir)
	}

	// Convert to absolute path
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	// Find repository root (look for .git directory)
	repoRoot, err := findGitRoot(absBaseDir)
	if err != nil {
		// Not a git repository - return nil (no gitignore checking)
		return nil, nil
	}

	// Collect all gitignore matchers and patterns
	var matchers []gitignore.IgnoreMatcher
	var allPatterns []string

	// Load repository .gitignore files (from root down to baseDir)
	repoMatchers, repoPatterns, err := loadRepositoryGitignoreMatchers(repoRoot, absBaseDir)
	if err != nil {
		// Log but don't fail - gitignore is optional
		_ = err
	} else {
		matchers = append(matchers, repoMatchers...)
		allPatterns = append(allPatterns, repoPatterns...)
	}

	// Load .git/info/exclude
	excludeMatcher, excludePatterns, err := loadGitInfoExcludeMatcher(repoRoot)
	if err != nil {
		// Log but don't fail
		_ = err
	} else if excludeMatcher != nil {
		matchers = append(matchers, excludeMatcher)
		allPatterns = append(allPatterns, excludePatterns...)
	}

	// NOTE: Global gitignore support disabled for security concerns
	// The loadGlobalGitignoreMatcher() function had critical security issues:
	//   - Arbitrary command execution (exec.Command with git)
	//   - Path traversal vulnerabilities (tilde expansion without validation)
	// Will re-enable in future PR with proper security measures.
	// See: https://github.com/replicatedhq/replicated/pull/634

	// If no matchers found, return nil (no gitignore checking needed)
	if len(matchers) == 0 {
		return nil, nil
	}

	return &GitignoreChecker{
		baseDir:         absBaseDir,
		repoRoot:        repoRoot,
		matchers:        matchers,
		ignoredPatterns: allPatterns,
	}, nil
}

// findGitRoot walks up the directory tree to find the .git directory.
// Returns the repository root directory or an error if not found.
func findGitRoot(startDir string) (string, error) {
	currentDir := startDir

	for {
		gitDir := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return currentDir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root without finding .git
			return "", fmt.Errorf("not a git repository")
		}
		currentDir = parentDir
	}
}

// loadRepositoryGitignoreMatchers loads all .gitignore files from repo root down to the target directory.
// Returns matchers and patterns for each .gitignore file found.
func loadRepositoryGitignoreMatchers(repoRoot, targetDir string) ([]gitignore.IgnoreMatcher, []string, error) {
	var matchers []gitignore.IgnoreMatcher
	var allPatterns []string

	// Walk from repo root to target directory, collecting .gitignore files
	currentDir := repoRoot
	for {
		gitignorePath := filepath.Join(currentDir, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			// File exists, create a matcher from it
			matcher, err := gitignore.NewGitIgnore(gitignorePath)
			if err == nil {
				matchers = append(matchers, matcher)
				// Also read patterns for bypass checking
				if patterns, err := readGitignoreFile(gitignorePath); err == nil {
					allPatterns = append(allPatterns, patterns...)
				}
			}
		}

		// If we've reached the target directory, stop
		if currentDir == targetDir {
			break
		}

		// If target is not under current dir, we're done
		relPath, err := filepath.Rel(currentDir, targetDir)
		if err != nil || strings.HasPrefix(relPath, "..") {
			break
		}

		// Move to next subdirectory towards target
		nextDir := filepath.Join(currentDir, strings.Split(relPath, string(filepath.Separator))[0])
		if nextDir == currentDir {
			break
		}
		currentDir = nextDir
	}

	return matchers, allPatterns, nil
}

// loadGitInfoExcludeMatcher loads matcher from .git/info/exclude
func loadGitInfoExcludeMatcher(repoRoot string) (gitignore.IgnoreMatcher, []string, error) {
	excludePath := filepath.Join(repoRoot, ".git", "info", "exclude")
	if _, err := os.Stat(excludePath); err != nil {
		// File doesn't exist
		return nil, nil, nil
	}

	matcher, err := gitignore.NewGitIgnore(excludePath)
	if err != nil {
		return nil, nil, err
	}

	patterns, err := readGitignoreFile(excludePath)
	if err != nil {
		return matcher, nil, nil
	}

	return matcher, patterns, nil
}

// loadGlobalGitignoreMatcher loads matcher from the global gitignore file
// DISABLED: Security concerns with exec.Command - command execution and path traversal risks
// TODO: Re-enable with proper security in future PR:
//   - Parse ~/.gitconfig directly (no command execution)
//   - Validate paths to prevent traversal attacks
//   - Add comprehensive input validation
// See: https://github.com/replicatedhq/replicated/pull/634
func loadGlobalGitignoreMatcher() (gitignore.IgnoreMatcher, []string, error) {
	// Disabled for security - will re-enable with proper validation in future PR
	return nil, nil, nil
}

// readGitignoreFile reads and parses a gitignore file
func readGitignoreFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not an error - file just doesn't exist
		}
		return nil, err
	}

	var patterns []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	return patterns, nil
}

// ShouldIgnore checks if a path should be ignored based on gitignore rules.
// Returns true if the path matches any gitignore pattern.
// Path can be absolute or relative to the base directory.
func (g *GitignoreChecker) ShouldIgnore(path string) bool {
	if g == nil || len(g.matchers) == 0 {
		return false
	}

	// Convert to absolute path if relative
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(g.baseDir, path)
	}

	// Check if path is within repository
	relPath, err := filepath.Rel(g.repoRoot, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		// Path is outside repository - don't ignore
		return false
	}

	// Check if path or any parent directory is ignored
	// We need to check if it's a directory for proper matching
	isDir := false
	if info, err := os.Stat(absPath); err == nil {
		isDir = info.IsDir()
	}

	// Check all matchers - if any match, the path is ignored
	// The go-gitignore library expects absolute paths
	for _, matcher := range g.matchers {
		if matcher.Match(absPath, isDir) {
			return true
		}
	}

	// Also check if any parent directory is ignored
	// This handles cases where a file is inside an ignored directory
	currentPath := filepath.Dir(absPath)
	for currentPath != g.repoRoot && len(currentPath) > len(g.repoRoot) {
		// Check if this parent directory is ignored
		if info, err := os.Stat(currentPath); err == nil && info.IsDir() {
			for _, matcher := range g.matchers {
				if matcher.Match(currentPath, true) {
					return true
				}
			}
		}

		// Move up to parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break
		}
		currentPath = parentPath
	}

	return false
}

// PathMatchesIgnoredPattern checks if a config path explicitly references
// a gitignored pattern and should bypass gitignore checking.
//
// This implements the "explicit bypass" rule: if a user specifies a path
// that contains a gitignored pattern (e.g., "./vendor/**" when "vendor/"
// is gitignored), we assume they want to lint that path anyway.
//
// Examples:
//   - "./vendor/**" when "vendor/" is gitignored -> true (bypass)
//   - "./dist/chart" when "dist/" is gitignored -> true (bypass)
//   - "./charts/**" when "vendor/" is gitignored -> false (respect gitignore)
func (g *GitignoreChecker) PathMatchesIgnoredPattern(configPath string) bool {
	if g == nil || len(g.ignoredPatterns) == 0 {
		return false
	}

	// Normalize the config path
	cleanPath := filepath.Clean(configPath)
	cleanPath = filepath.ToSlash(cleanPath)

	// Remove leading "./" for comparison
	cleanPath = strings.TrimPrefix(cleanPath, "./")

	// Extract directory components from the config path
	// We want to check if any part of the path matches a gitignored pattern
	pathParts := strings.Split(cleanPath, "/")

	for _, pattern := range g.ignoredPatterns {
		// Normalize pattern
		pattern = strings.TrimSpace(pattern)
		pattern = filepath.ToSlash(pattern)

		// Skip negation patterns (they don't indicate explicit bypass)
		if strings.HasPrefix(pattern, "!") {
			continue
		}

		// Remove trailing slash from pattern for comparison
		pattern = strings.TrimSuffix(pattern, "/")

		// Check if any path component exactly matches the pattern
		for i, part := range pathParts {
			// Skip glob wildcards in path components
			if strings.ContainsAny(part, "*?[") {
				continue
			}

			// Check for exact directory match
			if part == pattern {
				return true
			}

			// Check if pattern matches the path prefix
			// (e.g., "vendor" matches "./vendor/chart")
			pathPrefix := strings.Join(pathParts[:i+1], "/")
			if pathPrefix == pattern {
				return true
			}
		}

		// Also check if the whole cleaned path starts with the pattern
		if strings.HasPrefix(cleanPath, pattern+"/") || cleanPath == pattern {
			return true
		}

		// Handle wildcard patterns (e.g., "*.log")
		// If the config path explicitly includes a wildcard pattern, that's intentional
		if strings.ContainsAny(pattern, "*?[") {
			// Simple check: does the config path contain the non-wildcard part?
			patternBase := strings.TrimSuffix(pattern, "*")
			if patternBase != "" && strings.Contains(cleanPath, patternBase) {
				return true
			}
		}
	}

	return false
}
