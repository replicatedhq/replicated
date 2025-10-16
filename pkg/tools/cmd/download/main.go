package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/replicatedhq/replicated/pkg/tools"
)

type downloadResult struct {
	tool    string
	version string
	success bool
	err     error
}

func main() {
	// Parse arguments - which tools to download
	requestedTools := parseArgs(os.Args[1:])

	if len(requestedTools) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: go run main.go <tool1> <tool2> ...\n")
		fmt.Fprintf(os.Stderr, "Available tools: helm, preflight, support-bundle\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  go run main.go helm\n")
		fmt.Fprintf(os.Stderr, "  go run main.go helm preflight\n")
		fmt.Fprintf(os.Stderr, "  go run main.go helm preflight support-bundle\n")
		os.Exit(1)
	}

	// Parse .replicated config to get versions
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  No .replicated config file found in current directory or parent directories\n")
		fmt.Fprintf(os.Stderr, "Cannot determine tool versions without config. Skipping download.\n")
		os.Exit(1)
	}

	toolVersions := tools.GetToolVersions(config)

	fmt.Printf("Detected platform: %s-%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Downloading tools: %v\n", requestedTools)
	fmt.Println()

	// Use the Downloader to download to cache
	downloader := tools.NewDownloader()
	ctx := context.Background()

	// Track results
	var results []downloadResult

	// Download requested tools
	for _, toolName := range requestedTools {
		version := toolVersions[toolName]
		if version == "" {
			fmt.Printf("⚠️  No version found for tool %s in config\n", toolName)
			results = append(results, downloadResult{toolName, "", false, fmt.Errorf("no version in config")})
			continue
		}

		fmt.Printf("Downloading %s %s...\n", toolName, version)
		if err := downloader.Download(ctx, toolName, version); err != nil {
			fmt.Printf("⚠️  Version %s not found or failed to download: %v\n", version, err)
			results = append(results, downloadResult{toolName, version, false, err})
		} else {
			results = append(results, downloadResult{toolName, version, true, nil})
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("Download Summary:")
	successCount := 0
	for _, r := range results {
		if r.success {
			fmt.Printf("  ✓ %s %s - success\n", r.tool, r.version)
			successCount++
		} else {
			fmt.Printf("  ✗ %s %s - failed: %v\n", r.tool, r.version, r.err)
		}
	}

	if successCount == len(results) {
		cacheDir, _ := tools.GetCacheDir()
		fmt.Printf("\n✅ All %d tools downloaded successfully to %s\n", successCount, cacheDir)
		fmt.Println("\nCached tools:")
		showCachedTools(cacheDir)
		os.Exit(0)
	} else {
		fmt.Printf("\n⚠️  Downloaded %d/%d tools\n", successCount, len(results))
		if successCount > 0 {
			cacheDir, _ := tools.GetCacheDir()
			fmt.Println("\nSuccessfully downloaded:")
			showCachedTools(cacheDir)
		}
		os.Exit(1)
	}
}

func showCachedTools(cacheDir string) {
	filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() != ".DS_Store" {
			sizeMB := float64(info.Size()) / 1024 / 1024
			fmt.Printf("  %s (%.0fM)\n", path, sizeMB)
		}
		return nil
	})
}

func parseArgs(args []string) []string {
	var tools []string
	validTools := map[string]bool{
		"helm":           true,
		"preflight":      true,
		"support-bundle": true,
	}

	for _, arg := range args {
		if validTools[arg] {
			tools = append(tools, arg)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: unknown tool %q (valid: helm, preflight, support-bundle)\n", arg)
		}
	}

	return tools
}
