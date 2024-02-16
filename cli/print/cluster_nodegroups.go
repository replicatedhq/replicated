package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var nodeGroupsTmplSrc = `ID	NAME	DEFAULT	INSTANCE TYPE	NODES	DISK
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .IsDefault }}	{{ .InstanceType }}	{{.NodeCount}}	{{.DiskGiB}}
{{ end }}`

var nodeGroupsTmpl = template.Must(template.New("nodegroups").Parse(nodeGroupsTmplSrc))

func NodeGroups(outputFormat string, w *tabwriter.Writer, addOns []*types.NodeGroup) error {
	switch outputFormat {
	case "table", "wide":
		if err := nodeGroupsTmpl.Execute(w, addOns); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(addOns, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return w.Flush()
}
