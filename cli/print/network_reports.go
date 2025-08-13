package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

// Table formatting for network reports
var networkReportTmplTableHeaderSrc = `CREATED AT	SRC IP	DST IP	SRC PORT	DST PORT	PROTOCOL	COMMAND	PID	DNS QUERY	SERVICE`
var networkReportTmplTableRowSrc = `{{ range . -}}
{{ padding (printf "%s" (.CreatedAt | localeTime)) 20 }}	{{ padding .EventData.SrcIP 15 }}	{{ padding .EventData.DstIP 15 }}	{{ if .EventData.SrcPort }}{{ padding (printf "%d" .EventData.SrcPort) 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ if .EventData.DstPort }}{{ padding (printf "%d" .EventData.DstPort) 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ padding .EventData.Protocol 8 }}	{{ padding .EventData.Command 15 }}	{{ if .EventData.PID }}{{ padding (printf "%d" .EventData.PID) 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ padding .EventData.DNSQueryName 20 }}	{{ padding .EventData.LikelyService 15 }}
{{ end }}`

var networkReportTmplTableSrc = fmt.Sprintln(networkReportTmplTableHeaderSrc) + networkReportTmplTableRowSrc
var networkReportTmplTable = template.Must(template.New("networkReport").Funcs(funcs).Parse(networkReportTmplTableSrc))
var networkReportTmplTableNoHeader = template.Must(template.New("networkReport").Funcs(funcs).Parse(networkReportTmplTableRowSrc))

// NetworkEventsWithData represents network events with parsed event data
type NetworkEventsWithData struct {
	CreatedAt time.Time
	EventData *types.NetworkEventData
}

// NetworkReport prints network report in various formats
func NetworkReport(outputFormat string, w *tabwriter.Writer, report *types.NetworkReport) error {
	switch outputFormat {
	case "table":
		if len(report.Events) == 0 {
			_, err := fmt.Fprintln(w, "No network events found.")
			return err
		}
		eventsWithData, err := parseNetworkEventsData(report.Events)
		if err != nil {
			return fmt.Errorf("failed to parse network events: %v", err)
		}
		if err := networkReportTmplTable.Execute(w, eventsWithData); err != nil {
			return err
		}
	case "json":
		reportBytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report to json: %v", err)
		}
		if _, err := fmt.Fprintln(w, string(reportBytes)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return w.Flush()
}

// NetworkEvents prints network events in table format (for watch mode)
func NetworkEvents(outputFormat string, w *tabwriter.Writer, events []*types.NetworkEvent, includeHeader bool) error {
	switch outputFormat {
	case "table":
		if len(events) == 0 {
			return nil
		}
		eventsWithData, err := parseNetworkEventsData(events)
		if err != nil {
			return fmt.Errorf("failed to parse network events: %v", err)
		}
		if includeHeader {
			if err := networkReportTmplTable.Execute(w, eventsWithData); err != nil {
				return err
			}
		} else {
			if err := networkReportTmplTableNoHeader.Execute(w, eventsWithData); err != nil {
				return err
			}
		}
	case "json":
		for _, event := range events {
			eventBytes, err := json.Marshal(event)
			if err != nil {
				continue // Skip events that can't be marshaled
			}
			if _, err := fmt.Fprintln(w, string(eventBytes)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return w.Flush()
}

// parseNetworkEventsData parses the JSON event data for template consumption
func parseNetworkEventsData(events []*types.NetworkEvent) ([]*NetworkEventsWithData, error) {
	var eventsWithData []*NetworkEventsWithData

	for _, event := range events {
		var eventData types.NetworkEventData
		if err := json.Unmarshal([]byte(event.EventData), &eventData); err != nil {
			// For events that can't be parsed, create a minimal entry
			eventsWithData = append(eventsWithData, &NetworkEventsWithData{
				CreatedAt: event.CreatedAt,
				EventData: &types.NetworkEventData{},
			})
		} else {
			eventsWithData = append(eventsWithData, &NetworkEventsWithData{
				CreatedAt: event.CreatedAt,
				EventData: &eventData,
			})
		}
	}

	return eventsWithData, nil
}
