package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseLint(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint a YAML",
		Long:  "Lint a YAML",
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lintReleaseYaml, "yaml", "", "The YAML config to lint. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.lintReleaseYamlFile, "yaml-file", "", "The file name with YAML config to lint.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.lintReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")

	cmd.RunE = r.releaseLint
}

func (r *runners) releaseLint(cmd *cobra.Command, args []string) error {
	if r.args.lintReleaseYaml == "" && r.args.lintReleaseYamlFile == "" && r.args.lintReleaseYamlDir == "" {
		return fmt.Errorf("yaml is required")
	}

	if r.args.lintReleaseYaml != "" && r.args.lintReleaseYamlFile != "" {
		return fmt.Errorf("only yaml or yaml-file has to be specified")
	}

	if r.args.lintReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		r.args.lintReleaseYaml = string(bytes)
	}

	if r.args.lintReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.lintReleaseYamlFile)
		if err != nil {
			return err
		}
		r.args.lintReleaseYamlFile = string(bytes)
	}

	if r.args.lintReleaseYamlDir != "" {
		var err error
		r.args.lintReleaseYaml, err = readYAMLDir(r.args.lintReleaseYamlDir)
		if err != nil {
			return errors.Wrap(err, "read yaml dir")
		}
	}

	lintResult, err := r.api.LintRelease(r.appID, r.appType, r.args.lintReleaseYaml)
	if err != nil {
		return err
	}

	if err := print.LintErrors(r.w, lintResult); err != nil {
		return err
	}

	var hasError bool
	for _, msg := range lintResult {
		if msg.Type == "error" {
			hasError = true
			break
		}
	}

	if hasError {
		return errors.New("one or more errors found")
	}

	return nil
}
