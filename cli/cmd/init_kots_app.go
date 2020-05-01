package cmd

import (
	"fmt"
	yaml "github.com/replicatedhq/yaml/v3"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
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
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (r *runners) initKotsApp(_ *cobra.Command, args []string) error {

	chartYamlPath := filepath.Join(args[0], "Chart.yaml")

	chartYaml := ChartYaml{}

	bytes, err := ioutil.ReadFile(chartYamlPath)

	fmt.Printf("%s\n", bytes)

	fmt.Printf("%v\n", args[0])

	yaml.Unmarshal(bytes, &chartYaml)

	fmt.Printf("%v\n", chartYaml)

	// create helm chart kots kind

	return err
}
