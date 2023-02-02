package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/pkg/version"
)

func Version() *cobra.Command {
	var versionJson bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the current version and exit",
		Long:  `Print the current version and exit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			build := version.GetBuild()

			if !versionJson {
				version.Print()
			} else {
				versionInfo, err := json.MarshalIndent(build, "", "    ")
				if err != nil {
					return err
				}
				fmt.Println(string(versionInfo))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&versionJson, "json", false, "output version info in json")

	cmd.AddCommand(versionUpgradeCmd())

	return cmd
}
