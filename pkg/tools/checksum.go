package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// httpClient for checksum downloads with timeout
var checksumHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// VerifyHelmChecksum verifies a Helm binary against its .sha256sum file
func VerifyHelmChecksum(data []byte, archiveURL string) error {
	// Helm provides per-file checksums: <url>.sha256sum
	checksumURL := archiveURL + ".sha256sum"

	// Download checksum file with timeout
	resp, err := checksumHTTPClient.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("downloading checksum file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("checksum file not found (HTTP %d): %s", resp.StatusCode, checksumURL)
	}

	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading checksum file: %w", err)
	}

	// Parse checksum (format: "abc123  helm-v3.14.4-darwin-arm64.tar.gz")
	parts := strings.Fields(string(checksumData))
	if len(parts) < 1 {
		return fmt.Errorf("invalid checksum file format")
	}
	expectedSum := parts[0]

	// Calculate actual checksum of the archive data
	hash := sha256.Sum256(data)
	actualSum := hex.EncodeToString(hash[:])

	// Verify match
	if actualSum != expectedSum {
		return fmt.Errorf("checksum mismatch: got %s, want %s", actualSum, expectedSum)
	}

	return nil
}

// VerifyTroubleshootChecksum verifies preflight or support-bundle against checksums.txt
func VerifyTroubleshootChecksum(data []byte, version, filename string) error {
	// Troubleshoot provides a single checksums file for all binaries
	checksumURL := fmt.Sprintf("https://github.com/replicatedhq/troubleshoot/releases/download/v%s/troubleshoot_%s_checksums.txt", version, version)

	// Download checksums file with timeout
	resp, err := checksumHTTPClient.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("downloading checksums file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("checksums file not found (HTTP %d): %s", resp.StatusCode, checksumURL)
	}

	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading checksums file: %w", err)
	}

	// Find the checksum for our specific file
	// Format: "abc123  preflight_darwin_all.tar.gz"
	var expectedSum string
	for _, line := range strings.Split(string(checksumData), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == filename {
			expectedSum = parts[0]
			break
		}
	}

	if expectedSum == "" {
		return fmt.Errorf("checksum not found for %s in checksums file", filename)
	}

	// Calculate actual checksum of the archive data
	hash := sha256.Sum256(data)
	actualSum := hex.EncodeToString(hash[:])

	// Verify match
	if actualSum != expectedSum {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", filename, actualSum, expectedSum)
	}

	return nil
}
