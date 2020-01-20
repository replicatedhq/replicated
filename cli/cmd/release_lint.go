package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseLint(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint a directory of KOTS manifests",
		Long: "Lint a directory of KOTS manifests",
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lintReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")

	cmd.RunE = r.releaseLint
}

func (r *runners) releaseLint(cmd *cobra.Command, args []string) error {
	if  r.args.lintReleaseYamlDir == "" {
		return fmt.Errorf("yaml is required")
	}

	lintReleaseYAML, err := readYAMLDir(r.args.lintReleaseYamlDir)
	if err != nil {
		return errors.Wrap(err, "read yaml dir")
	}

	lintResult, err := r.api.LintRelease(r.appID, r.appType, lintReleaseYAML)
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
