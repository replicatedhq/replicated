package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
)

var channelLicenseCountsTmplSrc = `
LICENSE_TYPE	ACTIVE	AIRGAP	INACTIVE	TOTAL
{{ range $licenseType, $counts := . -}}
{{ $licenseType }}	{{ $counts.Active }}	{{ $counts.Airgap }}	{{ $counts.Inactive }}	{{ $counts.Total }}
{{ end }}`

var channelLicenseCountsTmpl = template.Must(template.New("ChannelLicenseCounts").Parse(channelLicenseCountsTmplSrc))

type licenseTypeCounts struct {
	Active, Airgap, Inactive, Total int64
}

func LicenseCounts(format string, w *tabwriter.Writer, counts *channels.LicenseCounts) error {
	switch format {
	case "json":
		out, err := json.MarshalIndent(counts, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
		return w.Flush()
	case "table":
		countsByLicenseType := make(map[string]*licenseTypeCounts)

		var getOrSetLicenseCounts = func(licenseType string) *licenseTypeCounts {
			licenseCounts, ok := countsByLicenseType[licenseType]
			if !ok {
				licenseCounts = &licenseTypeCounts{}
				countsByLicenseType[licenseType] = licenseCounts
			}
			return licenseCounts
		}

		for licenseType, count := range counts.Active {
			getOrSetLicenseCounts(licenseType).Active = count
		}
		for licenseType, count := range counts.Airgap {
			getOrSetLicenseCounts(licenseType).Airgap = count
		}
		for licenseType, count := range counts.Inactive {
			getOrSetLicenseCounts(licenseType).Inactive = count
		}
		for licenseType, count := range counts.Total {
			getOrSetLicenseCounts(licenseType).Total = count
		}

		if len(countsByLicenseType) == 0 {
			if _, err := fmt.Fprintln(w, "No active licenses in channel"); err != nil {
				return err
			}
			return w.Flush()
		}

		if err := channelLicenseCountsTmpl.Execute(w, countsByLicenseType); err != nil {
			return err
		}

		return w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
