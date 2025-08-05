package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID_OR_NAME [ID_OR_NAME …]",
		Aliases: []string{"delete"},
		Short:   "Remove test VM(s) immediately, with options to filter by name, tag, or remove all VMs.",
		Long: `The 'rm' command allows you to remove test VMs from your account immediately. You can specify one or more VM IDs or names directly, or use flags to filter which VMs to remove based on their tags, or simply remove all VMs at once.

This command supports multiple filtering options, including removing VMs by their name or ID, by specific tags, or by specifying the '--all' flag to remove all VMs in your account.

When specifying a name that matches multiple VMs, all VMs with that name will be removed.

You can also use the '--dry-run' flag to simulate the removal without actually deleting the VMs.

VMs are currently a beta feature.`,
		Example: `# Remove a VM by ID or name
replicated vm rm aaaaa11

# Remove multiple VMs by ID or name
replicated vm rm aaaaa11 bbbbb22 ccccc33

# Remove all VMs with a specific tag
replicated vm rm --tag env=dev

# Remove all VMs
replicated vm rm --all

# Perform a dry run of removing all VMs
replicated vm rm --all --dry-run`,
		RunE:              r.removeVMs,
		ValidArgsFunction: r.completeVMIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringArrayVar(&r.args.removeVMNames, "name", []string{}, "Name of the vm to remove (can be specified multiple times)")
	cmd.RegisterFlagCompletionFunc("name", r.completeVMNames)
	cmd.Flag("name").Deprecated = "use ID_OR_NAME arguments instead"
	cmd.Flags().StringArrayVar(&r.args.removeVMTags, "tag", []string{}, "Tag of the vm to remove (key=value format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.removeVMAll, "all", false, "remove all vms")

	cmd.Flags().BoolVar(&r.args.removeVMDryRun, "dry-run", false, "Dry run")

	return cmd
}

func (r *runners) removeVMs(_ *cobra.Command, args []string) error {
	if len(args) == 0 && !r.args.removeVMAll && len(r.args.removeVMNames) == 0 && len(r.args.removeVMTags) == 0 {
		return errors.New("One of ID, --all, --name or --tag flag required")
	} else if len(args) > 0 && (r.args.removeVMAll || len(r.args.removeVMNames) > 0 || len(r.args.removeVMTags) > 0) {
		return errors.New("cannot specify ID and --all, --name or --tag flag")
	} else if len(args) == 0 && r.args.removeVMAll && (len(r.args.removeVMNames) > 0 || len(r.args.removeVMTags) > 0) {
		return errors.New("cannot specify --all and --name or --tag flag")
	} else if len(args) == 0 && !r.args.removeVMAll && len(r.args.removeVMNames) > 0 && len(r.args.removeVMTags) > 0 {
		return errors.New("cannot specify --name and --tag flag")
	}

	if len(r.args.removeVMNames) > 0 {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list vms")
		}
		for _, vm := range vms {
			for _, name := range r.args.removeVMNames {
				if vm.Name == name {
					err := removeVM(r, vm.ID)
					if err != nil {
						return errors.Wrap(err, "remove vm")
					}
				}
			}
		}
	}

	if len(r.args.removeVMTags) > 0 {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list vms")
		}
		tags, err := parseTags(r.args.removeVMTags)
		if err != nil {
			return errors.Wrap(err, "parse tags")
		}

		for _, vm := range vms {
			if len(vm.Tags) > 0 {
				for _, tag := range tags {
					for _, vmTag := range vm.Tags {
						if vmTag.Key == tag.Key && vmTag.Value == tag.Value {
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

	if r.args.removeVMAll {
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
		_, err := r.kotsAPI.GetVM(arg)
		if err == nil {
			err := removeVM(r, arg)
			if err != nil {
				return errors.Wrap(err, "remove vm")
			}
			continue
		}

		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list vms")
		}

		found := false
		for _, vm := range vms {
			if vm.Name == arg {
				found = true
				err := removeVM(r, vm.ID)
				if err != nil {
					return errors.Wrap(err, "remove vm")
				}
			}
		}

		if !found {
			return errors.Errorf("VM with name or ID '%s' not found", arg)
		}
	}

	return nil
}

func removeVM(r *runners, vmID string) error {
	if r.args.removeVMDryRun {
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
