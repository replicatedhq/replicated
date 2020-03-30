package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func (r *runners) InitInstallerCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new installer spec",
		Long:  `Create a new installer spec by providing YAML configuration for a https://kurl.sh cluster.`,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createInstallerYaml, "yaml", "", "The YAML config for this installer. Use '-' to read from stdin.  Cannot be used with the `yaml-file` falg.")
	cmd.Flags().StringVar(&r.args.createInstallerYamlFile, "yaml-file", "", "The file name with YAML config for this installer.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createInstallerPromote, "promote", "", "Channel name or id to promote this installer to")
	cmd.Flags().BoolVar(&r.args.createInstallerPromoteEnsureChannel, "ensure-channel", false, "When used with --promote <channel>, will create the channel if it doesn't exist")

	cmd.RunE = r.installerCreate
}

func (r *runners) installerCreate(_ *cobra.Command, _ []string) error {
	if r.appType != "kots" {
		return errors.Errorf("Installer specs are only supported for KOTS applications, app %q has type %q", r.appID, r.appType)
	}

	if r.args.createInstallerYaml == "" &&
		r.args.createInstallerYamlFile == "" {
		return errors.New("one of --yaml, --yaml-file is required")
	}

	// can't ensure a channel if you didn't pass one
	if r.args.createInstallerPromoteEnsureChannel && r.args.createInstallerPromote == "" {
		return errors.New("cannot use the flag --ensure-channel without also using --promote <channel> ")
	}

	if r.args.createInstallerYaml != "" && r.args.createInstallerYamlFile != "" {
		return errors.New("only one of --yaml or --yaml-file may be specified")
	}

	if r.args.createInstallerYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return errors.Wrap(err, "read from stdin")
		}
		r.args.createInstallerYaml = string(bytes)
	}

	if r.args.createInstallerYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.createInstallerYamlFile)
		if err != nil {
			return errors.Wrap(err, "read file yaml")
		}
		r.args.createInstallerYaml = string(bytes)
	}

	// if the --promote param was used make sure it identifies exactly one
	// channel before proceeding
	var promoteChanID string
	if r.args.createInstallerPromote != "" {
		var err error
		promoteChanID, err = r.getOrCreateChannelForPromotion(
			r.args.createInstallerPromote,
			r.args.createInstallerPromoteEnsureChannel,
		)
		if err != nil {
			return errors.Wrapf(err, "get or create channel %q for promotion", promoteChanID)
		}
	}

	installerSpec, err := r.api.CreateInstaller(r.appID, r.appType, r.args.createInstallerYaml)
	if err != nil {
		return errors.Wrap(err, "create installer")
	}

	if _, err := fmt.Fprintf(r.w, "SEQUENCE: %d\n", installerSpec.Sequence); err != nil {
		return errors.Wrap(err, "print sequence to r.w")
	}
	r.w.Flush()

	// don't send a version label as its not really meaningful
	noVersionLabel := ""

	if promoteChanID != "" {
		if err := r.api.PromoteInstaller(
			r.appID,
			r.appType,
			installerSpec.Sequence,
			promoteChanID,
			noVersionLabel,
		); err != nil {
			return errors.Wrap(err, "promote installer")
		}

		// ignore error since operation was successful
		fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", promoteChanID, installerSpec.Sequence)
		r.w.Flush()
	}

	return nil
}
