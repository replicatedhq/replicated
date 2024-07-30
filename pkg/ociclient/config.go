package ociclient

import (
	"context"
	"encoding/json"
	"os"
	"time"

	digest "github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

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

	return uploadBlob(ctx, tempFile.Name(), repoURL, jwtToken, false, "")
}
