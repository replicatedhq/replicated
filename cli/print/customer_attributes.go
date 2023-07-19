package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var customerAttrsTmplSrc = `ID: {{ .Cust.ID }}
NAME: {{ .Cust.Name }}
EMAIL: {{ .Cust.Email }}
LICENSE_ID: {{ .Cust.InstallationID }}
EXPIRES: {{if .Cust.Expires }}{{ .Cust.Expires }}{{else}}Never{{end}}
{{if .Install}}
{{with .Login}}
LOGIN:

{{ . }}
{{end}}

INSTALL PREFLIGHT:

    curl https://krew.sh/preflight | bash
{{with .Preflight}}

PREFLIGHT:

{{ . }}
{{end}}
{{with .Install}}
INSTALL:

{{ . }}
{{end}}
{{end}}
`

var customerAttrsTmpl = template.Must(template.New("CustomerAttributes").Parse(customerAttrsTmplSrc))

func CustomerAttrs(outputFormat string,
	w *tabwriter.Writer,
	appType string,
	appSlug string,
	ch *types.KotsChannel,
	registryHostname string,
	cust *types.Customer,
) error {
	if outputFormat == "text" {
		attrs := struct {
			Login     string
			Preflight string
			Install   string
			Cust      *types.Customer
		}{
			Cust: cust,
		}
		if appType == "kots" {
			attrs.Login = loginCommand(registryHostname, cust.Email, cust.InstallationID)
			attrs.Preflight = preflightCommand(registryHostname, appSlug, ch)
			attrs.Install = installCommand(registryHostname, appSlug, ch)
		}
		if err := customerAttrsTmpl.Execute(w, attrs); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(cust, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func loginCommand(host, email, password string) string {
	if email == "" {
		return "Customer email required to login"
	}
	return fmt.Sprintf(`    helm registry login %s --username %s --password %s`, host, email, password)
}

func preflightCommand(host, appSlug string, ch *types.KotsChannel) string {
	preflightReleases := []types.ChartRelease{}
	for _, chart := range ch.ChartReleases {
		if chart.HasPreflightSpec {
			preflightReleases = append(preflightReleases, chart)
		}
	}

	if len(preflightReleases) == 0 {
		return "No preflight checks found"
	}

	var cmd string
	for i, chart := range preflightReleases {
		var delim string
		if i < len(preflightReleases)-1 {
			delim = " &&\n"
		}
		if ch.ChannelSlug == "stable" {
			cmd += fmt.Sprintf(`    helm template %s oci://%s/%s/%s | kubectl preflight -%s`,
				chart.Name, host, appSlug, chart.Name, delim)
		} else {
			cmd += fmt.Sprintf(`    helm template %s oci://%s/%s/%s/%s | kubectl preflight -%s`,
				chart.Name, host, appSlug, ch.ChannelSlug, chart.Name, delim)
		}
	}

	return cmd
}

func installCommand(host, appSlug string, ch *types.KotsChannel) string {
	var cmd string
	for i, chart := range ch.ChartReleases {
		var delim string
		if i < len(ch.ChartReleases)-1 {
			delim = " &&\n"
		}
		if ch.ChannelSlug == "stable" {
			cmd += fmt.Sprintf(`    helm install %s oci://%s/%s/%s%s`, chart.Name, host, appSlug, chart.Name, delim)
		} else {
			cmd += fmt.Sprintf(`    helm install %s oci://%s/%s/%s/%s%s`, chart.Name, host, appSlug, ch.ChannelSlug, chart.Name, delim)
		}
	}

	return cmd
}
