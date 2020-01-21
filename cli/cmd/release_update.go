package cmd

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "update SEQUENCE",
		Short: "Updated a release's yaml config",
		Long:  "Updated a release's yaml config",
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateReleaseYaml, "yaml", "", "The new YAML config for this release. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")

	cmd.RunE = r.releaseUpdate
}

func (r *runners) releaseUpdate(cmd *cobra.Command, args []string) error {
	if r.args.updateReleaseYaml == "" && r.args.updateReleaseYamlFile == "" && r.args.updateReleaseYamlDir == "" {
		return fmt.Errorf("yaml is required")
	}

	if r.args.updateReleaseYaml != "" && r.args.updateReleaseYamlFile != "" {
		return fmt.Errorf("only yaml or yaml-file has to be specified")
	}

	if r.args.updateReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		r.args.updateReleaseYaml = string(bytes)
	}

	if r.args.updateReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.updateReleaseYamlFile)
		if err != nil {
			return err
		}
		r.args.updateReleaseYaml = string(bytes)
	}

	if len(args) < 1 {
		return errors.New("release sequence is required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid release sequence: %s", args[0])
	}

	if r.args.updateReleaseYamlDir != "" {
		r.args.updateReleaseYaml, err = readYAMLDir(r.args.createReleaseYamlDir)
		if err != nil {
			return errors.Wrap(err, "read yaml dir")
		}
	}
	if err := r.api.UpdateRelease(r.appID, r.appType, seq, r.args.updateReleaseYaml); err != nil {
		return fmt.Errorf("Failure setting new yaml config for release: %v", err)
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Release %d updated\n", seq)
	r.w.Flush()

	return nil
}
