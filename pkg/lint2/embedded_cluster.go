package lint2

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ExpandManifestGlobs expands a list of manifest glob patterns to the set of
// matching YAML files, applying the same gitignore and hidden-path filtering
// used by the other linters. All YAML files are returned (no kind filtering).
func ExpandManifestGlobs(manifestPatterns []string) ([]string, error) {
	gitignoreChecker, _ := NewGitignoreChecker(".")

	seen := make(map[string]bool)
	var files []string

	for _, pattern := range manifestPatterns {
		cleanPattern := filepath.Clean(pattern)
		skipHidden := !patternTargetsHiddenPath(cleanPattern)

		var checker *GitignoreChecker
		if gitignoreChecker != nil && !gitignoreChecker.PathMatchesIgnoredPattern(cleanPattern) {
			checker = gitignoreChecker
		}

		yamlPatterns, err := buildYAMLPatterns(cleanPattern)
		if err != nil {
			// pattern doesn't end in a recognized suffix; skip it
			continue
		}

		for _, p := range yamlPatterns {
			var matches []string
			if checker != nil {
				matches, _ = GlobFiles(p, WithGitignoreChecker(checker))
			} else {
				matches, _ = GlobFiles(p)
			}

			for _, m := range matches {
				if skipHidden && isHiddenPath(m) {
					continue
				}
				if !seen[m] {
					seen[m] = true
					files = append(files, m)
				}
			}
		}
	}

	return files, nil
}

// ecLintIssue is an issue from the EC CLI lint output. It implements LintIssue.
type ecLintIssue struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

func (i ecLintIssue) GetLine() int       { return i.Line }
func (i ecLintIssue) GetColumn() int     { return i.Column }
func (i ecLintIssue) GetMessage() string { return i.Message }
func (i ecLintIssue) GetField() string   { return i.Field }

const ecConfigAPIVersion = "embeddedcluster.replicated.com/v1beta1"

// ecConfigManifest is the minimal structure needed to identify and read an EC Config manifest.
type ecConfigManifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		Version string `yaml:"version"`
	} `yaml:"spec"`
}

// DiscoverECVersion walks the given paths (files or directories) looking for a
// manifest with apiVersion embeddedcluster.replicated.com/v1beta1 and kind Config,
// and returns its spec.version. Returns an error if no manifest is found.
func DiscoverECVersion(paths []string) (string, error) {
	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			continue
		}

		var yamlFiles []string
		if info.IsDir() {
			_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".yaml" || ext == ".yml" {
					yamlFiles = append(yamlFiles, path)
				}
				return nil
			})
		} else {
			yamlFiles = []string{root}
		}

		for _, f := range yamlFiles {
			version, err := parseECVersionFromFile(f)
			if err != nil || version == "" {
				continue
			}
			return version, nil
		}
	}
	return "", fmt.Errorf("no embedded-cluster Config manifest found in paths: %s", strings.Join(paths, ", "))
}

// parseECVersionFromFile reads a YAML file and returns the spec.version if it is
// an embedded-cluster Config manifest; returns an empty string otherwise.
func parseECVersionFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Iterate through all YAML documents in the file (separated by ---).
	// yaml.Unmarshal only reads the first document, so a multi-document file
	// would silently miss an EC Config manifest in a non-first document.
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var manifest ecConfigManifest
		if err := decoder.Decode(&manifest); err != nil {
			break // io.EOF or parse error; stop iterating
		}
		if manifest.Kind == "Config" && manifest.APIVersion == ecConfigAPIVersion {
			return manifest.Spec.Version, nil
		}
	}
	return "", nil
}

// LintEmbeddedCluster runs `ec lint --format json <paths...>` and returns structured results.
// The binary must already be available at ecBinaryPath. disableChecks is passed as --disable
// to the EC CLI; if empty the caller should supply the defaults.
func LintEmbeddedCluster(ctx context.Context, paths []string, ecBinaryPath string, disableChecks []string) (*LintResult, error) {
	if len(paths) == 0 {
		return &LintResult{Success: true, Messages: []LintMessage{}}, nil
	}

	args := []string{"lint", "--format", "json"}
	if len(disableChecks) > 0 {
		args = append(args, "--disable", strings.Join(disableChecks, ","))
	}
	args = append(args, paths...)
	cmd := exec.CommandContext(ctx, ecBinaryPath, args...)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// EC lint exits non-zero when lint issues are found; we still parse the output.
	// parseLintJSON handles stderr mixed into CombinedOutput and trailing non-JSON content.
	parsed, jsonErr := parseLintJSON[ecLintIssue](outputStr)
	if jsonErr != nil {
		if err != nil {
			return nil, fmt.Errorf("ec lint failed: %w\nOutput: %s", err, outputStr)
		}
		return nil, fmt.Errorf("failed to parse ec lint output: %w\nOutput: %s", jsonErr, outputStr)
	}

	messages := convertLintOutputToMessages(parsed)

	// Success = no ERROR-severity messages
	success := true
	for _, msg := range messages {
		if msg.Severity == "ERROR" {
			success = false
			break
		}
	}

	return &LintResult{
		Success:  success,
		Messages: messages,
	}, nil
}
