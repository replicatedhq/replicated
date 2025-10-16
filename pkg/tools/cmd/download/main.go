package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/replicatedhq/replicated/pkg/tools"
)

type downloadResult struct {
	tool    string
	version string
	success bool
	err     error
}

func main() {
	platformOS := runtime.GOOS
	platformArch := runtime.GOARCH

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

	fmt.Printf("Detected platform: %s-%s\n", platformOS, platformArch)
	fmt.Printf("Downloading tools: %v\n", requestedTools)
	fmt.Println()

	toolsDir := "pkg/tools/embedded"

	// Track results
	var results []downloadResult

	// Download requested tools with versions from config
	if shouldDownload("helm", requestedTools) {
		version := toolVersions[tools.ToolHelm]
		fmt.Printf("Using helm version from config: %s\n", version)
		if err := downloadHelm(version, platformOS, platformArch, toolsDir); err != nil {
			fmt.Printf("⚠️  Version %s not found or failed to download\n", version)
			results = append(results, downloadResult{"helm", version, false, err})
		} else {
			results = append(results, downloadResult{"helm", version, true, nil})
		}
	}

	if shouldDownload("preflight", requestedTools) {
		version := toolVersions[tools.ToolPreflight]
		fmt.Printf("Using preflight version from config: %s\n", version)
		if err := downloadPreflight(version, platformOS, platformArch, toolsDir); err != nil {
			fmt.Printf("⚠️  Version %s not found or failed to download\n", version)
			results = append(results, downloadResult{"preflight", version, false, err})
		} else {
			results = append(results, downloadResult{"preflight", version, true, nil})
		}
	}

	if shouldDownload("support-bundle", requestedTools) {
		version := toolVersions[tools.ToolSupportBundle]
		fmt.Printf("Using support-bundle version from config: %s\n", version)
		if err := downloadSupportBundle(version, platformOS, platformArch, toolsDir); err != nil {
			fmt.Printf("⚠️  Version %s not found or failed to download\n", version)
			results = append(results, downloadResult{"support-bundle", version, false, err})
		} else {
			results = append(results, downloadResult{"support-bundle", version, true, nil})
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
		fmt.Printf("\n✅ All %d tools downloaded successfully to %s\n", successCount, toolsDir)
		fmt.Println("\nDirectory structure:")
		showDownloadedFiles(toolsDir)
		os.Exit(0)
	} else {
		fmt.Printf("\n⚠️  Downloaded %d/%d tools to %s\n", successCount, len(results), toolsDir)
		if successCount > 0 {
			fmt.Println("\nSuccessfully downloaded:")
			showDownloadedFiles(toolsDir)
		}
		os.Exit(1)
	}
}

func showDownloadedFiles(toolsDir string) {
	filepath.Walk(toolsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			// Skip .DS_Store and other system files
			if info.Name() == ".DS_Store" {
				return nil
			}
			sizeMB := float64(info.Size()) / 1024 / 1024
			fmt.Printf("  %s (%.0fM)\n", path, sizeMB)
		}
		return nil
	})
}

// parseArgs extracts tool names from command-line arguments
// Usage: go run main.go [helm] [preflight] [support-bundle]
// If no args, returns empty slice (download all)
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

// shouldDownload checks if a tool should be downloaded
func shouldDownload(tool string, requestedTools []string) bool {
	for _, t := range requestedTools {
		if t == tool {
			return true
		}
	}
	return false
}

func downloadHelm(version, platformOS, platformArch, toolsDir string) error {
	fmt.Printf("  → Downloading helm %s for %s-%s...\n", version, platformOS, platformArch)

	// Download to memory first, only create directory if successful
	var data []byte
	var err error

	if platformOS == "windows" {
		url := fmt.Sprintf("https://get.helm.sh/helm-v%s-windows-%s.zip", version, platformArch)
		data, err = downloadAndExtractToMemory(url, "windows-"+platformArch+"/helm.exe", true)
	} else {
		url := fmt.Sprintf("https://get.helm.sh/helm-v%s-%s-%s.tar.gz", version, platformOS, platformArch)
		data, err = downloadAndExtractToMemory(url, platformOS+"-"+platformArch+"/helm", false)
	}

	if err != nil {
		fmt.Printf("    ✗ Failed to download helm\n")
		return err
	}

	// Download successful - now create directory and write file
	dest := filepath.Join(toolsDir, "helm", version, platformOS+"-"+platformArch)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	binaryName := "helm"
	if platformOS == "windows" {
		binaryName = "helm.exe"
	}
	helmPath := filepath.Join(dest, binaryName)

	if err := os.WriteFile(helmPath, data, 0755); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("    ✓ Saved to %s\n", helmPath)
	return nil
}

func downloadPreflight(version, platformOS, platformArch, toolsDir string) error {
	fmt.Printf("  → Downloading preflight %s for %s-%s...\n", version, platformOS, platformArch)

	// Troubleshoot uses different naming: darwin_all (universal), {os}_{arch}
	var assetName string
	if platformOS == "darwin" {
		assetName = "preflight_darwin_all.tar.gz"
	} else {
		assetName = fmt.Sprintf("preflight_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, assetName)

	binaryName := "preflight"
	if platformOS == "windows" {
		binaryName = "preflight.exe"
	}

	// Download to memory first
	data, err := downloadAndExtractToMemory(url, binaryName, false)
	if err != nil {
		fmt.Printf("    ✗ Failed to download preflight from %s\n", assetName)
		return err
	}

	// Download successful - now create directory and write file
	dest := filepath.Join(toolsDir, "preflight", version, platformOS+"-"+platformArch)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	preflightPath := filepath.Join(dest, binaryName)
	if err := os.WriteFile(preflightPath, data, 0755); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("    ✓ Saved to %s\n", preflightPath)
	return nil
}

func downloadSupportBundle(version, platformOS, platformArch, toolsDir string) error {
	fmt.Printf("  → Downloading support-bundle %s for %s-%s...\n", version, platformOS, platformArch)

	// Troubleshoot uses different naming: darwin_all (universal), {os}_{arch}
	var assetName string
	if platformOS == "darwin" {
		assetName = "support-bundle_darwin_all.tar.gz"
	} else {
		assetName = fmt.Sprintf("support-bundle_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, assetName)

	binaryName := "support-bundle"
	if platformOS == "windows" {
		binaryName = "support-bundle.exe"
	}

	// Download to memory first
	data, err := downloadAndExtractToMemory(url, binaryName, false)
	if err != nil {
		fmt.Printf("    ✗ Failed to download support-bundle from %s\n", assetName)
		return err
	}

	// Download successful - now create directory and write file
	dest := filepath.Join(toolsDir, "support-bundle", version, platformOS+"-"+platformArch)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	sbPath := filepath.Join(dest, binaryName)
	if err := os.WriteFile(sbPath, data, 0755); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("    ✓ Saved to %s\n", sbPath)
	return nil
}

// downloadAndExtractToMemory downloads an archive and extracts a specific file to memory
// Returns the file contents as bytes. Only creates directories if download succeeds.
func downloadAndExtractToMemory(url, fileInArchive string, isZip bool) ([]byte, error) {
	// Download
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read entire response into memory
	archiveData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if isZip {
		return extractFromZip(archiveData, fileInArchive)
	}
	return extractFromTarGz(archiveData, fileInArchive)
}

// extractFromZip extracts a specific file from zip archive in memory
func extractFromZip(archiveData []byte, fileInArchive string) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("reading zip: %w", err)
	}

	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, fileInArchive) {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("opening file in zip: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("reading file: %w", err)
			}

			return data, nil
		}
	}

	return nil, fmt.Errorf("file %q not found in archive", fileInArchive)
}

// extractFromTarGz extracts a specific file from tar.gz archive in memory
func extractFromTarGz(archiveData []byte, fileInArchive string) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return nil, fmt.Errorf("decompressing gzip: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		if strings.HasSuffix(header.Name, fileInArchive) {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("reading file: %w", err)
			}

			return data, nil
		}
	}

	return nil, fmt.Errorf("file %q not found in archive", fileInArchive)
}
