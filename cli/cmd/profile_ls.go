package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
)

func (r *runners) InitProfileLsCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all authentication profiles",
		Long: `List all authentication profiles configured in ~/.replicated/config.yaml.

The default profile is indicated with an asterisk (*).`,
		Example: `# List all profiles
replicated profile ls`,
		SilenceUsage: true,
		RunE:         r.profileLs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) profileLs(_ *cobra.Command, _ []string) error {
	profiles, defaultProfile, err := credentials.ListProfiles()
	if err != nil {
		return errors.Wrap(err, "failed to list profiles")
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles configured")
		fmt.Println("")
		fmt.Println("To add a profile, run:")
		fmt.Println("  replicated profile add <name>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DEFAULT\tNAME\tAPI ORIGIN\tREGISTRY ORIGIN")

	for name, profile := range profiles {
		isDefault := ""
		if name == defaultProfile {
			isDefault = "*"
		}

		apiOrigin := profile.APIOrigin
		if apiOrigin == "" {
			apiOrigin = "<default>"
		}

		registryOrigin := profile.RegistryOrigin
		if registryOrigin == "" {
			registryOrigin = "<default>"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			isDefault,
			name,
			apiOrigin,
			registryOrigin,
		)
	}

	w.Flush()
	return nil
}
