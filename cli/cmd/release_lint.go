package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

var (
	validFailOnValues = map[string]interface{}{
		"error": nil,
		"warn":  nil,
		"info":  nil,
		"none":  nil,
		"":      nil,
	}
)

func (r *runners) InitReleaseLint(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "lint",
		Short:        "Lint a directory of KOTS manifests",
		Long:         "Lint a directory of KOTS manifests",
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lintReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.lintReleaseFailOn, "fail-on", "error", "The minimum severity to cause the command to exit with a non-zero exit code. Supported values are [info, warn, error, none].")

	cmd.RunE = r.releaseLint
}

func (r *runners) releaseLint(cmd *cobra.Command, args []string) error {
	if r.args.lintReleaseYamlDir == "" {
		return errors.Errorf("yaml is required")
	}

	if _, ok := validFailOnValues[r.args.lintReleaseFailOn]; !ok {
		return errors.Errorf("fail-on value %q not supported, supported values are [info, warn, error, none]", r.args.lintReleaseFailOn)
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

	if hasError := shouldFail(lintResult, r.args.lintReleaseFailOn); hasError {
		return errors.Errorf("One or more errors of severity %q or higher were found", r.args.lintReleaseFailOn)
	}

	return nil
}

// this is not especially fancy but it will do
func shouldFail(lintResult []types.LintMessage, failOn string) bool {
	switch failOn {
	case "", "none":
		return false
	case "error":
		for _, msg := range lintResult {
			if msg.Type == "error" {
				return true
			}
		}
		return false
	case "warn":
		for _, msg := range lintResult {
			if msg.Type == "error" || msg.Type == "warn" {
				return true
			}
		}
		return false
	default:
		// "info" or anything else, fall through and fail if there's any messages at all
	}
	return len(lintResult) > 0
}
