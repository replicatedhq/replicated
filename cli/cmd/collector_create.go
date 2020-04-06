package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new collector",
		Long:  "Create a new collector by providing a name and YAML configuration",
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createCollectorYaml, "yaml", "", "The YAML config for this collector. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.createCollectorYamlFile, "yaml-file", "", "The file name with YAML config for this collector.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createCollectorName, "name", "", "The name for this collector")

	cmd.RunE = r.collectorCreate
}

func (r *runners) collectorCreate(cmd *cobra.Command, args []string) error {

	if r.args.createCollectorName == "" {
		return fmt.Errorf("collector name is required")
	}

	if r.args.createCollectorYaml == "" && r.args.createCollectorYamlFile == "" {
		return fmt.Errorf("yaml is required")
	}

	if r.args.createCollectorYaml != "" && r.args.createCollectorYamlFile != "" {
		return fmt.Errorf("only one of yaml or yaml-file may be specified")
	}

	if r.args.createCollectorYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		r.args.createCollectorYaml = string(bytes)
	}

	if r.args.createCollectorYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.createCollectorYamlFile)
		if err != nil {
			return err
		}
		r.args.createCollectorYaml = string(bytes)
	}

	_, err := r.api.CreateCollector(r.appID, r.appType, r.args.createCollectorName, r.args.createCollectorYaml)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Collector %s created\n", r.args.createCollectorName)
	r.w.Flush()

	return nil

}
