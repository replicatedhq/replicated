package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "update SPEC_ID",
		Short: "Update a collector's name or yaml config",
		Long:  "Update a collector's name or yaml config",
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateCollectorYaml, "yaml", "", "The new YAML config for this collector. Use '-' to read from stdin. Cannot be used with the --yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.updateCollectorYamlFile, "yaml-file", "", "The file name with YAML config for this collector. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.updateCollectorName, "name", "", "The name for this collector")

	cmd.RunE = r.collectorUpdate
}

func (r *runners) collectorUpdate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("collector spec ID is required")
	}
	specID := args[0]

	if r.args.updateCollectorName == "" && r.args.updateCollectorYaml == "" && r.args.updateCollectorYamlFile == "" {
		return errors.New("one of --name, --yaml or --yaml-file is required")
	}

	if r.args.updateCollectorYaml != "" && r.args.updateCollectorYamlFile != "" {
		return errors.New("only one of --yaml or --yaml-file may be specified")
	}

	if (strings.HasSuffix(r.args.updateCollectorYaml, ".yaml") || strings.HasSuffix(r.args.updateCollectorYaml, ".yml")) &&
		len(strings.Split(r.args.updateCollectorYaml, " ")) == 1 {
		return errors.New("use the --yaml-file flag when passing a yaml filename")
	}

	if r.args.updateCollectorYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		r.args.updateCollectorYaml = string(bytes)
	}

	if r.args.updateCollectorYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.updateCollectorYamlFile)
		if err != nil {
			return err
		}
		r.args.updateCollectorYaml = string(bytes)
	}

	if r.args.updateCollectorYaml != "" {
		_, err := r.api.UpdateCollector(r.appID, specID, r.args.updateCollectorYaml)
		if err != nil {
			return errors.Wrap(err, "failure setting updates for collector")
		}
	}

	if r.args.updateCollectorName != "" {
		_, err := r.api.UpdateCollectorName(r.appID, specID, r.args.updateCollectorName)
		if err != nil {
			return errors.Wrap(err, "failure setting new yaml config for collector")
		}
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Collector %s updated\n", specID)
	r.w.Flush()

	return nil
}
