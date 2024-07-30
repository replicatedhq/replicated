package ociclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	digest "github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Blob struct {
	Digest       string
	Size         int64
	RelativePath string
	Permissions  os.FileMode
}

func uploadBlob(ctx context.Context, filePath, repoURL, jwtToken string) (*Blob, error) {
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
	}

	digest := fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	permissions := stat.Mode()

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

func createAndUploadConfig(ctx context.Context, repoURL, jwtToken string, blobs []*Blob) (*Blob, error) {
	now := time.Now()
	config := v1.Image{
		Created: &now,
		Config: v1.ImageConfig{
			Env:        []string{},
			Entrypoint: []string{},
			Cmd:        []string{},
			Volumes:    map[string]struct{}{},
			WorkingDir: "",
			User:       "",
		},
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{},
		},
		History: []v1.History{
			{
				Created:   &now,
				CreatedBy: "file upload",
			},
		},
	}

	for _, blob := range blobs {
		config.RootFS.DiffIDs = append(config.RootFS.DiffIDs, digest.Digest(blob.Digest))
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	tempFile, err := os.CreateTemp("", "config.json")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(configBytes)
	if err != nil {
		return nil, err
	}
	tempFile.Close()

	return uploadBlob(ctx, tempFile.Name(), repoURL, jwtToken)
}

func uploadManifest(ctx context.Context, blobs []*Blob, configBlob *Blob, repoURL, jwtToken, tag string) error {
	var layers []v1.Descriptor
	for _, blob := range blobs {
		layers = append(layers, v1.Descriptor{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Digest:    digest.Digest(blob.Digest),
			Size:      blob.Size,
			Annotations: map[string]string{
				"org.opencontainers.image.layer.path":        blob.RelativePath,
				"org.opencontainers.image.layer.permissions": fmt.Sprintf("%o", blob.Permissions),
			},
		})
	}

	manifest := v1.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Config: v1.Descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    digest.Digest(configBlob.Digest),
			Size:      configBlob.Size,
		},
		Layers: layers,
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/manifests/%s", repoURL, tag), bytes.NewReader(manifestBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(manifestBytes)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload manifest: %s", resp.Status)
	}

	return nil
}
