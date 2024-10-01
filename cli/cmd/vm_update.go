package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update vm settings",
		Long:  `vm update can be used to update vm settings`,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&r.args.updateClusterName, "name", "", "Name of the vm to update.")
	cmd.RegisterFlagCompletionFunc("name", r.completeVMNames)

	cmd.PersistentFlags().StringVar(&r.args.updateClusterID, "id", "", "id of the vm to update (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeVMIDs)

	return cmd
}

func (r *runners) ensureUpdateVMIDArg(args []string) error {
	// by default, we look at args[0] as the id
	// but if it's not provided, we look for a viper flag named "name" and use it
	// as the name of the vm, not the id
	if len(args) > 0 {
		r.args.updateClusterID = args[0]
	} else if r.args.updateClusterName != "" {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list vms")
		}
		for _, vm := range vms {
			if vm.Name == r.args.updateClusterName {
				r.args.updateClusterID = vm.ID
				break
			}
		}
	} else if r.args.updateClusterID != "" {
		// do nothing
		// but this is here for readability
	} else {
		return errors.New("must provide vm id or name")
	}

	return nil
}
