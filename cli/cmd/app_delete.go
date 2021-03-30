package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
	"os"
)

func (r *runners) InitAppDelete(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete NAME",
		Short:        "delete kots apps",
		Long:         `Delete a kots app. There is no undo for this operation, use with caution.`,
		RunE:         r.deleteApp,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().BoolVarP(&r.args.deleteAppForceYes, "force", "f", false, "Skip confirmation prompt. There is no undo for this action.")

	return cmd
}

func (r *runners) deleteApp(_ *cobra.Command, args []string) error {
	log := print.NewLogger(r.w)
	if len(args) != 1 {
		return errors.New("missing app slug or id")
	}
	appName := args[0]

	log.ActionWithSpinner("Fetching App")
	app, err := r.kotsAPI.GetApp(appName)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "list apps")
	}
	log.FinishSpinner()

	apps := []types.AppAndChannels{{ App: app}}

	err = print.Apps(r.w, apps)
	if err != nil {
		return errors.Wrap(err, "print app")
	}

	if !r.args.deleteAppForceYes {
		answer, err := promptConfirmDelete()
		if err != nil {
			return errors.Wrap(err, "confirm deletion")
		}

		if answer != "yes" {
			return errors.New("prompt declined")
		}
	}


	log.ActionWithSpinner("Deleting App")
	err = r.kotsAPI.DeleteKOTSApp(app.ID)
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
			if input == "no" {
				return nil
			}

			if input != "yes" {
				return errors.New(`only "yes" will be accepted`)
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
