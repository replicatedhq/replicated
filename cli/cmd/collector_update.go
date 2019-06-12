package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "update SPEC_ID",
		Short: "Updated a collectors's yaml config",
		Long:  "Updated a collectors's yaml config",
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateCollectorYaml, "yaml", "", "The new YAML config for this collector. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.updateCollectorYamlFile, "yaml-file", "", "The file name with YAML config for this collector.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.updateCollectorName, "name", "", "The name for this collector")

	cmd.RunE = r.collectorUpdate
}

func (r *runners) collectorUpdate(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return errors.New("collector spec ID is required")
	}
	specID := args[0]

	if r.args.updateCollectorName == "" && r.args.updateCollectorYaml == "" && r.args.updateCollectorYamlFile == "" {
		return fmt.Errorf("name or yaml is required")
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

	if r.args.updateCollectorYaml != "" {
		_, err := r.api.UpdateCollector(r.appID, r.appType, specID, r.args.updateCollectorYaml)
		if err != nil {
			return fmt.Errorf("Failure setting updates for collector: %v", err)
		}
	}

	if r.args.updateCollectorName != "" {
		_, err := r.api.UpdateCollectorName(r.appID, r.appType, specID, r.args.updateCollectorName)
		if err != nil {
			return fmt.Errorf("Failure setting new yaml config for collector: %v", err)
		}
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Collector %s updated\n", specID)
	r.w.Flush()

	return nil
}
