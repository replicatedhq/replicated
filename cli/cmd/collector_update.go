package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "update NAMEE",
		Short: "Updated a collectors's yaml config",
		Long:  "Updated a collectors's yaml config",
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateCollectorYaml, "yaml", "", "The new YAML config for this release. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.updateCollectorYamlFile, "yaml-file", "", "The file name with YAML config for this release.  Cannot be used with the `yaml` flag.")

	cmd.RunE = r.collectorUpdate
}

func (r *runners) collectorUpdate(cmd *cobra.Command, args []string) error {
	if r.args.updateCollectorYaml == "" && r.args.updateCollectorYamlFile == "" {
		return fmt.Errorf("yaml is required")
	}

	if r.args.updateCollectorYaml != "" && r.args.updateCollectorYamlFile != "" {
		return fmt.Errorf("only yaml or yaml-file has to be specified")
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

	if len(args) < 1 {
		return errors.New("release sequence is required")
	}
	name := args[0]

	if err := r.platformAPI.UpdateCollector(r.appID, name, r.args.updateCollectorYaml); err != nil {
		return fmt.Errorf("Failure setting new yaml config for release: %v", err)
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Collector %s updated\n", name)
	r.w.Flush()

	return nil
}
