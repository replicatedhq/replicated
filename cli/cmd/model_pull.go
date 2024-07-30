package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func (r *runners) InitModelPull(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "pull [NAME:TAG] [DESTINATION]",
		Short:        "pull a model from the model repository",
		Long:         `pull a model from the model repository`,
		RunE:         r.pullModel,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) pullModel(cmd *cobra.Command, args []string) error {
	// get the endpoint
	endpoint, err := r.kotsAPI.GetModelsEndpoint()
	if err != nil {
		return err
	}

	if endpoint == "" {
		return errors.New("Could not find models endpoint")
	}
	fmt.Printf("Pulling model from %s\n", endpoint)

	name, tag := nameToNameAndTag(args[0])

	// oras copy this from the endpoint to the local registry
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", endpoint, name))
	if err != nil {
		return err
	}
	creds, err := credentials.GetCurrentCredentials()
	if err != nil {
		return err
	}

	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(endpoint, auth.Credential{
			Username: "",
			Password: creds.APIToken,
		}),
	}

	// ensure that destination exists and is a directory
	destination := args[1]
	if _, err := os.Stat(destination); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(destination, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// ensure that destination is empty
	files, err := listFilesInPath(destination)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		return fmt.Errorf("destination %q is not empty", destination)
	}

	fs, err := file.New(destination)
	if err != nil {
		return err
	}
	defer fs.Close()

	// Copy the manifest from the remote repository
	manifestDescriptor, err := oras.Copy(context.Background(), repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	// Fetch the manifest content
	manifestContent, err := fs.Fetch(context.Background(), manifestDescriptor)
	if err != nil {
		return err
	}
	defer manifestContent.Close()

	// Read the manifest content into a byte slice
	manifestBytes, err := io.ReadAll(manifestContent)
	if err != nil {
		return err
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return err
	}

	// Download each layer in the manifest
	for _, layer := range manifest.Layers {
		relativePath := layer.Annotations["org.opencontainers.image.layer.path"]
		if relativePath == "" {
			return fmt.Errorf("missing relative path annotation for layer %s", layer.Digest)
		}

		permissionsStr := layer.Annotations["org.opencontainers.image.layer.permissions"]
		if permissionsStr == "" {
			return fmt.Errorf("missing permissions annotation for layer %s", layer.Digest)
		}

		permissions, err := strconv.ParseUint(permissionsStr, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid permissions annotation for layer %s: %v", layer.Digest, err)
		}

		layerPath := filepath.Join(destination, relativePath)
		layerDir := filepath.Dir(layerPath)
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return err
		}

		layerFile, err := os.Create(layerPath)
		if err != nil {
			return err
		}

		err = downloadBlob(context.Background(), repo, layer, layerFile)
		if err != nil {
			layerFile.Close()
			return err
		}
		layerFile.Close()

		if err := os.Chmod(layerPath, os.FileMode(permissions)); err != nil {
			return err
		}
	}

	fmt.Printf("Model pulled successfully to %s\n", destination)
	return nil
}

func downloadBlob(ctx context.Context, repo *remote.Repository, descriptor v1.Descriptor, file *os.File) error {
	blobContent, err := repo.Blobs().Fetch(ctx, descriptor)
	if err != nil {
		return err
	}
	defer blobContent.Close()

	if _, err := io.Copy(file, blobContent); err != nil {
		return err
	}

	return nil
}
