package cmd

import (
	"context"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

type deleteAppOpts struct {
	force bool
}

func (r *runners) InitAppRm(parent *cobra.Command) *cobra.Command {
	opts := deleteAppOpts{}
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "rm NAME",
		Aliases: []string{"delete"},
		Short:   "Delete an application",
		Long: `Delete an application from your Replicated account.

This command allows you to permanently remove an application from your account.
Once deleted, the application and all associated data will be irretrievably lost.

Use this command with caution as there is no way to undo this operation.`,
		Example: `# Delete a app named "My App"
replicated app delete "My App"

# Delete an app and skip the confirmation prompt
replicated app delete "Another App" --force

# Delete an app and output the result in JSON format
replicated app delete "Custom App" --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if len(args) != 1 {
				return errors.New("missing app slug or id")
			}
			return r.deleteApp(ctx, cmd, args[0], opts, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt. There is no undo for this action.")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) deleteApp(ctx context.Context, cmd *cobra.Command, appName string, opts deleteAppOpts, outputFormat string) error {
	log := logger.NewLogger(r.w)

	log.ActionWithSpinner("Fetching App")
	app, err := r.kotsAPI.GetApp(ctx, appName, true)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "list apps")
	}
	log.FinishSpinner()

	apps := []types.AppAndChannels{
		{
			App: app,
		},
	}

	err = print.Apps(outputFormat, r.w, apps)
	if err != nil {
		return errors.Wrap(err, "print app")
	}

	if !opts.force {
		answer, err := promptConfirmDelete()
		if err != nil {
			return errors.Wrap(err, "confirm deletion")
		}

		if answer != "yes" {
			return errors.New("prompt declined")
		}
	}

	log.ActionWithSpinner("Deleting App")
	err = r.kotsAPI.DeleteKOTSApp(ctx, app.ID)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "delete app")
	}
	log.FinishSpinner()

	return nil
}

var templates = &promptui.PromptTemplates{
	Prompt:  "{{ . | bold }} ",
	Valid:   "{{ . | green }} ",
	Invalid: "{{ . | red }} ",
	Success: "{{ . | bold }} ",
}

func promptConfirmDelete() (string, error) {
	prompt := promptui.Prompt{
		Label:     "Delete the above listed application? There is no undo:",
		Templates: templates,
		Default:   "",
		Validate: func(input string) error {
			// "no" will exit with a "prompt declined" error, just in case they don't think to ctrl+c
			if input == "no" || input == "yes" {
				return nil
			}
			return errors.New(`only "yes" will be accepted`)
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
