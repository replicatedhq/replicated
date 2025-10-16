package tools

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Downloader handles downloading tool binaries
type Downloader struct{}

// NewDownloader creates a new downloader
func NewDownloader() *Downloader {
	return &Downloader{}
}

// Download downloads a tool to the cache directory with checksum verification
func (d *Downloader) Download(ctx context.Context, name, version string) error {
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
		binaryData, err = extractFromTarGz(archiveData, binaryName)
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

	// Download archive
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response: %w", err)
	}

	// Checksum URL is the archive URL + .sha256sum (pass URL not checksumURL for verification)
	return data, url, nil
}

// downloadPreflightArchive downloads the preflight archive
func (d *Downloader) downloadPreflightArchive(version string) ([]byte, string, string, error) {
	platformOS := runtime.GOOS
	platformArch := runtime.GOARCH

	// Troubleshoot uses different naming
	var filename string
	if platformOS == "darwin" {
		filename = "preflight_darwin_all.tar.gz"
	} else {
		filename = fmt.Sprintf("preflight_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, filename)

	// Download archive
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", "", fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", fmt.Errorf("reading response: %w", err)
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
	var filename string
	if platformOS == "darwin" {
		filename = "support-bundle_darwin_all.tar.gz"
	} else {
		filename = fmt.Sprintf("support-bundle_%s_%s.tar.gz", platformOS, platformArch)
	}

	url := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/%s", version, filename)

	// Download archive
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", "", fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", fmt.Errorf("reading response: %w", err)
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
