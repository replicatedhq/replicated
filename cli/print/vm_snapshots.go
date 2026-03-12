package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"
	"github.com/replicatedhq/replicated/pkg/types"
)

var vmSnapshotFuncs = template.FuncMap{
	"shortID": func(id string) string {
		if len(id) > 8 {
			return id[:8]
		}
		return id
	},
}

var vmSnapshotsTmplTableHeaderSrc = `ID	VM ID	NAME	STATUS	CREATED	EXPIRES`
var vmSnapshotsTmplTableRowSrc = `{{ range . -}}
{{ shortID .ID }}	{{ shortID .VMID }}	{{ if .Name }}{{ padding .Name 27 }}{{ else }}{{ padding "-" 27 }}{{ end }}	{{ padding .Status 12 }}	{{ padding (printf "%s" (localeTime .CreatedAt)) 22 }}	{{ if .ExpiresAt.IsZero }}{{ padding "-" 22 }}{{ else }}{{ padding (printf "%s" (localeTime .ExpiresAt)) 22 }}{{ end }}
{{ end }}`

var vmSnapshotsTmplTableSrc = fmt.Sprintln(vmSnapshotsTmplTableHeaderSrc) + vmSnapshotsTmplTableRowSrc
var vmSnapshotsTmplTable = template.Must(template.New("vmSnapshots").Funcs(funcs).Funcs(vmSnapshotFuncs).Parse(vmSnapshotsTmplTableSrc))
var vmSnapshotsTmplTableNoHeader = template.Must(template.New("vmSnapshots").Funcs(funcs).Funcs(vmSnapshotFuncs).Parse(vmSnapshotsTmplTableRowSrc))

func VMSnapshots(outputFormat string, w *tabwriter.Writer, snapshots []*types.VMSnapshot, header bool) error {
	switch outputFormat {
	case "table", "wide":
		if header {
			if err := vmSnapshotsTmplTable.Execute(w, snapshots); err != nil {
				return err
			}
		} else {
			if err := vmSnapshotsTmplTableNoHeader.Execute(w, snapshots); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(snapshots, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func VMSnapshot(outputFormat string, w *tabwriter.Writer, snapshot *types.VMSnapshot) error {
	return VMSnapshots(outputFormat, w, []*types.VMSnapshot{snapshot}, true)
}

func NoVMSnapshots(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table", "wide":
		_, err := fmt.Fprintln(w, "No snapshots found. Use the `replicated vm snapshot create` command to create a new snapshot.")
		if err != nil {
			return err
		}
	case "json":
		if _, err := fmt.Fprintln(w, "[]"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}
