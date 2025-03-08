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
			currentVersion := version.Version()
				
			// For version command, do a synchronous update check
			updateChecker, err := version.NewUpdateChecker(currentVersion, "replicatedhq/replicated/cli") 
			if err == nil {
				// If we're in a development build or unknown version, still try to get latest
				// version info but don't compare versions
				updateInfo, err := updateChecker.GetUpdateInfo()
				if err == nil && updateInfo != nil {
					// Update the build info with latest version
					version.SetUpdateInfo(updateInfo)
					// Also save to cache for future commands
					version.SaveUpdateCache(currentVersion, updateInfo)
				}
			}
			
			// Now get the (potentially updated) build info
			build := version.GetBuild()

			if !versionJson {
				// Special handling for development/unknown version when printing
				if currentVersion == "unknown" || currentVersion == "development" {
					fmt.Printf("replicated version %s (development build)\n", currentVersion)
					if build.UpdateInfo != nil {
						fmt.Printf("Latest release: %s\n", build.UpdateInfo.LatestVersion)
						if build.UpdateInfo.CanUpgradeInPlace {
							fmt.Printf("To automatically upgrade, run \"replicated version upgrade\"\n")
						} else if build.UpdateInfo.ExternalUpgradeCommand != "" {
							fmt.Printf("To upgrade, run \"%s\"\n", build.UpdateInfo.ExternalUpgradeCommand)
						}
					}
				} else {
					version.Print()
				}
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