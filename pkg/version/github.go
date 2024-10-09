package version

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrUnknownArchiveType        = errors.New("unknown archive type")
	ErrNoMatchingArchitectures   = errors.New("no matching architectures")
	ErrNoAssets                  = errors.New("no assets")
	ErrChecksumMismatch          = errors.New("checksum mismatch")
	ErrUnsupportedChecksumFormat = errors.New("unsupported checksum format")
	ErrTimeoutExceeded           = errors.New("timeout exceeded")
)

type githubAsset struct {
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type gitHubReleaseInfo struct {
	TagName     string        `json:"tag_name"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

var (
	ErrReleaseNotFound = errors.New("release not found")
)

// downloadVersion will download and extract the specific version, returning
// a path to the extracted file in the archive
// it's the responsibility of the caller to clean up the extracted file
func downloadVersion(version string, requireChecksumMatch bool) (string, error) {
	releaseInfo, err := getReleaseDetails(http.DefaultClient.Timeout, "github.com", "replicatedhq", "replicated", version)
	if err != nil {
		return "", errors.Wrap(err, "get release details")
	}

	asset, err := bestAsset(releaseInfo.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", errors.Wrap(err, "best asset")
	}

	archivePath, fileInArchivePath, err := downloadFile(asset.BrowserDownloadURL, http.DefaultClient.Timeout)
	if err != nil {
		return "", errors.Wrap(err, "download file")
	}
	defer os.Remove(archivePath)

	checksumAsset, err := checksum(releaseInfo.Assets, asset.Name)
	if err != nil {
		return "", errors.Wrap(err, "checksum")
	}

	if checksumAsset != nil {
		desiredChecksum, err := downloadAndParseChecksum(http.DefaultClient.Timeout, checksumAsset.BrowserDownloadURL, asset.Name)
		if err != nil {
			return "", errors.Wrap(err, "download and parse checksum")
		}

		actualChecksum, err := checksumFile(archivePath)
		if err != nil {
			return "", errors.Wrap(err, "checksum file")
		}

		if actualChecksum != desiredChecksum {
			return "", ErrChecksumMismatch
		}
	}

	return fileInArchivePath, nil
}

func downloadAndParseChecksum(timeout time.Duration, url string, assetName string) (string, error) {
	// download the file
	httpClient := http.Client{
		Timeout: timeout,
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		if os.IsTimeout(err) {
			return "", ErrTimeoutExceeded
		}
		return "", err
	}
	defer resp.Body.Close()

	// parse the file
	// supported formats are sha256[whitespace]filepath per line

	// first try to find the exact match
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 {
			if strings.HasSuffix(strings.TrimSpace(parts[1]), assetName) {
				return strings.TrimSpace(parts[0]), nil
			}
		}
	}

	return "", ErrUnsupportedChecksumFormat
}

func checksumFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// checksumAsset will search through the assets and attempt to find the
// download url for the sha256 checksum for the asset provided
// this works by looking for the asset name with the checksum appended to it
// it will return empty string and no error if there is not checksum
func checksum(assets []githubAsset, assetName string) (*githubAsset, error) {
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		if strings.HasPrefix(asset.Name, assetName) {
			if strings.HasSuffix(asset.Name, ".sha256") {
				return &asset, nil
			}
		}
	}

	// no exact match, look for a common checksums file
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		if strings.Contains(asset.Name, "checksums") {
			if strings.HasSuffix(asset.Name, ".txt") {
				return &asset, nil
			}
		}
	}

	return nil, nil
}

// bestAsset will search through the assets, find the best (most appropriate)
// asset for the os nad arch provided. this will that asset
// for the asset
func bestAsset(assets []githubAsset, goos string, goarch string) (*githubAsset, error) {
	if len(assets) == 0 {
		return nil, ErrNoAssets
	}

	// find the most appropriate asset
	for _, asset := range assets {
		if asset.State != "uploaded" || asset.ContentType == "application/octet-stream" {
			continue
		}

		lowercaseName := strings.ToLower(asset.Name)
		if strings.Contains(lowercaseName, goos) {
			if strings.Contains(lowercaseName, goarch) {
				return &asset, nil
			}
		}
	}

	// we didn't find a specific match, look for the os with "all" for the arch
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		lowercaseName := strings.ToLower(asset.Name)
		if strings.Contains(lowercaseName, runtime.GOOS) {
			if strings.Contains(lowercaseName, "all") {
				return &asset, nil
			}
		}
	}

	return nil, ErrNoMatchingArchitectures
}

// downloadFile will return two strings:
//   - the path to the downloaded file (the archive)
//   - the path to the file that is probably the binary
func downloadFile(url string, timeout time.Duration) (string, string, error) {
	tmpFile, err := ioutil.TempFile("", "replicated-update")
	if err != nil {
		return "", "", errors.Wrap(err, "create temp file")
	}

	httpClient := http.Client{
		Timeout: timeout,
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		if os.IsTimeout(err) {
			return "", "", ErrTimeoutExceeded
		}
		return "", "", errors.Wrap(err, "get file")
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "copy file")
	}

	probableFile, err := findProbableFileInWhatMightBeAnArchive(tmpFile.Name())
	if err != nil {
		return "", "", errors.Wrap(err, "find probable file")
	}

	return tmpFile.Name(), probableFile, nil
}

func findProbableFileInWhatMightBeAnArchive(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	// check if it's a gzip file
	_, err = gzip.NewReader(f)
	if err == nil {
		return findProbableFileInGzip(path)
	}

	return "", ErrUnknownArchiveType
}

func findProbableFileInGzip(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", errors.Wrap(err, "open gzip file")
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", errors.Wrap(err, "read next file")
		}

		if header.Typeflag == tar.TypeReg {
			// if the file is executable and matches the name of the current process
			// then it's almost certainly the file we want

			if isLikelyFile(header.Mode, header.Name, filepath.Base(os.Args[0])) {
				tmpFile, err := ioutil.TempFile("", "replicated-update")
				if err != nil {
					return "", errors.Wrap(err, "create temp file")
				}

				defer func() {
					if err := tmpFile.Close(); err != nil {
						panic(err)
					}
				}()

				if _, err := io.Copy(tmpFile, tr); err != nil {
					return "", errors.Wrap(err, "copy file")
				}

				// set the mode on the file to match
				if err := os.Chmod(tmpFile.Name(), os.FileMode(header.Mode)); err != nil {
					return "", errors.Wrap(err, "set file mode")
				}

				return tmpFile.Name(), nil
			}
		}
	}

	return "", errors.New("unable to find matching file in archive")
}

func isLikelyFile(mode int64, name string, currentExecutableName string) bool {
	if mode&0111 != 0 {
		if currentExecutableName == filepath.Base(name) {
			return true
		}
	}

	return false
}

func getReleaseDetails(timeout time.Duration, host string, owner string, repo string, releaseName string) (*gitHubReleaseInfo, error) {
	uri := ""

	if releaseName == "latest" {
		uri = fmt.Sprintf("%s/repos/%s/%s/releases/latest", host, owner, repo)
	} else {
		uri = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", host, owner, repo, releaseName)
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	httpClient := http.Client{
		Timeout: timeout,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, ErrTimeoutExceeded
		}
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrReleaseNotFound
		}

		if resp.StatusCode == http.StatusForbidden && strings.Contains(resp.Header.Get("X-RateLimit-Remaining"), "0") {
			return nil, nil // don't make the caller handle this
		}

		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	releaseInfo := gitHubReleaseInfo{}

	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	return &releaseInfo, nil
}
