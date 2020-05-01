package cmd

import (
	"encoding/json"
	"fmt"
	yaml "github.com/replicatedhq/yaml/v3"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
	kotskinds "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
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



	bytes, err := ioutil.ReadFile(chartYamlPath)
	if err != nil {
		return err
	}

	//fmt.Printf("%s\n", bytes)

	//fmt.Printf("%v\n", args[0])

	chartYaml := ChartYaml{}
	yaml.Unmarshal(bytes, &chartYaml)

	fmt.Printf("%v\n", chartYaml)

	// create helm chart kots kind

	//fmt.Printf("%v\n", kotsHelmCrd)

	kotsHelmCrd := kotskinds.HelmChart{
		Spec: kotskinds.HelmChartSpec{
			Chart: kotskinds.ChartIdentifier{
				ChartVersion: chartYaml.Version,
			},
		},
	}

	bytes, err = json.MarshalIndent(kotsHelmCrd, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", bytes)

	return nil
}
