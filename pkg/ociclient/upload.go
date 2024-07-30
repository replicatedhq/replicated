package ociclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

type Blob struct {
	Digest       string
	Size         int64
	RelativePath string
	Permissions  os.FileMode
}

func uploadBlob(ctx context.Context, filePath, repoURL, jwtToken string, showProgress bool, modelName string) (*Blob, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/blobs/uploads/", repoURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("failed to initiate upload: %s", resp.Status)
	}

	uploadURL := resp.Header.Get("Location")
	if uploadURL == "" {
		return nil, fmt.Errorf("no upload URL returned")
	}

	chunkSize, err := determineChunkSize(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to determine chunk size: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	permissions := stat.Mode()

	var bar *progressbar.ProgressBar
	if showProgress {
		// just the filename on the progress bar
		filename := filepath.Base(filePath)
		bar = progressbar.NewOptions(int(size),
			progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription(fmt.Sprintf("[cyan][%s][reset] Uploading %s...", modelName, filename)),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
	}

	hasher := sha256.New()
	buf := make([]byte, chunkSize)
	totalSize := int64(0)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		hasher.Write(buf[:n])
		totalSize += int64(n)

		chunk := bytes.NewReader(buf[:n])
		contentRange := fmt.Sprintf("bytes %d-%d/%d", totalSize-int64(n), totalSize-1, totalSize)

		req, err := http.NewRequest("PATCH", uploadURL, chunk)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", n))
		req.Header.Set("Content-Range", contentRange)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			return nil, fmt.Errorf("failed to upload chunk: %s", resp.Status)
		}

		uploadURL = resp.Header.Get("Location")
		if uploadURL == "" {
			return nil, fmt.Errorf("no upload URL returned")
		}

		if bar != nil {
			bar.Add64(int64(n))
		}
	}

	if bar != nil {
		fmt.Println("")
		bar.RenderBlank()
	}

	digest := fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))

	req, err = http.NewRequest("PUT", uploadURL+"?digest="+digest, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to finalize upload: %s", resp.Status)
	}

	return &Blob{
		Digest:       digest,
		Size:         size,
		RelativePath: filePath,
		Permissions:  permissions,
	}, nil
}
