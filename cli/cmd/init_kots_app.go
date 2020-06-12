package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/kots/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/initkotsapp"
	"github.com/replicatedhq/yaml/v3"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func (r *runners) InitInitKotsApp(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "init-kots-app DIRECTORY",
		Short:        "Print the YAML config for a release",
		Long:         "Print the YAML config for a release",
		Hidden:       true,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.RunE = r.initKotsApp

	cmd.Flags().BoolVar(&r.args.initKotsAppSkipPrompt, "skip-prompt", false, "Skip Helm Chart name prompt.")
}

func (r *runners) initKotsApp(_ *cobra.Command, args []string) error {

	baseDirectory := args[0]
	chartYamlPath := filepath.Join(baseDirectory, "Chart.yaml")

	log := logger.NewLogger()

	log.ActionWithSpinner("Reading %s", chartYamlPath)

	bytes, err := ioutil.ReadFile(chartYamlPath)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "read Chart.yaml file")
	}
	time.Sleep(1 * time.Second)
	log.FinishSpinner()

	chartYaml := initkotsapp.ChartYaml{}
	yaml.Unmarshal(bytes, &chartYaml)

	appName := chartYaml.Name

	if !r.args.initKotsAppSkipPrompt {
		appName, err = promptForAppName(chartYaml.Name)
		if err != nil {
			return errors.Wrap(err, "prompt for app name")

		}
	}

	// prepare kots directory
	kotsBasePath := filepath.Join(baseDirectory, "kots")
	kotsManifestsPath := filepath.Join(kotsBasePath, "manifests")

	err = os.MkdirAll(kotsManifestsPath, 0755)
	if err != nil {
		return errors.Wrap(err, "create kots manifests directory")
	}
	log.ActionWithoutSpinner("Writing Files to %s", kotsBasePath)
	log.ActionWithSpinner("Writing Makefile")

	// Makefile
	err = initkotsapp.MakeFile(kotsBasePath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing .gitignore")

	// .gitignore
	err = initkotsapp.Gitignore(kotsBasePath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing .helmignore")

	// .helmignore
	err = initkotsapp.Helmignore(baseDirectory)
	if err != nil {
		log.FinishSpinnerWithError()
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing %s.yaml", chartYaml.Name)

	// Helm Chart CRD
	err = initkotsapp.HelmChartFile(chartYaml, kotsManifestsPath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing replicated-app.yaml")
	// App CRD
	err = initkotsapp.AppFile(chartYaml, appName, kotsManifestsPath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing preflight.yaml")
	// Preflight CRD
	err = initkotsapp.PreflightFile(chartYaml, kotsManifestsPath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing config.yaml")
	// Config CRD
	err = initkotsapp.ConfigFile(chartYaml, kotsManifestsPath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	log.ActionWithSpinner("Writing support-bundle.yaml")
	// Support Bundle CRD
	err = initkotsapp.SupportBundleFile(kotsManifestsPath)
	if err != nil {
		return err
	}
	log.FinishSpinner()

	return nil
}

func promptForAppName(chartName string) (string, error) {

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | bold }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "Enter the app chartName to use",
		Templates: templates,
		Default:   chartName,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("invalid app name")
			}

			return nil
		},
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		return result, nil
	}
}
