package print

import (
	"fmt"
	"text/tabwriter"
	"text/template"

	channels "github.com/replicatedhq/replicated/gen/go/channels"
)

var channelAdoptionTmplSrc = `
LICENSE_TYPE	CURRENT	PREVIOUS	OTHER
	ACTIVE/ALL	ACTIVE/ALL	ACTIVE/ALL
{{ range $licenseType, $counts := . -}}
{{ $licenseType }}	{{ $counts.Current.Active }}/{{ $counts.Current.All }}	{{ $counts.Previous.Active }}/{{ $counts.Previous.All }}	{{ $counts.Other.Active }}/{{ $counts.Other.All }}
{{ end }}
`
var channelAdoptionTmpl = template.Must(template.New("ChannelAdoption").Parse(channelAdoptionTmplSrc))

type allActiveCounts struct {
	All    int64
	Active int64
}

type licenseAdoption struct {
	Current  allActiveCounts
	Previous allActiveCounts
	Other    allActiveCounts
}

func ChannelAdoption(w *tabwriter.Writer, adoption *channels.ChannelAdoption) error {
	countsByLicense := make(map[string]licenseAdoption)

	var getOrSetLicenseAdoption = func(licenseType string) *licenseAdoption {
		licenseAdoption, ok := countsByLicense[licenseType]
		if !ok {
			countsByLicense[licenseType] = licenseAdoption
		}
		return &licenseAdoption
	}

	// current
	for licenseType, count := range adoption.CurrentVersionCountActive {
		getOrSetLicenseAdoption(licenseType).Current.Active = count
	}
	for licenseType, count := range adoption.CurrentVersionCountAll {
		getOrSetLicenseAdoption(licenseType).Current.All = count
	}

	// previous
	for licenseType, count := range adoption.PreviousVersionCountActive {
		getOrSetLicenseAdoption(licenseType).Previous.Active = count
	}
	for licenseType, count := range adoption.PreviousVersionCountAll {
		getOrSetLicenseAdoption(licenseType).Previous.All = count
	}

	// other
	for licenseType, count := range adoption.OtherVersionCountActive {
		getOrSetLicenseAdoption(licenseType).Other.Active = count
	}
	for licenseType, count := range adoption.OtherVersionCountAll {
		getOrSetLicenseAdoption(licenseType).Other.All = count
	}

	if len(countsByLicense) == 0 {
		if _, err := fmt.Fprintln(w, "No licenses in channel"); err != nil {
			return err
		}
		return w.Flush()
	}

	if err := channelAdoptionTmpl.Execute(w, countsByLicense); err != nil {
		return err
	}

	return w.Flush()
}
