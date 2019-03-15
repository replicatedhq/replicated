package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

var lintReleaseYaml string

var releaseLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint a YAML",
	Long:  "Lint a YAML",
}

func init() {
	releaseCmd.AddCommand(releaseLintCmd)

	releaseLintCmd.Flags().StringVar(&lintReleaseYaml, "yaml", "", "The YAML config to lint. Use '-' to read from stdin")
}

func (r *runners) releaseLint(cmd *cobra.Command, args []string) error {
	if lintReleaseYaml == "" {
		return fmt.Errorf("yaml is required")
	}

	if lintReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		lintReleaseYaml = string(bytes)
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
