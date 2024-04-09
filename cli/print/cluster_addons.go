package print

import (
	"encoding/json"
	"fmt"
	"log"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var addonsTmplHeaderSrc = `ID	TYPE 	STATUS	DATA`
var addonsTmplRowSrc = `{{ range . -}}
{{ .ID }}	{{ Type . }}	{{ printf "%-12s" .Status }}	{{ Data . }}
{{ end }}`
var addonsTmplSrc = fmt.Sprintln(addonsTmplHeaderSrc) + addonsTmplRowSrc
var addonsTmpl = template.Must(template.New("addons").Funcs(addonsFuncs).Parse(addonsTmplSrc))
var addonsTmplNoHeader = template.Must(template.New("addons").Funcs(addonsFuncs).Parse(addonsTmplRowSrc))

var addonsFuncs = template.FuncMap{
	"Type": func(addon *types.ClusterAddon) string {
		return addon.TypeName()
	},
	"Data": addonData,
}

func init() {
	for k, v := range funcs {
		addonsFuncs[k] = v
	}
}

func Addons(outputFormat string, w *tabwriter.Writer, addons []*types.ClusterAddon, header bool) error {
	switch outputFormat {
	case "table", "wide":
		if header {
			if err := addonsTmpl.Execute(w, addons); err != nil {
				return err
			}
		} else {
			if err := addonsTmplNoHeader.Execute(w, addons); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(addons, "", "  ")
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

func Addon(outputFormat string, w *tabwriter.Writer, addon *types.ClusterAddon) error {
	switch outputFormat {
	case "table", "wide":
		if err := addonsTmpl.Execute(w, []*types.ClusterAddon{addon}); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(addon, "", "  ")
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

func addonData(addon *types.ClusterAddon) string {
	switch {
	case addon.ObjectStore != nil:
		return addonObjectStoreData(*addon.ObjectStore)
	case addon.Postgres != nil:
		return addonPostgresData(*addon.Postgres)
	default:
		return ""
	}
}

func addonObjectStoreData(data types.ClusterAddonObjectStore) string {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal object store data: %v", err)
	}
	return string(b)
}

func addonPostgresData(data types.ClusterAddonPostgres) string {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal postgres data: %v", err)
	}
	return string(b)
}
