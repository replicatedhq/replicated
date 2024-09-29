package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMNodeGroupList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ls [ID]",
		Short:             "List node groups for a vm",
		Long:              `List node groups for a vm`,
		Args:              cobra.ExactArgs(1),
		RunE:              r.listVMNodeGroups,
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	return cmd
}

func (r *runners) listVMNodeGroups(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("vm id is required")
	}
	vmID := args[0]

	vm, err := r.kotsAPI.GetVM(vmID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "get vm")
	}

	return print.NodeGroups(r.outputFormat, r.w, vm.NodeGroups)
}
