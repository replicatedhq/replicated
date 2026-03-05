package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

var vmSnapshotFuncs = template.FuncMap{
	"shortID": func(id string) string {
		if len(id) > 8 {
			return id[:8]
		}
		return id
	},
	"formatBytes": func(b *int64) string {
		if b == nil {
			return "-"
		}
		const (
			kb = 1024
			mb = kb * 1024
			gb = mb * 1024
		)
		switch {
		case *b >= gb:
			return fmt.Sprintf("%.1f GB", float64(*b)/float64(gb))
		case *b >= mb:
			return fmt.Sprintf("%.1f MB", float64(*b)/float64(mb))
		case *b >= kb:
			return fmt.Sprintf("%.1f KB", float64(*b)/float64(kb))
		default:
			return fmt.Sprintf("%d B", *b)
		}
	},
	"optionalTime": func(t *time.Time) string {
		if t == nil || t.IsZero() {
			return "-"
		}
		return t.Local().Format("2006-01-02 15:04 MST")
	},
}

var vmSnapshotsTmplTableHeaderSrc = `ID	VM ID	STATUS	SIZE	CREATED	READY AT	ERROR`
var vmSnapshotsTmplTableRowSrc = `{{ range . -}}
{{ shortID .ID }}	{{ shortID .VMID }}	{{ padding .Status 12 }}	{{ formatBytes .SizeBytes }}	{{ localeTime .CreatedAt }}	{{ optionalTime .ReadyAt }}	{{ if .Error }}{{ .Error }}{{ else }}-{{ end }}
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
