package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DetectedResources holds resources found during auto-detection
type DetectedResources struct {
	Charts         []string
	Preflights     []string
	SupportBundles []string
	Manifests      []string
	ValuesFiles    []string
}

// AutoDetectResources searches the directory tree for Helm charts and preflight specs
func AutoDetectResources(startPath string) (*DetectedResources, error) {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting current directory: %w", err)
		}
	}

	// Make startPath absolute for consistent path resolution
	absStartPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	resources := &DetectedResources{
		Charts:         []string{},
		Preflights:     []string{},
		SupportBundles: []string{},
		Manifests:      []string{},
		ValuesFiles:    []string{},
	}

	// Track directories that might contain manifests
	manifestDirs := make(map[string]bool)

	// Walk the directory tree
	err = filepath.Walk(absStartPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Handle directories
		if info.IsDir() {
			name := info.Name()
			// Skip hidden directories and common ignore patterns
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}

			// Check if this is a manifest directory
			dirName := strings.ToLower(name)
			if dirName == "manifests" || dirName == "replicated" ||
				dirName == "kustomize" || dirName == "k8s" ||
				dirName == "kubernetes" || dirName == "yaml" {
				relPath, err := filepath.Rel(absStartPath, path)
				if err == nil && relPath != "." {
					manifestDirs[relPath] = true
				}
			}

			// Continue walking subdirectories
			return nil
		}

		// Detect Helm charts by Chart.yaml or Chart.yml
		if info.Name() == "Chart.yaml" || info.Name() == "Chart.yml" {
			chartDir := filepath.Dir(path)
			// Make path relative to start path
			relPath, err := filepath.Rel(absStartPath, chartDir)
			if err == nil {
				resources.Charts = append(resources.Charts, relPath)
			}
		}

		// Detect values files
		fileName := strings.ToLower(info.Name())
		if fileName == "values.yaml" || fileName == "values.yml" ||
			strings.HasPrefix(fileName, "values-") && (strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml")) {
			relPath, err := filepath.Rel(absStartPath, path)
			if err == nil {
				resources.ValuesFiles = append(resources.ValuesFiles, relPath)
			}
		}

		// Detect Troubleshoot specs (Preflight and SupportBundle) by parsing YAML
		if strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
			kind, err := getYAMLKind(path)
			if err == nil {
				relPath, err := filepath.Rel(absStartPath, path)
				if err == nil {
					switch kind {
					case "Preflight":
						resources.Preflights = append(resources.Preflights, relPath)
					case "SupportBundle":
						resources.SupportBundles = append(resources.SupportBundles, relPath)
					}
				}
			}
		}

		return nil
	})

	// Convert manifest directories to glob patterns
	for dir := range manifestDirs {
		// Suggest a pattern like "./manifests/**/*.yaml"
		if !strings.HasPrefix(dir, ".") {
			dir = "./" + dir
		}
		pattern := filepath.Join(dir, "**", "*.yaml")
		resources.Manifests = append(resources.Manifests, pattern)
	}

	if err != nil {
		return nil, fmt.Errorf("walking directory tree: %w", err)
	}

	return resources, nil
}

// WriteConfigFile writes a config to a file using flow-style format
func WriteConfigFile(config *Config, path string) error {
	// Ensure the config file path is either .replicated or .replicated.yaml
	if filepath.Base(path) != ".replicated" && filepath.Base(path) != ".replicated.yaml" {
		return fmt.Errorf("config file must be named .replicated or .replicated.yaml")
	}

	// Build YAML content manually to match the example format
	var sb strings.Builder

	// App metadata
	if config.AppId != "" {
		sb.WriteString(fmt.Sprintf("appId: %q\n", config.AppId))
	}
	if config.AppSlug != "" {
		sb.WriteString(fmt.Sprintf("appSlug: %q\n", config.AppSlug))
	}

	// Promotion settings
	if len(config.PromoteToChannelIds) > 0 {
		sb.WriteString("promoteToChannelIds: [")
		for i, id := range config.PromoteToChannelIds {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", id))
		}
		sb.WriteString("]\n")
	}

	if len(config.PromoteToChannelNames) > 0 {
		sb.WriteString("promoteToChannelNames: [")
		for i, name := range config.PromoteToChannelNames {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", name))
		}
		sb.WriteString("]\n")
	}

	// Charts
	if len(config.Charts) > 0 {
		sb.WriteString("charts: [\n")
		for i, chart := range config.Charts {
			sb.WriteString("  {\n")
			sb.WriteString(fmt.Sprintf("    path: %q,\n", chart.Path))
			sb.WriteString(fmt.Sprintf("    chartVersion: %q,\n", chart.ChartVersion))
			sb.WriteString(fmt.Sprintf("    appVersion: %q,\n", chart.AppVersion))
			sb.WriteString("  },\n")
			_ = i
		}
		sb.WriteString("]\n")
	}

	// Preflights
	if len(config.Preflights) > 0 {
		sb.WriteString("preflights: [\n")
		for _, preflight := range config.Preflights {
			sb.WriteString("  {\n")
			sb.WriteString(fmt.Sprintf("    path: %q,\n", preflight.Path))
			sb.WriteString(fmt.Sprintf("    chartName: %q,\n", preflight.ChartName))
			sb.WriteString(fmt.Sprintf("    chartVersion: %q,\n", preflight.ChartVersion))
			sb.WriteString("  },\n")
		}
		sb.WriteString("]\n")
	}

	// Release label
	if config.ReleaseLabel != "" {
		sb.WriteString(fmt.Sprintf("releaseLabel: %q\n", config.ReleaseLabel))
	}

	// Manifests
	if len(config.Manifests) > 0 {
		sb.WriteString("manifests: [")
		for i, manifest := range config.Manifests {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", manifest))
		}
		sb.WriteString("]\n")
	}

	// Linting config
	if config.ReplLint != nil {
		sb.WriteString("repl-lint:\n")
		sb.WriteString(fmt.Sprintf("  version: %d\n", config.ReplLint.Version))
		sb.WriteString("  linters:\n")

		writeLintersConfig := func(name string, linter LinterConfig) {
			disabled := false
			if linter.Disabled != nil {
				disabled = *linter.Disabled
			}
			sb.WriteString(fmt.Sprintf("    %s:\n", name))
			sb.WriteString(fmt.Sprintf("      disabled: %t\n", disabled))
		}

		writeLintersConfig("helm", config.ReplLint.Linters.Helm)
		writeLintersConfig("preflight", config.ReplLint.Linters.Preflight)
		writeLintersConfig("support-bundle", config.ReplLint.Linters.SupportBundle)
	}

	// Write to file
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// IsNonInteractive checks if we're running in a non-interactive environment
func IsNonInteractive() bool {
	// Check CI environment variables
	if os.Getenv("CI") != "" {
		return true
	}

	// Check if stdin is not a terminal (piped input)
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return true // Assume non-interactive on error
	}

	// If not a character device, it's not interactive
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
}

// ConfigExists checks if a .replicated config file exists in the current directory or parents
func ConfigExists(startPath string) (bool, string, error) {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return false, "", fmt.Errorf("getting current directory: %w", err)
		}
	}

	// Make absolute
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return false, "", fmt.Errorf("resolving absolute path: %w", err)
	}

	currentDir := absPath

	for {
		// Try .replicated first, then .replicated.yaml
		candidates := []string{
			filepath.Join(currentDir, ".replicated"),
			filepath.Join(currentDir, ".replicated.yaml"),
		}

		for _, configPath := range candidates {
			if stat, err := os.Stat(configPath); err == nil && !stat.IsDir() {
				return true, configPath, nil
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root
			break
		}
		currentDir = parentDir
	}

	return false, "", nil
}

// getYAMLKind reads a YAML file and returns its kind field
func getYAMLKind(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var doc struct {
		Kind string `yaml:"kind"`
	}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", err
	}

	return doc.Kind, nil
}

// GetYAMLAPIVersion reads a YAML file and returns its apiVersion field
func GetYAMLAPIVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var doc struct {
		APIVersion string `yaml:"apiVersion"`
	}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", err
	}

	return doc.APIVersion, nil
}
