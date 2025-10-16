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
)

// Tool versions (should match defaults in pkg/tools/types.go)
const (
	HelmVersion          = "3.14.4"
	PreflightVersion     = "0.123.9"
	SupportBundleVersion = "0.123.9"
)

type downloadResult struct {
	tool    string
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

	fmt.Printf("Detected platform: %s-%s\n", platformOS, platformArch)
	fmt.Printf("Downloading tools: %v\n", requestedTools)
	fmt.Println()

	toolsDir := "pkg/tools/embedded"

	// Track results
	var results []downloadResult

	// Download requested tools (or all if none specified)
	if shouldDownload("helm", requestedTools) {
		if err := downloadHelm(HelmVersion, platformOS, platformArch, toolsDir); err != nil {
			results = append(results, downloadResult{"helm", false, err})
		} else {
			results = append(results, downloadResult{"helm", true, nil})
		}
	}

	if shouldDownload("preflight", requestedTools) {
		if err := downloadPreflight(PreflightVersion, platformOS, platformArch, toolsDir); err != nil {
			results = append(results, downloadResult{"preflight", false, err})
		} else {
			results = append(results, downloadResult{"preflight", true, nil})
		}
	}

	if shouldDownload("support-bundle", requestedTools) {
		if err := downloadSupportBundle(SupportBundleVersion, platformOS, platformArch, toolsDir); err != nil {
			results = append(results, downloadResult{"support-bundle", false, err})
		} else {
			results = append(results, downloadResult{"support-bundle", true, nil})
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("Download Summary:")
	successCount := 0
	for _, r := range results {
		if r.success {
			fmt.Printf("  ✓ %s - success\n", r.tool)
			successCount++
		} else {
			fmt.Printf("  ✗ %s - failed: %v\n", r.tool, r.err)
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
	dest := filepath.Join(toolsDir, "helm", version, platformOS+"-"+platformArch)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	fmt.Printf("  → Downloading helm %s for %s-%s...\n", version, platformOS, platformArch)

	// Windows uses .zip, others use .tar.gz
	var err error
	if platformOS == "windows" {
		url := fmt.Sprintf("https://get.helm.sh/helm-v%s-windows-%s.zip", version, platformArch)
		err = downloadAndExtractZip(url, dest, "windows-"+platformArch+"/helm.exe", "helm.exe")
	} else {
		url := fmt.Sprintf("https://get.helm.sh/helm-v%s-%s-%s.tar.gz", version, platformOS, platformArch)
		err = downloadAndExtractTarGz(url, dest, platformOS+"-"+platformArch+"/helm", "helm")
	}

	if err != nil {
		fmt.Printf("    ✗ Failed to download helm\n")
		return err
	}

	// Make executable (on Unix)
	binaryName := "helm"
	if platformOS == "windows" {
		binaryName = "helm.exe"
	}
	helmPath := filepath.Join(dest, binaryName)
	if err := os.Chmod(helmPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	fmt.Printf("    ✓ Saved to %s\n", helmPath)
	return nil
}

func downloadPreflight(version, platformOS, platformArch, toolsDir string) error {
	dest := filepath.Join(toolsDir, "preflight", version, platformOS+"-"+platformArch)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	fmt.Printf("  → Downloading preflight %s for %s-%s...\n", version, platformOS, platformArch)

	// Troubleshoot uses different naming: darwin_all (universal), {os}_{arch}
	var assetName string
	if platformOS == "darwin" {
		assetName = "preflight_darwin_all.tar.gz"
	} else {
		assetName = fmt.Sprintf("preflight_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, assetName)

	// Determine binary name (Windows uses .exe)
	binaryName := "preflight"
	if platformOS == "windows" {
		binaryName = "preflight.exe"
	}

	// Download and extract
	if err := downloadAndExtractTarGz(url, dest, binaryName, binaryName); err != nil {
		fmt.Printf("    ✗ Failed to download preflight from %s\n", assetName)
		return err
	}

	// Make executable
	preflightPath := filepath.Join(dest, binaryName)
	if err := os.Chmod(preflightPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	fmt.Printf("    ✓ Saved to %s\n", preflightPath)
	return nil
}

func downloadSupportBundle(version, platformOS, platformArch, toolsDir string) error {
	dest := filepath.Join(toolsDir, "support-bundle", version, platformOS+"-"+platformArch)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	fmt.Printf("  → Downloading support-bundle %s for %s-%s...\n", version, platformOS, platformArch)

	// Troubleshoot uses different naming: darwin_all (universal), {os}_{arch}
	var assetName string
	if platformOS == "darwin" {
		assetName = "support-bundle_darwin_all.tar.gz"
	} else {
		assetName = fmt.Sprintf("support-bundle_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, assetName)

	// Determine binary name (Windows uses .exe)
	binaryName := "support-bundle"
	if platformOS == "windows" {
		binaryName = "support-bundle.exe"
	}

	// Download and extract
	if err := downloadAndExtractTarGz(url, dest, binaryName, binaryName); err != nil {
		fmt.Printf("    ✗ Failed to download support-bundle from %s\n", assetName)
		return err
	}

	// Make executable
	sbPath := filepath.Join(dest, binaryName)
	if err := os.Chmod(sbPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	fmt.Printf("    ✓ Saved to %s\n", sbPath)
	return nil
}

// downloadAndExtractZip downloads a .zip file and extracts a specific file from it
func downloadAndExtractZip(url, destDir, fileInArchive, outputName string) error {
	// Download
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read entire response into memory (needed for zip reader)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// Create zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("reading zip: %w", err)
	}

	// Find and extract the target file
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, fileInArchive) {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("opening file in zip: %w", err)
			}
			defer rc.Close()

			outPath := filepath.Join(destDir, outputName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				return fmt.Errorf("extracting file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("file %q not found in archive", fileInArchive)
}

// downloadAndExtractTarGz downloads a tar.gz file and extracts a specific file from it
func downloadAndExtractTarGz(url, destDir, fileInArchive, outputName string) error {
	// Download
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Decompress gzip
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("decompressing gzip: %w", err)
	}
	defer gzReader.Close()

	// Read tar
	tarReader := tar.NewReader(gzReader)

	// Find and extract the target file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		// Check if this is the file we want
		if strings.HasSuffix(header.Name, fileInArchive) {
			outPath := filepath.Join(destDir, outputName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("extracting file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("file %q not found in archive", fileInArchive)
}
