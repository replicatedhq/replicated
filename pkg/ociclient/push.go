package ociclient

import (
	"context"
	"fmt"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

func UploadFiles(ctx context.Context, endpoint string, name string, tag string, filePaths []string) error {
	fs, err := file.New("")
	if err != nil {
		return err
	}
	defer fs.Close()

	fileDescriptors := []v1.Descriptor{}
	blobs := []*Blob{}

	jwtToken, err := getJWTToken(endpoint)
	if err != nil {
		return err
	}

	repoURL := fmt.Sprintf("https://%s/v2/%s", endpoint, name)

	for _, modelFile := range filePaths {
		fmt.Printf("Adding %s\n", modelFile)
		mediaType := "application/vnd.replicated.modelfile"

		fileDescriptor, err := fs.Add(ctx, modelFile, mediaType, modelFile)
		if err != nil {
			return err
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)

		blob, err := uploadBlob(ctx, modelFile, repoURL, jwtToken)
		if err != nil {
			return err
		}
		blobs = append(blobs, blob)
	}

	// Create and upload config layer
	configBlob, err := createAndUploadConfig(ctx, repoURL, jwtToken, blobs)
	if err != nil {
		return err
	}

	artifactType := "application/vnd.replicated.model"
	opts := oras.PackManifestOptions{
		Layers: fileDescriptors,
	}
	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return err
	}

	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		return err
	}

	err = uploadManifest(ctx, blobs, configBlob, repoURL, jwtToken, tag)
	if err != nil {
		return err
	}

	return nil
}
