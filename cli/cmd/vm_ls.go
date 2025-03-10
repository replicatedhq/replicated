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
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List test VMs and their status, with optional filters for start/end time and terminated VMs.",
		Long: `List all test VMs in your account, including their current status, distribution, version, and more. You can use optional flags to filter the output based on VM termination status, start time, or end time. This command can also watch the VM status in real-time.

By default, the command will return a table of all VMs, but you can switch to JSON or wide output formats for more detailed information. The command supports filtering to show only terminated VMs or to specify a time range for the query.

You can use the '--watch' flag to monitor VMs continuously. This will refresh the list of VMs every 2 seconds, displaying any updates in real-time, such as new VMs being created or existing VMs being terminated.

The command also allows you to customize the output format, supporting 'json', 'table', and 'wide' views for flexibility based on your needs.`,
		Example: `# List all active VMs
replicated vm ls

# List all VMs that were created after a specific start time
replicated vm ls --start-time 2024-10-01T00:00:00Z

# Show only terminated VMs
replicated vm ls --show-terminated

# Watch VM status changes in real-time
replicated vm ls --watch`,
		RunE: r.listVMs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.lsVMShowTerminated, "show-terminated", false, "when set, only show terminated vms")
	cmd.Flags().StringVar(&r.args.lsVMStartTime, "start-time", "", "start time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.args.lsVMEndTime, "end-time", "", "end time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide (default: table)")
	cmd.Flags().BoolVarP(&r.args.lsVMWatch, "watch", "w", false, "watch vms")

	return cmd
}

func (r *runners) listVMs(_ *cobra.Command, args []string) error {
	const longForm = "2006-01-02T15:04:05Z"
	var startTime, endTime *time.Time
	if r.args.lsVMStartTime != "" {
		st, err := time.Parse(longForm, r.args.lsVMStartTime)
		if err != nil {
			return errors.Wrap(err, "parse start time")
		}
		startTime = &st
	}
	if r.args.lsVMEndTime != "" {
		et, err := time.Parse(longForm, r.args.lsVMEndTime)
		if err != nil {
			return errors.Wrap(err, "parse end time")
		}
		endTime = &et
	}

	vms, err := r.kotsAPI.ListVMs(r.args.lsVMShowTerminated, startTime, endTime)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list vms")
	}

	header := true
	if r.args.lsVMWatch {

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
			newVMs, err := r.kotsAPI.ListVMs(r.args.lsVMShowTerminated, startTime, endTime)
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
					// reset EstimatedCost (as it is calculated on the fly and not stored in the API response)
					oldVM.EstimatedCost = 0
					if !reflect.DeepEqual(newVM, oldVM) {
						vmsToPrint = append(vmsToPrint, newVM)
					}
				}
			}

			// Check for removed vms and print them, changing their status to be "deleted"
			for id, vm := range oldVMMap {
				if _, found := newVMMap[id]; !found {
					vm.Status = types.VMStatusDeleted
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
