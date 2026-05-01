package print

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var notificationTemplateFuncs = template.FuncMap{
	"notificationName":        notificationName,
	"subscriptionType":        subscriptionType,
	"subscriptionDestination": subscriptionDestination,
	"subscriptionVerified":    subscriptionVerified,
	"webhookStatusCode":       webhookStatusCode,
}

var notificationSubscriptionsTmpl = template.Must(template.New("notification-subscriptions").Funcs(notificationTemplateFuncs).Parse(`ID	NAME	TYPE	DESTINATION	ENABLED	VERIFIED
{{ range . -}}
{{ .ID }}	{{ notificationName . }}	{{ subscriptionType . }}	{{ subscriptionDestination . }}	{{ .IsEnabled }}	{{ subscriptionVerified . }}
{{ end }}`))

var notificationEventsTmpl = template.Must(template.New("notification-events").Funcs(funcs).Parse(`ID	EVENT TYPE	STATUS	SUBSCRIPTION	CREATED	ATTEMPTS
{{ range . -}}
{{ .ID }}	{{ .EventType }}	{{ .Status }}	{{ .SubscriptionID }}	{{ localeTime .CreatedAt }}	{{ len .Attempts }}
{{ end }}`))

var notificationEventTypesTmpl = template.Must(template.New("notification-event-types").Parse(`KEY	DISPLAY NAME	CATEGORY	DESCRIPTION
{{ range . -}}
{{ .Key }}	{{ .DisplayName }}	{{ .Category }}	{{ .Description }}
{{ end }}`))

var notificationEmailActionTmpl = template.Must(template.New("notification-email-action").Parse(`SUCCESS	MESSAGE
{{ .Success }}	{{ .Message }}
`))

var notificationRetryTmpl = template.Must(template.New("notification-retry").Parse(`SUCCESS	MESSAGE
{{ .Success }}	{{ .Message }}
`))

var notificationWebhookTestTmpl = template.Must(template.New("notification-webhook-test").Funcs(notificationTemplateFuncs).Parse(`SUCCESS	STATUS CODE	DURATION (MS)	ERROR
{{ .Success }}	{{ webhookStatusCode .StatusCode }}	{{ .DurationMs }}	{{ .Error }}
`))

func Notifications(outputFormat string, w *tabwriter.Writer, subscriptions []types.NotificationSubscription) error {
	switch outputFormat {
	case "table":
		if err := notificationSubscriptionsTmpl.Execute(w, subscriptions); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(subscriptions, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func Notification(outputFormat string, w *tabwriter.Writer, subscription *types.NotificationSubscription) error {
	return Notifications(outputFormat, w, []types.NotificationSubscription{*subscription})
}

func NotificationEvents(outputFormat string, w *tabwriter.Writer, events []types.NotificationEvent) error {
	switch outputFormat {
	case "table":
		if err := notificationEventsTmpl.Execute(w, events); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NotificationEventTypes(outputFormat string, w *tabwriter.Writer, response *types.NotificationEventTypesResponse) error {
	switch outputFormat {
	case "table":
		if err := notificationEventTypesTmpl.Execute(w, response.EventTypes); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NotificationEmailAction(outputFormat string, w *tabwriter.Writer, response *types.NotificationEmailResponse) error {
	switch outputFormat {
	case "table":
		if err := notificationEmailActionTmpl.Execute(w, response); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NotificationRetry(outputFormat string, w *tabwriter.Writer, response *types.NotificationRetryResponse) error {
	switch outputFormat {
	case "table":
		if err := notificationRetryTmpl.Execute(w, response); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NotificationWebhookTest(outputFormat string, w *tabwriter.Writer, response *types.NotificationWebhookTestResponse) error {
	switch outputFormat {
	case "table":
		if err := notificationWebhookTestTmpl.Execute(w, response); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NoNotifications(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table":
		_, err := fmt.Fprintln(w, "No notification subscriptions found.")
		if err != nil {
			return err
		}
	case "json":
		if _, err := fmt.Fprintln(w, "[]"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NoNotificationEvents(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table":
		_, err := fmt.Fprintln(w, "No notification events found.")
		if err != nil {
			return err
		}
	case "json":
		if _, err := fmt.Fprintln(w, "[]"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func notificationName(sub types.NotificationSubscription) string {
	if strings.TrimSpace(sub.Name) != "" {
		return sub.Name
	}
	if len(sub.EventConfigs) == 1 {
		return sub.EventConfigs[0].EventType
	}
	if len(sub.EventConfigs) > 1 {
		return fmt.Sprintf("%d event types", len(sub.EventConfigs))
	}
	return "-"
}

func subscriptionType(sub types.NotificationSubscription) string {
	if strings.TrimSpace(sub.WebhookURL) != "" {
		return "webhook"
	}
	if strings.TrimSpace(sub.EmailAddress) != "" {
		return "email"
	}
	return "unknown"
}

func subscriptionDestination(sub types.NotificationSubscription) string {
	if strings.TrimSpace(sub.WebhookURL) != "" {
		return sub.WebhookURL
	}
	if strings.TrimSpace(sub.EmailAddress) != "" {
		return sub.EmailAddress
	}
	return "-"
}

func subscriptionVerified(sub types.NotificationSubscription) string {
	if strings.TrimSpace(sub.EmailAddress) == "" {
		return "-"
	}
	if sub.EmailVerified {
		return "true"
	}
	if sub.RequiresVerification {
		return "pending"
	}
	return "false"
}

func webhookStatusCode(code *int) string {
	if code == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *code)
}
