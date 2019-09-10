package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new release",
		Long: `Create a new release by providing YAML configuration for the next release in
  your sequence.`,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createReleaseYaml, "yaml", "", "The YAML config for this release. Use '-' to read from stdin.  Cannot be used with the `yaml-file` falg.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlDir, "yaml-dir", "", "The directory containing the 5 required YAML configs for a Kots release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleasePromote, "promote", "", "Channel name or id to promote this release to")
	cmd.Flags().StringVar(&r.args.createReleasePromoteNotes, "release-notes", "", "When used with --promote <channel>, sets the **markdown** release notes")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteRequired, "required", false, "When used with --promote <channel>, marks this release as required during upgrades.")
	cmd.Flags().StringVar(&r.args.createReleasePromoteVersion, "version", "", "When used with --promote <channel>, sets the version label for the release in this channel")

	cmd.RunE = r.releaseCreate
}

func (r *runners) releaseCreate(cmd *cobra.Command, args []string) error {

	if r.args.createReleaseYaml == "" && r.args.createReleaseYamlFile == "" && r.args.createReleaseYamlDir == "" {
		return fmt.Errorf("yaml is required")
	}

	if r.args.createReleaseYaml != "" && r.args.createReleaseYamlFile != "" {
		return fmt.Errorf("only one of yaml or yaml-file may be specified")
	}

	if r.args.createReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.createReleaseYamlFile)
		if err != nil {
			return err
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlDir != "" {
		files, err := ioutil.ReadDir(r.args.createReleaseYamlDir)
		if err != nil {
			return err
		}

		var bytes []byte

		type kotsSingleSpec map[string]interface{}
		var spec kotsSingleSpec
		var allKotsReleaseSpecs []kotsSingleSpec

		for _, file := range files {
			bytes, err = ioutil.ReadFile(r.args.createReleaseYamlDir + "/" + file.Name())
			spec = kotsSingleSpec{"name": file.Name(), "path": file.Name(), "content": string(bytes)}
			allKotsReleaseSpecs = append(allKotsReleaseSpecs, spec)

			if err != nil {
				return err
			}
		}

		jsonAllYamls, err := json.Marshal(allKotsReleaseSpecs)

		if err != nil {
			return err
		}
		r.args.createReleaseYaml = string(jsonAllYamls)

	}

	// if the --promote param was used make sure it identifies exactly one
	// channel before proceeding
	var promoteChanID string
	if r.args.createReleasePromote != "" {
		channels, err := r.api.ListChannels(r.appID)
		if err != nil {
			return err
		}

		promoteChannelIDs := make([]string, 0)
		for _, c := range channels {
			if c.ID == r.args.createReleasePromote || c.Name == r.args.createReleasePromote {
				promoteChannelIDs = append(promoteChannelIDs, c.ID)
			}
		}

		if len(promoteChannelIDs) == 0 {
			return fmt.Errorf("Channel %q not found", r.args.createReleasePromote)
		}
		if len(promoteChannelIDs) > 1 {
			return fmt.Errorf("Channel %q is ambiguous. Please use channel ID", r.args.createReleasePromote)
		}
		promoteChanID = promoteChannelIDs[0]
	}

	release, err := r.api.CreateRelease(r.appID, r.args.createReleaseYaml)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(r.w, "SEQUENCE: %d\n", release.Sequence); err != nil {
		return err
	}
	r.w.Flush()

	if promoteChanID != "" {
		if err := r.api.PromoteRelease(
			r.appID,
			release.Sequence,
			r.args.createReleasePromoteVersion,
			r.args.createReleasePromoteNotes,
			r.args.createReleasePromoteRequired,
			promoteChanID,
		); err != nil {
			return err
		}

		// ignore error since operation was successful
		fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", promoteChanID, release.Sequence)
		r.w.Flush()
	}

	return nil
}
