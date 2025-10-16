package tools

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// Download timeout per attempt
	downloadTimeout = 5 * time.Minute

	// Max retries for failed downloads
	maxRetries = 3

	// Initial backoff duration
	initialBackoff = 1 * time.Second
)

// Downloader handles downloading tool binaries
type Downloader struct {
	httpClient *http.Client
}

// NewDownloader creates a new downloader with timeout
func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: downloadTimeout,
		},
	}
}

// Download downloads a tool to the cache directory with checksum verification
// If the requested version fails, it will automatically fallback to latest stable
// Returns the actual version that was downloaded (may differ from requested version due to fallback)
func (d *Downloader) Download(ctx context.Context, name, version string) (string, error) {
	// Try with fallback
	actualVersion, err := d.DownloadWithFallback(ctx, name, version)
	return actualVersion, err
}

// downloadExact downloads a specific version without fallback (internal use)
func (d *Downloader) downloadExact(ctx context.Context, name, version string) error {
	// Get cache path
	cachePath, err := GetToolPath(name, version)
	if err != nil {
		return err
	}

	// Download binary and get checksum info
	var archiveData []byte
	var checksumURL, checksumFilename string

	switch name {
	case ToolHelm:
		archiveData, checksumURL, err = d.downloadHelmArchive(version)
	case ToolPreflight:
		archiveData, checksumURL, checksumFilename, err = d.downloadPreflightArchive(version)
	case ToolSupportBundle:
		archiveData, checksumURL, checksumFilename, err = d.downloadSupportBundleArchive(version)
	default:
		return fmt.Errorf("unknown tool: %s", name)
	}

	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}

	// Verify checksum
	if name == ToolHelm {
		if err := VerifyHelmChecksum(archiveData, checksumURL); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	} else {
		// Troubleshoot tools (preflight, support-bundle)
		if err := VerifyTroubleshootChecksum(archiveData, version, checksumFilename); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Extract binary from archive
	var binaryData []byte
	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}

	switch name {
	case ToolHelm:
		if runtime.GOOS == "windows" {
			binaryData, err = extractFromZip(archiveData, "windows-"+runtime.GOARCH+"/helm.exe")
		} else {
			binaryData, err = extractFromTarGz(archiveData, runtime.GOOS+"-"+runtime.GOARCH+"/helm")
		}
	case ToolPreflight, ToolSupportBundle:
		// Windows troubleshoot uses .zip, others use .tar.gz
		if runtime.GOOS == "windows" {
			binaryData, err = extractFromZip(archiveData, binaryName)
		} else {
			binaryData, err = extractFromTarGz(archiveData, binaryName)
		}
	}

	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	// Create directory only after successful download, verification, and extraction
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	// Write binary
	if err := os.WriteFile(cachePath, binaryData, 0755); err != nil {
		return fmt.Errorf("writing binary: %w", err)
	}

	return nil
}

// githubRelease represents a GitHub release
type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
}

// getLatestStableVersion fetches the latest non-prerelease version from GitHub
func getLatestStableVersion(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing release JSON: %w", err)
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")

	return version, nil
}

// DownloadWithFallback attempts to download the specified version, falling back to latest stable if it fails
func (d *Downloader) DownloadWithFallback(ctx context.Context, name, version string) (string, error) {
	// Try requested version first
	err := d.downloadExact(ctx, name, version)
	if err == nil {
		return version, nil
	}

	// If requested version failed, try latest stable
	fmt.Printf("⚠️  Version %s failed: %v\n", version, err)
	fmt.Printf("Attempting to download latest stable version...\n")

	var repo string
	switch name {
	case ToolHelm:
		repo = "helm/helm"
	case ToolPreflight, ToolSupportBundle:
		repo = "replicatedhq/troubleshoot"
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	latestVersion, err := getLatestStableVersion(repo)
	if err != nil {
		return "", fmt.Errorf("could not get latest version: %w", err)
	}

	fmt.Printf("Latest stable version: %s\n", latestVersion)

	// Try downloading latest
	if err := d.downloadExact(ctx, name, latestVersion); err != nil {
		return "", fmt.Errorf("latest version also failed: %w", err)
	}

	return latestVersion, nil
}

// downloadWithRetry downloads a URL with retry logic and exponential backoff
func (d *Downloader) downloadWithRetry(url string) ([]byte, error) {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("  Retry %d/%d after %v...\n", attempt, maxRetries-1, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		// Attempt download
		resp, err := d.httpClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("downloading: %w", err)
			continue // Retry
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			// Don't retry 404s (version doesn't exist)
			if resp.StatusCode == 404 {
				return nil, fmt.Errorf("HTTP 404: file not found")
			}
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue // Retry other status codes
		}

		// Success - read data
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue // Retry
		}

		return data, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadHelmArchive downloads the helm archive and returns archive data + checksum URL
func (d *Downloader) downloadHelmArchive(version string) ([]byte, string, error) {
	platformOS := runtime.GOOS
	platformArch := runtime.GOARCH

	var url string
	if platformOS == "windows" {
		url = fmt.Sprintf("https://get.helm.sh/helm-v%s-windows-%s.zip", version, platformArch)
	} else {
		url = fmt.Sprintf("https://get.helm.sh/helm-v%s-%s-%s.tar.gz", version, platformOS, platformArch)
	}

	// Download archive with retry
	data, err := d.downloadWithRetry(url)
	if err != nil {
		return nil, "", err
	}

	// Checksum URL is the archive URL + .sha256sum
	return data, url, nil
}

// downloadPreflightArchive downloads the preflight archive
func (d *Downloader) downloadPreflightArchive(version string) ([]byte, string, string, error) {
	platformOS := runtime.GOOS
	platformArch := runtime.GOARCH

	// Troubleshoot uses different naming
	// Windows uses .zip, others use .tar.gz
	var filename string
	if platformOS == "darwin" {
		filename = "preflight_darwin_all.tar.gz"
	} else if platformOS == "windows" {
		filename = fmt.Sprintf("preflight_%s_%s.zip", platformOS, platformArch)
	} else {
		filename = fmt.Sprintf("preflight_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, filename)

	// Download archive with retry
	data, err := d.downloadWithRetry(url)
	if err != nil {
		return nil, "", "", err
	}

	// For troubleshoot, we need the checksums.txt URL and the filename to look up
	checksumURL := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/troubleshoot_%s_checksums.txt", version, version)

	return data, checksumURL, filename, nil
}

// downloadSupportBundleArchive downloads the support-bundle archive
func (d *Downloader) downloadSupportBundleArchive(version string) ([]byte, string, string, error) {
	platformOS := runtime.GOOS
	platformArch := runtime.GOARCH

	// Troubleshoot uses different naming
	// Windows uses .zip, others use .tar.gz
	var filename string
	if platformOS == "darwin" {
		filename = "support-bundle_darwin_all.tar.gz"
	} else if platformOS == "windows" {
		filename = fmt.Sprintf("support-bundle_%s_%s.zip", platformOS, platformArch)
	} else {
		filename = fmt.Sprintf("support-bundle_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, filename)

	// Download archive with retry
	data, err := d.downloadWithRetry(url)
	if err != nil {
		return nil, "", "", err
	}

	// For troubleshoot, we need the checksums.txt URL and the filename to look up
	checksumURL := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/troubleshoot_%s_checksums.txt", version, version)

	return data, checksumURL, filename, nil
}

// extractFromZip extracts a specific file from a zip archive in memory
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

// extractFromTarGz extracts a specific file from a tar.gz archive in memory
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
