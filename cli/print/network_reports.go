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
var (
	networkReportTmplTableHeaderSrc = `TIMESTAMP	SRC IP	DST IP	SRC PORT	DST PORT	PROTOCOL	COMMAND	PID	DNS QUERY	SERVICE`
	networkReportTmplTableRowSrc    = `{{ range . -}}
{{ padding (printf "%s" (.Timestamp | localeTime)) 20 }}	{{ if .EventData.SrcIP }}{{ padding .EventData.SrcIP 15 }}{{ else }}{{ padding "-" 15 }}{{ end }}	{{ if .EventData.DstIP }}{{ padding .EventData.DstIP 15 }}{{ else }}{{ padding "-" 15 }}{{ end }}	{{ if .EventData.SrcPort }}{{ padding (printf "%d" .EventData.SrcPort) 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ if .EventData.DstPort }}{{ padding (printf "%d" .EventData.DstPort) 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ if .EventData.Protocol }}{{ padding .EventData.Protocol 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ if .EventData.Command }}{{ padding .EventData.Command 15 }}{{ else }}{{ padding "-" 15 }}{{ end }}	{{ if .EventData.PID }}{{ padding (printf "%d" .EventData.PID) 8 }}{{ else }}{{ padding "-" 8 }}{{ end }}	{{ if .EventData.DNSQueryName }}{{ padding .EventData.DNSQueryName 20 }}{{ else }}{{ padding "-" 20 }}{{ end }}	{{ if .EventData.LikelyService }}{{ padding .EventData.LikelyService 15 }}{{ else }}{{ padding "-" 15 }}{{ end }}
{{ end }}`
)

var (
	networkReportTmplTableSrc      = fmt.Sprintln(networkReportTmplTableHeaderSrc) + networkReportTmplTableRowSrc
	networkReportTmplTable         = template.Must(template.New("networkReport").Funcs(funcs).Parse(networkReportTmplTableSrc))
	networkReportTmplTableNoHeader = template.Must(template.New("networkReport").Funcs(funcs).Parse(networkReportTmplTableRowSrc))
)

// NetworkEventsWithData represents network events with parsed event data
type NetworkEventsWithData struct {
	Timestamp time.Time
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
func NetworkEvents(outputFormat string, w *tabwriter.Writer, events []*types.NetworkEventData, includeHeader bool) error {
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

// parseNetworkEventsData converts NetworkEventData to template format
func parseNetworkEventsData(events []*types.NetworkEventData) ([]*NetworkEventsWithData, error) {
	var eventsWithData []*NetworkEventsWithData

	for _, event := range events {
		// Extract timestamp
		var timestamp time.Time
		if parsedTime, err := time.Parse(time.RFC3339Nano, event.Timestamp); err == nil {
			timestamp = parsedTime
		}

		eventsWithData = append(eventsWithData, &NetworkEventsWithData{
			Timestamp: timestamp,
			EventData: event,
		})
	}

	return eventsWithData, nil
}
