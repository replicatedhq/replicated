package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

// Table formatting
var networksTmplTableHeaderSrc = `ID	NAME	STATUS	CREATED	EXPIRES	POLICY`
var networksTmplTableRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" (localeTime .CreatedAt)) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" (localeTime .ExpiresAt)) 30 }}{{end}}	{{if eq .Policy ""}}{{ padding "open" 30 }}{{else}}{{ padding (printf "%s" (.Policy)) 30 }}{{end}}
{{ end }}`
var networksTmplTableSrc = fmt.Sprintln(networksTmplTableHeaderSrc) + networksTmplTableRowSrc
var networksTmplTable = template.Must(template.New("networks").Funcs(funcs).Parse(networksTmplTableSrc))
var networksTmplTableNoHeader = template.Must(template.New("networks").Funcs(funcs).Parse(networksTmplTableRowSrc))

// Wide table formatting
var networksTmplWideHeaderSrc = `ID	NAME	STATUS	CREATED	EXPIRES	POLICY`
var networksTmplWideRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" (localeTime .CreatedAt)) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" (localeTime .ExpiresAt)) 30 }}{{end}}	{{if eq .Policy ""}}{{ padding "open" 30 }}{{else}}{{ padding (printf "%s" (.Policy)) 30 }}{{end}}
{{ end }}`
var networksTmplWideSrc = fmt.Sprintln(networksTmplWideHeaderSrc) + networksTmplWideRowSrc
var networksTmplWide = template.Must(template.New("networks").Funcs(funcs).Parse(networksTmplWideSrc))
var networksTmplWideNoHeader = template.Must(template.New("networks").Funcs(funcs).Parse(networksTmplWideRowSrc))

func Networks(outputFormat string, w *tabwriter.Writer, networks []*types.Network, header bool) error {
	switch outputFormat {
	case "table":
		if header {
			if err := networksTmplTable.Execute(w, networks); err != nil {
				return err
			}
		} else {
			if err := networksTmplTableNoHeader.Execute(w, networks); err != nil {
				return err
			}
		}
	case "wide":
		if header {
			if err := networksTmplWide.Execute(w, networks); err != nil {
				return err
			}
		} else {
			if err := networksTmplWideNoHeader.Execute(w, networks); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(networks, "", "  ")
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

func Network(outputFormat string, w *tabwriter.Writer, network *types.Network) error {
	switch outputFormat {
	case "table":
		if err := networksTmplTable.Execute(w, []*types.Network{network}); err != nil {
			return err
		}
	case "wide":
		if err := networksTmplWide.Execute(w, []*types.Network{network}); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent([]*types.Network{network}, "", "  ")
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

func NoNetworks(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table", "wide":
		_, err := fmt.Fprintln(w, "No networks found. Use the `replicated network create` command to create a new network.")
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
