package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryAdd(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "add",
		Short:        "add",
		Long:         `add`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().BoolVar(&r.args.addRegistrySkipValidation, "skip-validation", false, "Skip validation of the registry (not recommended)")

	return cmd
}

func contains(vals []string, val string) bool {
	for _, v := range vals {
		if v == val {
			return true
		}
	}
	return false
}

func (r *runners) readPasswordFromStdIn(label string) (string, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s:", label),
		Mask:      '*',
		Templates: templates,
		Default:   "",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("password cannot be empty")
			}
			return nil
		},
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return "", errors.New("interrupted")
			}
			continue
		}

		return result, nil
	}
}
