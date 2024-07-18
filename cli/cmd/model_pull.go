package cmd

import (
	"context"
	"fmt"

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
		Use:          "pull [NAME:TAG]",
		Short:        "pull a model from the model repository",
		Long:         `pull a model from the mdoel repository`,
		RunE:         r.pullModel,
		Args:         cobra.ExactArgs(1),
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

	// oras copy this from the endpoint to the local registry
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

	fs, err := file.New("/tmp/")
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	tag := "latest"
	manifestDescriptor, err := oras.Copy(context.Background(), repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}
	fmt.Println("manifest descriptor:", manifestDescriptor)

	return nil

}
