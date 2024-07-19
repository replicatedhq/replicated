package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/ociclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitModelPush(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "push [FILE] [NAME:TAG]",
		Short:        "push a model to the model repository",
		Long:         `push a model to the model repository`,
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
	name, tag := nameToNameAndTag(args[1])

	modelFiles, err := listFilesInPath(path)
	if err != nil {
		return err
	}

	fullPaths := []string{}
	for _, modelFile := range modelFiles {
		fullPaths = append(fullPaths, filepath.Base(modelFile))
	}

	if err := ociclient.UploadFiles(context.Background(), endpoint, name, tag, fullPaths); err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.w, "Model %s:%s has been pushed\n", name, tag)
	if err != nil {
		return err
	}

	return nil
}

// listFilesInPath returns a list of files in the given path
// with the returning array being the absolute path of each file
func listFilesInPath(path string) ([]string, error) {
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
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			files = append(files, absPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func nameToNameAndTag(nameAndTag string) (string, string) {
	parts := strings.Split(nameAndTag, ":")
	name := parts[0]
	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}

	return name, tag
}
