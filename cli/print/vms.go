package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

// Table formatting
var vmsTmplTableHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	STATUS	CREATED	EXPIRES`
var vmsTmplTableRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .Distribution 12 }}	{{ padding .Version 10 }}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" .CreatedAt) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" .ExpiresAt) 30 }}{{end}}
{{ end }}`
var vmsTmplTableSrc = fmt.Sprintln(vmsTmplTableHeaderSrc) + vmsTmplTableRowSrc
var vmsTmplTable = template.Must(template.New("vms").Funcs(funcs).Parse(vmsTmplTableSrc))
var vmsTmplTableNoHeader = template.Must(template.New("vms").Funcs(funcs).Parse(vmsTmplTableRowSrc))

// Wide table formatting
var vmsTmplWideHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	STATUS	CREATED	EXPIRES	TOTAL NODES	NODEGROUPS	TAGS`
var vmsTmplWideRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .Distribution 12 }}	{{ padding .Version 10 }}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" .CreatedAt) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" .ExpiresAt) 30 }}{{end}}	{{$nodecount:=0}}{{ range $index, $ng := .NodeGroups}}{{$nodecount = add $nodecount $ng.NodeCount}}{{ end }}{{ padding (printf "%d" $nodecount) 11 }}	{{ len .NodeGroups}}	{{ range $index, $tag := .Tags }}{{if $index}}, {{end}}{{ $tag.Key }}={{ $tag.Value }}{{ end }}
{{ end }}`
var vmsTmplWideSrc = fmt.Sprintln(vmsTmplWideHeaderSrc) + vmsTmplWideRowSrc
var vmsTmplWide = template.Must(template.New("vms").Funcs(funcs).Parse(vmsTmplWideSrc))
var vmsTmplWideNoHeader = template.Must(template.New("vms").Funcs(funcs).Parse(vmsTmplWideRowSrc))

func VMs(outputFormat string, w *tabwriter.Writer, vms []*types.VM, header bool) error {
	switch outputFormat {
	case "table":
		if header {
			if err := vmsTmplTable.Execute(w, vms); err != nil {
				return err
			}
		} else {
			if err := vmsTmplTableNoHeader.Execute(w, vms); err != nil {
				return err
			}
		}
	case "wide":
		if header {
			if err := vmsTmplWide.Execute(w, vms); err != nil {
				return err
			}
		} else {
			if err := vmsTmplWideNoHeader.Execute(w, vms); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(vms, "", "  ")
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

func NoVMs(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table", "wide":
		_, err := fmt.Fprintln(w, "No vms found. Use the `replicated vm create` command to create a new vm.")
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
