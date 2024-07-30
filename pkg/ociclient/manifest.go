package ociclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func uploadManifest(ctx context.Context, blobs []*Blob, configBlob *Blob, repoURL, jwtToken, tag string, baseDir string) error {
	var layers []v1.Descriptor
	for _, blob := range blobs {
		relativePath, err := filepath.Rel(baseDir, blob.RelativePath)
		if err != nil {
			return err
		}

		layers = append(layers, v1.Descriptor{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Digest:    digest.Digest(blob.Digest),
			Size:      blob.Size,
			Annotations: map[string]string{
				"org.opencontainers.image.layer.path":        relativePath,
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
