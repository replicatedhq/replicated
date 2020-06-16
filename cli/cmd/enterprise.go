package cmd

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/enterpriseclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseCommand(parent *cobra.Command) *cobra.Command {
	enterpriseCommand := &cobra.Command{
		Use:          "enterprise",
		Short:        "Manage enterprise channels, policies and installers",
		SilenceUsage: true,
		Long:         `The enterprise command allows approved enterprise to create custom installers, release channels and policies for vendors`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Use == "init" {
				r.enterpriseClient = enterpriseclient.NewHTTPClient(enterpriseOrigin, nil)
				return nil
			}

			if enterprisePrivateKeyPath == "" {
				enterprisePrivateKeyPath = os.Getenv("REPLICATED_PRIVATEKEY")
				if enterprisePrivateKeyPath == "" {
					return errors.New("missing private key")
				}
			}

			if _, err := os.Stat(enterprisePrivateKeyPath); err != nil {
				if os.IsNotExist(err) {
					return errors.Errorf("file %s does not exist", enterprisePrivateKeyPath)
				}

				return err
			}

			privateKeyContents, err := ioutil.ReadFile(enterprisePrivateKeyPath)
			if err != nil {
				return err
			}

			r.enterpriseClient = enterpriseclient.NewHTTPClient(enterpriseOrigin, privateKeyContents)

			return nil
		},
	}
	parent.AddCommand(enterpriseCommand)

	// TODO remove the app and token persistent flags

	enterpriseCommand.PersistentFlags().StringVar(&enterprisePrivateKeyPath, "private-key", enterprisePrivateKeyPath, "Path to the private key used to sign requests")

	return enterpriseCommand
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
