package cmd

import (
	"reflect"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List test vms",
		Long:  `List test vms`,
		RunE:  r.listVMs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.lsClusterShowTerminated, "show-terminated", false, "when set, only show terminated vms")
	cmd.Flags().StringVar(&r.args.lsClusterStartTime, "start-time", "", "start time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.args.lsClusterEndTime, "end-time", "", "end time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	cmd.Flags().BoolVarP(&r.args.lsClusterWatch, "watch", "w", false, "watch vms")

	return cmd
}

func (r *runners) listVMs(_ *cobra.Command, args []string) error {
	const longForm = "2006-01-02T15:04:05Z"
	var startTime, endTime *time.Time
	if r.args.lsClusterStartTime != "" {
		st, err := time.Parse(longForm, r.args.lsClusterStartTime)
		if err != nil {
			return errors.Wrap(err, "parse start time")
		}
		startTime = &st
	}
	if r.args.lsClusterEndTime != "" {
		et, err := time.Parse(longForm, r.args.lsClusterEndTime)
		if err != nil {
			return errors.Wrap(err, "parse end time")
		}
		endTime = &et
	}

	vms, err := r.kotsAPI.ListVMs(r.args.lsClusterShowTerminated, startTime, endTime)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list vms")
	}

	header := true
	if r.args.lsClusterWatch {

		// Checks to see if the outputFormat is table
		if r.outputFormat != "table" && r.outputFormat != "wide" {
			return errors.New("watch is only supported for table output")
		}

		vmsToPrint := make([]*types.VM, 0)

		// Prints the intial list of vms
		if len(vms) == 0 {
			print.NoVMs(r.outputFormat, r.w)
		} else {
			vmsToPrint = append(vmsToPrint, vms...)
		}

		// Runs until ctrl C is recognized
		for range time.Tick(2 * time.Second) {
			newVMs, err := r.kotsAPI.ListVMs(r.args.lsClusterShowTerminated, startTime, endTime)

			if err != nil {
				if err == promptui.ErrInterrupt {
					return errors.New("interrupted")
				}

				return errors.Wrap(err, "watch vms")
			}

			// Create a map from the IDs of the new vms
			newVMMap := make(map[string]*types.VM)
			for _, newVM := range newVMs {
				newVMMap[newVM.ID] = newVM
			}

			// Create a map from the IDs of the old vms
			oldVMMap := make(map[string]*types.VM)
			for _, vm := range vms {
				oldVMMap[vm.ID] = vm
			}

			// Check for new vms and print them
			for id, newVM := range newVMMap {
				if oldVM, found := oldVMMap[id]; !found {
					vmsToPrint = append(vmsToPrint, newVM)
				} else {
					// Check if properties of existing vms have changed
					if !reflect.DeepEqual(newVM, oldVM) {
						vmsToPrint = append(vmsToPrint, newVM)
					}
				}
			}

			// Check for removed vms and print them, changing their status to be "deleted"
			for id, vm := range oldVMMap {
				if _, found := newVMMap[id]; !found {
					vm.Status = types.ClusterStatusDeleted
					vmsToPrint = append(vmsToPrint, vm)
				}
			}

			// Prints the vms
			if len(vmsToPrint) > 0 {
				print.VMs(r.outputFormat, r.w, vmsToPrint, header)
				header = false // only print the header once
			}

			vms = newVMs
			vmsToPrint = make([]*types.VM, 0)
		}
	}

	if len(vms) == 0 {
		return print.NoVMs(r.outputFormat, r.w)
	}

	return print.VMs(r.outputFormat, r.w, vms, true)
}
