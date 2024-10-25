package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

var vmFuncs = template.FuncMap{
	"CreditsToDollarsDisplay": CreditsToDollarsDisplay,
}

// Table formatting
var vmsTmplTableHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	STATUS	CREATED	EXPIRES	COST`
var vmsTmplTableRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .Distribution 12 }}	{{ padding .Version 10 }}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" (localeTime .CreatedAt)) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" (localeTime .ExpiresAt)) 30 }}{{end}}	{{ padding (CreditsToDollarsDisplay .EstimatedCost) 11 }}
{{ end }}`
var vmsTmplTableSrc = fmt.Sprintln(vmsTmplTableHeaderSrc) + vmsTmplTableRowSrc
var vmsTmplTable = template.Must(template.New("vms").Funcs(vmFuncs).Funcs(funcs).Parse(vmsTmplTableSrc))
var vmsTmplTableNoHeader = template.Must(template.New("vms").Funcs(vmFuncs).Funcs(funcs).Parse(vmsTmplTableRowSrc))

// Wide table formatting
var vmsTmplWideHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	STATUS	CREATED	EXPIRES	COST	TAGS`
var vmsTmplWideRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .Distribution 12 }}	{{ padding .Version 10 }}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" (localeTime .CreatedAt)) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" (localeTime .ExpiresAt)) 30 }}{{end}}	{{ padding (CreditsToDollarsDisplay .EstimatedCost) 11 }}	{{ range $index, $tag := .Tags }}{{if $index}}, {{end}}{{ $tag.Key }}={{ $tag.Value }}{{ end }}
{{ end }}`
var vmsTmplWideSrc = fmt.Sprintln(vmsTmplWideHeaderSrc) + vmsTmplWideRowSrc
var vmsTmplWide = template.Must(template.New("vms").Funcs(vmFuncs).Funcs(funcs).Parse(vmsTmplWideSrc))
var vmsTmplWideNoHeader = template.Must(template.New("vms").Funcs(vmFuncs).Funcs(funcs).Parse(vmsTmplWideRowSrc))

// VM versions
var vmVersionsTmplSrc = `Supported VM distributions and versions are:
{{ range $d := . -}}
DISTRIBUTION: {{ $d.Name }}
• VERSIONS: {{ range $i, $v := $d.Versions -}}{{if $i}}, {{end}}{{ $v }}{{ end }}
• INSTANCE TYPES: {{ range $i, $it := $d.InstanceTypes -}}{{if $i}}, {{end}}{{ $it }}{{ end }}{{if $d.Status}}
• ENABLED: {{ $d.Status.Enabled }}
• STATUS: {{ $d.Status.Status }}
• DETAILS: {{ $d.Status.StatusMessage }}{{end}}

{{ end }}`
var vmVersionsTmpl = template.Must(template.New("vmVersions").Funcs(funcs).Parse(vmVersionsTmplSrc))

func VMs(outputFormat string, w *tabwriter.Writer, vms []*types.VM, header bool) error {
	for _, vm := range vms {
		updateEstimatedVMCost(vm)
	}
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

func NoVMVersions(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table":
		_, err := fmt.Fprintln(w, "No VM versions found.")
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

func VMVersions(outputFormat string, w *tabwriter.Writer, versions []*types.VMVersion) error {
	switch outputFormat {
	case "table":
		if err := vmVersionsTmpl.Execute(w, versions); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(versions, "", "  ")
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

func VM(outputFormat string, w *tabwriter.Writer, vm *types.VM) error {
	updateEstimatedVMCost(vm)
	switch outputFormat {
	case "table":
		if err := vmsTmplTable.Execute(w, []*types.VM{vm}); err != nil {
			return err
		}
	case "wide":
		if err := vmsTmplWide.Execute(w, []*types.VM{vm}); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(vm, "", "  ")
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

func updateEstimatedVMCost(vm *types.VM) {
	if vm.EstimatedCost != 0 {
		return
	}
	if vm.TotalCredits > 0 {
		vm.EstimatedCost = vm.TotalCredits
	} else {
		expireDuration, _ := time.ParseDuration(vm.TTL)
		minutesRunning := int64(expireDuration.Minutes())
		totalCredits := int64(minutesRunning) * vm.CreditsPerHour / 60.0
		vm.EstimatedCost = vm.FlatFee + totalCredits
	}
}
