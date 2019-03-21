package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

var lintReleaseYaml string
var lintReleaseYamlFile string

var releaseLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint a YAML",
	Long:  "Lint a YAML",
}

func init() {
	releaseCmd.AddCommand(releaseLintCmd)

	releaseLintCmd.Flags().StringVar(&lintReleaseYaml, "yaml", "", "The YAML config to lint. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	releaseLintCmd.Flags().StringVar(&lintReleaseYamlFile, "yaml-file", "", "The file name with YAML config to lint.  Cannot be used with the `yaml` flag.")
}

func (r *runners) releaseLint(cmd *cobra.Command, args []string) error {
	if lintReleaseYaml == "" && lintReleaseYamlFile == "" {
		return fmt.Errorf("yaml is required")
	}

	if lintReleaseYaml != "" && lintReleaseYamlFile != "" {
		return fmt.Errorf("only yaml or yaml-file has to be specified")
	}

	if lintReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		lintReleaseYaml = string(bytes)
	}

	if lintReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(lintReleaseYamlFile)
		if err != nil {
			return err
		}
		lintReleaseYamlFile = string(bytes)
	}

	lintResult, err := r.api.LintRelease(r.appID, lintReleaseYaml)
	if err != nil {
		return err
	}

	if err := print.LintErrors(r.w, lintResult); err != nil {
		return err
	}

	return nil
}
