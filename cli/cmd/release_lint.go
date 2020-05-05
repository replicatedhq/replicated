package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
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

// releaseLint uses the replicatedhq/kots-lint service. This currently uses
// the hosted version (lint.replicated.com). There are not changes and no auth required or sent.
// This could be vendored in and run locally (respecting the size of the polcy files)
func (r *runners) releaseLint(cmd *cobra.Command, args []string) error {
	if r.args.lintReleaseYamlDir == "" {
		return errors.Errorf("yaml is required")
	}

	if _, ok := validFailOnValues[r.args.lintReleaseFailOn]; !ok {
		return errors.Errorf("fail-on value %q not supported, supported values are [info, warn, error, none]", r.args.lintReleaseFailOn)
	}

	lintReleaseYAML, err := tarYAMLDir(r.args.lintReleaseYamlDir)
	if err != nil {
		return errors.Wrap(err, "failed to read yaml dir")
	}

	lintResult, err := r.api.LintRelease(r.appType, lintReleaseYAML)
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

func tarYAMLDir(yamlDir string) ([]byte, error) {
	archiveDir, err := ioutil.TempDir("", "replicated")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp dir for archive")
	}
	defer os.RemoveAll(archiveDir)

	archiveFile := filepath.Join(archiveDir, "kots-release.tar")

	tar := archiver.Tar{
		ImplicitTopLevelFolder: true,
	}

	if err := tar.Archive([]string{yamlDir}, archiveFile); err != nil {
		return nil, errors.Wrap(err, "failed to archive")
	}

	data, err := ioutil.ReadFile(archiveFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read archive file")
	}

	return data, nil
}
