package cmd

import (
	"fmt"

	"github.com/replicatedhq/replicated/pkg/version"
	"github.com/spf13/cobra"
)

func versionUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "upgrade",
		Short:        "Upgrade the replicated CLI to the latest version",
		Long:         `Download, verify, and upgrade the Replicated CLI to the latest version`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := version.VerifyCanUpgrade()
			if err != nil {
				return err
			}

			if err := version.PerformUpgrade(); err != nil {
				return err
			}

			fmt.Printf("Upgrade complete\n")

			return nil
		},
	}

	return cmd
}
