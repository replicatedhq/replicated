package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMUpdateRBACPolicy(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "rbac-policy [ID_OR_NAME]",
		Hidden: true,
		Short:  "(alpha) Update the RBAC policy assigned to a VM.",
		Long: `(alpha) The 'rbac-policy' command assigns or removes the RBAC policy on a running VM.

When a policy is assigned, the VM's OIDC client credentials are used by the replicated CLI
inside the VM to authenticate with vendor-api automatically using that policy's permissions.
Pass an empty string to '--rbac-policy-name' to remove the policy from the VM.

Note: this feature is currently in alpha and requires the cmx_vm_rbac feature flag to be enabled.`,
		Example: `# Assign an RBAC policy to a VM by VM ID
replicated vm update rbac-policy aaaaa11 --rbac-policy-name "Read Only"

# Assign an RBAC policy to a VM by VM name
replicated vm update rbac-policy my-test-vm --rbac-policy-name "Read Only"

# Remove the RBAC policy from a VM
replicated vm update rbac-policy my-test-vm --rbac-policy-name ""`,
		RunE:              r.updateVMRBACPolicy,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeVMIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateVMRBACPolicyName, "rbac-policy-name", "", "(alpha) Name of the RBAC policy to assign to the VM (pass empty string to remove)")
	cmd.MarkFlagRequired("rbac-policy-name")

	return cmd
}

func (r *runners) updateVMRBACPolicy(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		vmID, err := r.getVMIDFromArg(args[0])
		if err != nil {
			return errors.Wrap(err, "get vm id from arg")
		}
		r.args.updateVMID = vmID
	} else if err := r.ensureUpdateVMIDArg(args); err != nil {
		return errors.Wrap(err, "ensure vm id arg")
	}

	var policyID string
	if r.args.updateVMRBACPolicyName != "" {
		p, err := r.kotsAPI.GetPolicyByName(r.args.updateVMRBACPolicyName)
		if err != nil {
			return errors.Wrap(err, "get rbac policy")
		}
		policyID = p.ID
	}

	if err := r.kotsAPI.UpdateVMRBACPolicy(r.args.updateVMID, policyID); err != nil {
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		}
		return errors.Wrap(err, "update vm rbac policy")
	}

	fmt.Fprintln(r.w, "RBAC policy updated.")
	return nil
}
