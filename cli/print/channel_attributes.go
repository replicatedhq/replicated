package print

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var channelAttrsTmplSrc = `ID:	{{ .Chan.ID }}
NAME:	{{ .Chan.Name }}
DESCRIPTION:	{{ .Chan.Description }}
RELEASE:	{{ if ge .Chan.ReleaseSequence 1 }}{{ .Chan.ReleaseSequence }}{{ else }}	{{ end }}
VERSION:	{{ .Chan.ReleaseLabel }}
{{ if not .Chan.IsHelmOnly -}}
{{ with .Existing -}}
EXISTING:

{{ . }}
{{ end }}
{{ with .Embedded -}}
EMBEDDED:

{{ . }}
{{ end }}
{{ with .Airgap -}}
AIRGAP:

{{ . }}
{{ end -}}
{{ end -}}
`

var channelAttrsTmpl = template.Must(template.New("ChannelAttributes").Parse(channelAttrsTmplSrc))

func ChannelAttrs(outputFormat string,
	w *tabwriter.Writer,
	appType string,
	appSlug string,
	appChan *types.Channel,
) error {
	if outputFormat == "text" {
		attrs := struct {
			Existing string
			Embedded string
			Airgap   string
			Chan     *types.Channel
		}{
			Chan: appChan,
		}
		if appType == "kots" {
			attrs.Existing = existingInstallCommand(appSlug, appChan.Slug)
			attrs.Embedded = embeddedInstallCommand(appSlug, appChan.Slug)
			attrs.Airgap = embeddedAirgapInstallCommand(appSlug, appChan.Slug)
		}
		if err := channelAttrsTmpl.Execute(w, attrs); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(appChan, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

const embeddedInstallBaseURL = "https://k8s.kurl.sh"

var embeddedInstallOverrideURL = os.Getenv("EMBEDDED_INSTALL_BASE_URL")

func embeddedInstallCommand(appSlug, chanSlug string) string {

	kurlBaseURL := embeddedInstallBaseURL
	if embeddedInstallOverrideURL != "" {
		kurlBaseURL = embeddedInstallOverrideURL
	}

	kurlURL := fmt.Sprintf("%s/%s-%s", kurlBaseURL, appSlug, chanSlug)
	if chanSlug == "stable" {
		kurlURL = fmt.Sprintf("%s/%s", kurlBaseURL, appSlug)
	}
	return fmt.Sprintf(`    curl -fsSL %s | sudo bash`, kurlURL)

}

func embeddedAirgapInstallCommand(appSlug, chanSlug string) string {

	kurlBaseURL := embeddedInstallBaseURL
	if embeddedInstallOverrideURL != "" {
		kurlBaseURL = embeddedInstallOverrideURL
	}

	slug := fmt.Sprintf("%s-%s", appSlug, chanSlug)
	if chanSlug == "stable" {
		slug = appSlug
	}
	kurlURL := fmt.Sprintf("%s/bundle/%s.tar.gz", kurlBaseURL, slug)

	return fmt.Sprintf(`    curl -fSL -o %s.tar.gz %s
    # ... scp or sneakernet %s.tar.gz to airgapped machine, then
    tar xvf %s.tar.gz
    sudo bash ./install.sh airgap`, slug, kurlURL, slug, slug)

}

func existingInstallCommand(appSlug, chanSlug string) string {

	slug := appSlug
	if chanSlug != "stable" {
		slug = fmt.Sprintf("%s/%s", appSlug, chanSlug)
	}

	return fmt.Sprintf(`    curl -fsSL https://kots.io/install | bash
    kubectl kots install %s`, slug)
}
