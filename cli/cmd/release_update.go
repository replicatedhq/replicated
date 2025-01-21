package cmd

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

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

	cmd.Flags().StringVar(&r.args.updateReleaseYaml, "yaml", "", "The new YAML config for this release. Use '-' to read from stdin. Cannot be used with the --yaml-file flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseChart, "chart", "", "Helm chart to create the release from. Cannot be used with the --yaml, --yaml-file, or --yaml-dir flags.")

	cmd.Flags().MarkHidden("chart")

	cmd.RunE = r.releaseUpdate
}

func (r *runners) releaseUpdate(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	specSources := []string{
		r.args.updateReleaseYaml,
		r.args.updateReleaseYamlFile,
		r.args.updateReleaseYamlDir,
		r.args.updateReleaseChart,
	}

	numSources := 0
	for _, specSource := range specSources {
		if specSource != "" {
			numSources++
		}
	}

	if numSources == 0 {
		return errors.New("one of --yaml, --yaml-file, --yaml-dir, or --chart is required")
	}

	if numSources > 1 {
		return errors.New("only one of --yaml, --yaml-file, --yaml-dir, or --chart may be specified")
	}

	if (strings.HasSuffix(r.args.updateReleaseYaml, ".yaml") || strings.HasSuffix(r.args.updateReleaseYaml, ".yml")) &&
		len(strings.Split(r.args.updateReleaseYaml, " ")) == 1 {
		return errors.New("use the --yaml-file flag when passing a yaml filename")
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
		return errors.Errorf("invalid release sequence: %s", args[0])
	}

	if r.args.updateReleaseYamlDir != "" {
		r.args.updateReleaseYaml, err = makeReleaseFromDir(r.args.updateReleaseYamlDir)
		if err != nil {
			return errors.Wrap(err, "make release from dir")
		}
	} else if r.args.updateReleaseChart != "" {
		r.args.updateReleaseYaml, err = makeReleaseFromChart(r.args.updateReleaseChart)
		if err != nil {
			return errors.Wrap(err, "make release from chart")
		}
	}

	if err := r.api.UpdateRelease(r.appID, r.appType, seq, r.args.updateReleaseYaml); err != nil {
		return errors.Wrap(err, "failure setting new yaml config for release")
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Release %d updated\n", seq)
	r.w.Flush()

	return nil
}
