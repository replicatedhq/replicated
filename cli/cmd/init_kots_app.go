package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
	yaml "github.com/replicatedhq/yaml/v3"
)

func (r *runners) InitInitKotsApp(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "init-kots-app DIRECTORY",
		Short: "Print the YAML config for a release",
		Long:  "Print the YAML config for a release",
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)
	cmd.RunE = r.initKotsApp
}

type ChartYaml struct {

	Name string
	Version string

}

func (r *runners) initKotsApp(_ *cobra.Command, args []string) error {

	chart_yaml_path := filepath.Join(args[0], "Chart.yaml")

	chart_yaml := ChartYaml{}

	bytes, err := ioutil.ReadFile(chart_yaml_path)

	fmt.Printf("%s\n", bytes)

	fmt.Printf("%v\n", args[0])



	yaml.Unmarshal(bytes, &chart_yaml)

	fmt.Printf("%v\n", chart_yaml)

	return err
}
