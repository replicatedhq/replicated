package cmd

import (
	"reflect"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all VM snapshots.",
		Long: `The 'vm snapshot ls' command lists all snapshots for the team.

This command is useful for viewing existing snapshots, their status, and creation/expiry time. The output format can be customized using the --output flag.

You can use the '--watch' flag to monitor snapshot status continuously. This will refresh the list every 2 seconds.

VM snapshots are currently an alpha feature.`,
		Example: `# List all snapshots
replicated vm snapshot ls

# List snapshots in JSON format
replicated vm snapshot ls --output json

# Watch snapshot status changes in real-time
replicated vm snapshot ls --watch`,
		RunE: r.vmSnapshotList,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")
	cmd.Flags().BoolVarP(&r.args.vmSnapshotWatch, "watch", "w", false, "watch snapshots")

	return cmd
}

func (r *runners) vmSnapshotList(_ *cobra.Command, args []string) error {
	snapshots, err := r.kotsAPI.ListVMSnapshots()
	if err != nil {
		return errors.Wrap(err, "list vm snapshots")
	}

	header := true
	if r.args.vmSnapshotWatch {
		if r.outputFormat != "table" && r.outputFormat != "wide" {
			return errors.New("watch is only supported for table output")
		}

		snapshotsToPrint := make([]*types.VMSnapshot, 0)

		if len(snapshots) == 0 {
			print.NoVMSnapshots(r.outputFormat, r.w)
		} else {
			snapshotsToPrint = append(snapshotsToPrint, snapshots...)
		}

		for range time.Tick(2 * time.Second) {
			newSnapshots, err := r.kotsAPI.ListVMSnapshots()
			if err != nil {
				if err == promptui.ErrInterrupt {
					return errors.New("interrupted")
				}
				return errors.Wrap(err, "watch vm snapshots")
			}

			newMap := make(map[string]*types.VMSnapshot)
			for _, s := range newSnapshots {
				newMap[s.ID] = s
			}

			oldMap := make(map[string]*types.VMSnapshot)
			for _, s := range snapshots {
				oldMap[s.ID] = s
			}

			for id, newS := range newMap {
				if oldS, found := oldMap[id]; !found {
					snapshotsToPrint = append(snapshotsToPrint, newS)
				} else if !reflect.DeepEqual(newS, oldS) {
					snapshotsToPrint = append(snapshotsToPrint, newS)
				}
			}

			for id, s := range oldMap {
				if _, found := newMap[id]; !found {
					s.Status = "deleted"
					snapshotsToPrint = append(snapshotsToPrint, s)
				}
			}

			if len(snapshotsToPrint) > 0 {
				print.VMSnapshots(r.outputFormat, r.w, snapshotsToPrint, header)
				header = false
			}

			snapshots = newSnapshots
			snapshotsToPrint = make([]*types.VMSnapshot, 0)
		}
	}

	if len(snapshots) == 0 {
		return print.NoVMSnapshots(r.outputFormat, r.w)
	}

	return print.VMSnapshots(r.outputFormat, r.w, snapshots, true)
}
