package print

import (
	"bytes"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"log"
	"text/template"
	"time"
)

const appDetail = `App Status: {{.Instance.AppStatus}}
App Version: {{.Instance.AppVersion}}
Versions Behind: {{.Instance.VersionsBehind}}
Version Age: {{.AgeAbsolute}} ({{.AgeRelative}} behind latest)
Last Check-in: {{.Instance.LastCheckIn | Date}}
`

var appDetailTemplate = template.Must(
	template.New("customer instance inspect app detail").
		Funcs(template.FuncMap{"Date": NiceDate}).
		Parse(appDetail),
)

const clusterInfo = `Cluster Type: {{.Instance.K8sDistribution}}
K8s Version: {{.Instance.K8sVersion}}
KOTS Version: {{.Instance.KOTSVersion}}
First Seen: {{.Instance.FirstCheckinAt | Date}}
`

var clusterInfoTemplate = template.Must(
	template.New("customer instance inspect cluster info").
		Funcs(template.FuncMap{"Date": NiceDate}).
		Parse(clusterInfo),
)

const insights = `Time to Install 

Instance: {{.TTIInstance}}
License: {{.TTILicense}}
`

var insightsTemplate = template.Must(
	template.New("customer instance inspect insights").
		Funcs(template.FuncMap{"Date": NiceDate}).
		Parse(insights),
)

func InstanceInspect(instance *types.CustomerInstance, uptime *types.CustomerInstanceUptime, uptimeInterval time.Duration) error {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	appDetails := &bytes.Buffer{}
	if err := appDetailTemplate.Execute(appDetails, instance); err != nil {
		return errors.Wrap(err, "template app info")
	}

	clusterInfo := &bytes.Buffer{}
	if err := clusterInfoTemplate.Execute(clusterInfo, instance); err != nil {
		return errors.Wrap(err, "template cluster info")
	}

	insights := &bytes.Buffer{}
	if err := insightsTemplate.Execute(insights, instance); err != nil {
		return errors.Wrap(err, "template insights")
	}

	xoffset := 39
	appDetailsTile := widgets.NewParagraph()
	appDetailsTile.Text = appDetails.String()
	appDetailsTile.SetRect(0, 0, xoffset, 7)

	clusterInfoTile := widgets.NewParagraph()
	clusterInfoTile.Text = clusterInfo.String()
	clusterInfoTile.SetRect(xoffset, 0, 2*xoffset, 7)

	insightsTile := widgets.NewParagraph()
	insightsTile.Text = insights.String()
	insightsTile.SetRect(2*xoffset, 0, 3*xoffset, 7)

	uptimeChart := widgets.NewStackedBarChart()
	uptimeChart.Title = "Instance Uptime (last 3 days)"
	uptimeChart.MaxVal = 110
	uptimeChart.NumFormatter = func(f float64) string {
		if f == 0 {
			return ""
		}

		hours := f / 100 * uptimeInterval.Hours()
		duration, _ := time.ParseDuration(fmt.Sprintf("%fh", hours))

		return instance.FormatDuration(&duration)
	}
	uptimeChart.BarColors = []ui.Color{
		ui.ColorBlack,
		ui.ColorRed,
		ui.ColorYellow,
		ui.ColorGreen,
	}
	uptimeChart.LabelStyles = []ui.Style{
		{Fg: ui.ColorWhite},
	}
	uptimeChart.NumStyles = []ui.Style{
		{Fg: ui.ColorWhite, Modifier: ui.ModifierBold},
		{Fg: ui.ColorWhite},
	}

	for _, interval := range uptime.Histogram {
		uptimeChart.Labels = append(uptimeChart.Labels, fmt.Sprintf("-%.0fh", time.Now().Sub(interval.StartTime.Time).Hours()))
		up := interval.Statuses.Ready + interval.Statuses.Updating
		down := interval.Statuses.Unavailable + interval.Statuses.Missing
		degraded := interval.Statuses.Degraded
		unknown := interval.Statuses.Unknown
		uptimeChart.Data = append(uptimeChart.Data, []float64{unknown, down, degraded, up})

	}
	uptimeChart.SetRect(0, 8, (3*xoffset)-1, 20)
	uptimeChart.BarWidth = 5

	ui.Render(uptimeChart)

	ui.Render(appDetailsTile)
	ui.Render(clusterInfoTile)
	ui.Render(insightsTile)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
	return nil
}
