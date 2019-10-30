package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type kotsSingleSpec struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Content  string   `json:"content"`
	Children []string `json:"children"`
}

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
	cmd.Flags().StringVar(&r.args.createReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleasePromote, "promote", "", "Channel name or id to promote this release to")
	cmd.Flags().StringVar(&r.args.createReleasePromoteNotes, "release-notes", "", "When used with --promote <channel>, sets the **markdown** release notes")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteRequired, "required", false, "When used with --promote <channel>, marks this release as required during upgrades.")
	cmd.Flags().StringVar(&r.args.createReleasePromoteVersion, "version", "", "When used with --promote <channel>, sets the version label for the release in this channel")

	cmd.RunE = r.releaseCreate
}

func (r *runners) releaseCreate(cmd *cobra.Command, args []string) error {
	if r.args.createReleaseYaml == "" && r.args.createReleaseYamlFile == "" && r.args.createReleaseYamlDir == "" {
		return errors.New("yaml is required")
	}

	if r.args.createReleaseYaml != "" && r.args.createReleaseYamlFile != "" {
		return errors.New("only one of yaml or yaml-file may be specified")
	}

	if r.args.createReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return errors.Wrap(err, "read from stdin")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.createReleaseYamlFile)
		if err != nil {
			return errors.Wrap(err, "read file yaml")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlDir != "" {
		var allKotsReleaseSpecs []kotsSingleSpec
		err := filepath.Walk(r.args.createReleaseYamlDir, func(path string, info os.FileInfo, err error) error {

			singlefile := strings.TrimPrefix(path, r.args.createReleaseYamlDir)

			if err != nil {
				return errors.Wrapf(err, "walk %s", info.Name())
			}

			if info.IsDir() {
				return nil
			}
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			ext := filepath.Ext(info.Name())
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}

			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "read file %s", path)
			}

			spec := kotsSingleSpec{
				Name:     info.Name(),
				Path:     singlefile,
				Content:  string(bytes),
				Children: []string{},
			}
			allKotsReleaseSpecs = append(allKotsReleaseSpecs, spec)
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "walk %s", r.args.createReleaseYamlDir)
		}

		jsonAllYamls, err := json.Marshal(allKotsReleaseSpecs)
		if err != nil {
			return errors.Wrap(err, "marshal spec")
		}
		r.args.createReleaseYaml = string(jsonAllYamls)
	}

	// if the --promote param was used make sure it identifies exactly one
	// channel before proceeding
	var promoteChanID string
	if r.args.createReleasePromote != "" {
		channels, err := r.api.ListChannels(r.appID, r.appType)
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

	release, err := r.api.CreateRelease(r.appID, r.appType, r.args.createReleaseYaml)
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
			r.appType,
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
