package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (r *runners) InitModelPush(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "push [FILE] [NAME:TAG]",
		Short:        "push a model to the model repository",
		Long:         `push a model to the mdoel repository`,
		RunE:         r.pushModel,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) pushModel(cmd *cobra.Command, args []string) error {
	endpoint, err := r.kotsAPI.GetModelsEndpoint()
	if err != nil {
		return err
	}

	if endpoint == "" {
		return errors.New("Could not find models endpoint")
	}
	fmt.Printf("Pushing model to %s\n", endpoint)

	path := args[0]
	nameAndTag := args[1]
	parts := strings.Split(nameAndTag, ":")
	name := parts[0]
	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}

	modelFiles, err := listFilesInPath(path)
	if err != nil {
		return err
	}

	fs, err := file.New("")
	if err != nil {
		return err
	}
	defer fs.Close()
	ctx := context.Background()

	fileDescriptors := []v1.Descriptor{}

	for _, modelFile := range modelFiles {
		fmt.Printf("Adding %s\n", modelFile)
		mediaType := "application/vnd.replicated.modelfile"

		fileName := filepath.Base(modelFile)
		fileDescriptor, err := fs.Add(ctx, fileName, mediaType, modelFile)
		if err != nil {
			return err
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)
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

	repo, err := remote.NewRepository(endpoint + "/model")
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

	_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.w, "Model %s:%s has been pushed\n", name, tag)
	if err != nil {
		return err
	}

	return err
}

func listFilesInPath(path string) ([]string, error) {
	// if path is a file, not a directory, return it
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	var files []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
