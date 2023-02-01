package print

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/util"
	"text/template"
	"time"
)

var customerInstancesTmplSrc = `ID	STATUS	VERSION	VERSIONS-BEHIND	VERSION-AGE	VERSION-AGE-RELATIVE	LAST-CHECKIN	TTI-LICENSE	TTI-INSTANCE	FIRST-SEEN	FIRST-READY	LICENSE-CREATED	LATEST-CHANNEL-RELEASE	FULL-ID
{{ range . -}}
{{ .ShortID }}	{{ .Instance.AppStatus }}	{{.Instance.AppVersion}}	{{.Instance.VersionsBehind}}	{{.AgeAbsolute}}	{{.AgeRelative}}	{{.Instance.LastCheckIn | Date}}	{{.TTILicense }}	{{.TTIInstance}}	{{.Instance.FirstCheckinAt | Date}}	{{.Instance.FirstReadyAt | Date}}	{{.Insights.LicenseCreatedAt | Date }}	{{.Instance.CurrentVersionReleasedAt | Date}}	{{.Instance.ID}}	
{{ end }}`

var customerInstancesTmplLiteSrc = `ID	STATUS	VERSION	VERSIONS-BEHIND	VERSION-AGE	VERSION-AGE-RELATIVE	LAST-CHECKIN	TTI-LICENSE	TTI-INSTANCE	FIRST-READY	
{{ range . -}}
{{ .ShortID }}	{{ .Instance.AppStatus }}	{{.Instance.AppVersion}}	{{.Instance.VersionsBehind}}	{{.AgeAbsolute}}	{{.AgeRelative}}	{{.Instance.LastCheckIn | Date}}	{{.TTILicense }}	{{.TTIInstance}}	{{.Instance.FirstReadyAt | Date}}	
{{ end }}`

func NiceDate(maybeTime interface{}) (string, error) {
	var t time.Time
	switch value := maybeTime.(type) {
	case time.Time:
		t = value
	case *util.Time:
		if value == nil {
			return "--", nil
		}
		t = value.Time
	default:
		return "", errors.Errorf("invalid type %T", maybeTime)
	}

	if time.Now().Sub(t).Hours() < 1*24 {
		return t.Format("2006-01-02 15:04"), nil
	}

	return t.Format("2006-01-02 00:00"), nil
}

var CustomerInstancesTmplWide = template.Must(
	template.New("customer instances").
		Funcs(template.FuncMap{"Date": NiceDate}).
		Parse(customerInstancesTmplSrc),
)
var CustomerInstancesTmplLite = template.Must(
	template.New("customer instances lite").
		Funcs(template.FuncMap{"Date": NiceDate}).
		Parse(customerInstancesTmplLiteSrc),
)
