package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm ID [ID …]",
		Short: "Remove test VM",
		Long: `Removes a VM immediately.

You can specify the --all flag to terminate all vms.`,
		RunE:              r.removeVMs,
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringArrayVar(&r.args.removeClusterNames, "name", []string{}, "Name of the vm to remove (can be specified multiple times)")
	cmd.RegisterFlagCompletionFunc("name", r.completeVMNames)
	cmd.Flags().StringArrayVar(&r.args.removeClusterTags, "tag", []string{}, "Tag of the vm to remove (key=value format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.removeClusterAll, "all", false, "remove all vms")

	cmd.Flags().BoolVar(&r.args.removeClusterDryRun, "dry-run", false, "Dry run")

	return cmd
}

func (r *runners) removeVMs(_ *cobra.Command, args []string) error {
	if len(args) == 0 && !r.args.removeClusterAll && len(r.args.removeClusterNames) == 0 && len(r.args.removeClusterTags) == 0 {
		return errors.New("One of ID, --all, --name or --tag flag required")
	} else if len(args) > 0 && (r.args.removeClusterAll || len(r.args.removeClusterNames) > 0 || len(r.args.removeClusterTags) > 0) {
		return errors.New("cannot specify ID and --all, --name or --tag flag")
	} else if len(args) == 0 && r.args.removeClusterAll && (len(r.args.removeClusterNames) > 0 || len(r.args.removeClusterTags) > 0) {
		return errors.New("cannot specify --all and --name or --tag flag")
	} else if len(args) == 0 && !r.args.removeClusterAll && len(r.args.removeClusterNames) > 0 && len(r.args.removeClusterTags) > 0 {
		return errors.New("cannot specify --name and --tag flag")
	}

	if len(r.args.removeClusterNames) > 0 {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list vms")
		}
		for _, vm := range vms {
			for _, name := range r.args.removeClusterNames {
				if vm.Name == name {
					err := removeVM(r, vm.ID)
					if err != nil {
						return errors.Wrap(err, "remove vm")
					}
				}
			}
		}
	}

	if len(r.args.removeClusterTags) > 0 {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list vms")
		}
		tags, err := parseTags(r.args.removeClusterTags)
		if err != nil {
			return errors.Wrap(err, "parse tags")
		}

		for _, vm := range vms {
			if vm.Tags != nil && len(vm.Tags) > 0 {
				for _, tag := range tags {
					for _, clusterTag := range vm.Tags {
						if clusterTag.Key == tag.Key && clusterTag.Value == tag.Value {
							err := removeVM(r, vm.ID)
							if err != nil {
								return errors.Wrap(err, "remove vm")
							}
						}
					}
				}
			}
		}
	}

	if r.args.removeClusterAll {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list vms")
		}
		for _, vm := range vms {
			err := removeVM(r, vm.ID)
			if err != nil {
				return errors.Wrap(err, "remove vm")
			}
		}
	}

	for _, arg := range args {
		err := removeVM(r, arg)
		if err != nil {
			return errors.Wrap(err, "remove vm")
		}
	}

	return nil
}

func removeVM(r *runners, vmID string) error {
	if r.args.removeClusterDryRun {
		fmt.Printf("would remove vm %s\n", vmID)
		return nil
	}
	err := r.kotsAPI.RemoveVM(vmID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "remove vm")
	} else {
		fmt.Printf("removed vm %s\n", vmID)
	}
	return nil
}
