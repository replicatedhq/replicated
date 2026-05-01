package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNotificationCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Manage event notifications",
		Long:  "List, create, update, test, and manage event notification subscriptions and delivery events.",
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationSubscriptionCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscription",
		Short: "Manage notification subscriptions",
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationEventCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage notification delivery events",
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationEventTypeCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-type",
		Short: "List available notification event types",
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationEmailCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email",
		Short: "Manage notification email verification",
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationWebhookCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Test notification webhooks",
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationSubscriptionList(parent *cobra.Command) *cobra.Command {
	var outputFormat, search, subscriptionType string

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List notification subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := r.kotsAPI.ListNotificationSubscriptions(search, subscriptionType)
			if err != nil {
				return errors.Wrap(err, "list notification subscriptions")
			}
			if len(resp.Subscriptions) == 0 {
				return print.NoNotifications(outputFormat, r.w)
			}
			return print.Notifications(outputFormat, r.w, resp.Subscriptions)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&search, "search", "", "Text search filter")
	cmd.Flags().StringVar(&subscriptionType, "type", "", "Filter by subscription type: personal|team")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	return cmd
}

func (r *runners) InitNotificationSubscriptionGet(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get ID",
		Short: "Get a notification subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscription, err := r.kotsAPI.GetNotificationSubscription(args[0])
			if err != nil {
				return errors.Wrap(err, "get notification subscription")
			}
			return print.Notification(outputFormat, r.w, subscription)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	return cmd
}

func (r *runners) InitNotificationSubscriptionCreate(parent *cobra.Command) *cobra.Command {
	var outputFormat, file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a notification subscription from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := readNotificationCreateRequest(file)
			if err != nil {
				return err
			}
			subscription, err := r.kotsAPI.CreateNotificationSubscription(req)
			if err != nil {
				return errors.Wrap(err, "create notification subscription")
			}
			return print.Notification(outputFormat, r.w, subscription)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&file, "file", "", "Path to a JSON file containing the subscription definition")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func (r *runners) InitNotificationSubscriptionUpdate(parent *cobra.Command) *cobra.Command {
	var outputFormat, file string

	cmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update a notification subscription from a JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := readNotificationUpdateRequest(file)
			if err != nil {
				return err
			}
			subscription, err := r.kotsAPI.UpdateNotificationSubscription(args[0], req)
			if err != nil {
				return errors.Wrap(err, "update notification subscription")
			}
			return print.Notification(outputFormat, r.w, subscription)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&file, "file", "", "Path to a JSON file containing the subscription patch")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func (r *runners) InitNotificationSubscriptionRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID",
		Aliases: []string{"delete", "remove"},
		Short:   "Delete a notification subscription",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := r.kotsAPI.DeleteNotificationSubscription(args[0]); err != nil {
				return errors.Wrap(err, "delete notification subscription")
			}
			_, err := fmt.Fprintf(r.w, "Notification subscription %s deleted\n", args[0])
			if err != nil {
				return err
			}
			return r.w.Flush()
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) InitNotificationSubscriptionEvents(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "events ID",
		Short: "List delivery events for a notification subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := r.kotsAPI.ListNotificationSubscriptionEvents(args[0])
			if err != nil {
				return errors.Wrap(err, "list notification subscription events")
			}
			if len(resp.Events) == 0 {
				return print.NoNotificationEvents(outputFormat, r.w)
			}
			return print.NotificationEvents(outputFormat, r.w, resp.Events)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	return cmd
}

func (r *runners) InitNotificationEventList(parent *cobra.Command) *cobra.Command {
	var outputFormat string
	opts := kotsclient.ListNotificationEventsOpts{}

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List notification delivery events",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.StartTime != "" {
				if _, err := time.Parse(time.RFC3339, opts.StartTime); err != nil {
					return errors.Wrap(err, "parse start-time")
				}
			}
			if opts.EndTime != "" {
				if _, err := time.Parse(time.RFC3339, opts.EndTime); err != nil {
					return errors.Wrap(err, "parse end-time")
				}
			}
			resp, err := r.kotsAPI.ListNotificationEvents(opts)
			if err != nil {
				return errors.Wrap(err, "list notification events")
			}
			if len(resp.Events) == 0 {
				return print.NoNotificationEvents(outputFormat, r.w)
			}
			return print.NotificationEvents(outputFormat, r.w, resp.Events)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by category prefix: all|instance|release|customer|channel|support|support_bundle")
	cmd.Flags().StringArrayVar(&opts.EventTypes, "event-type", []string{}, "Filter by event type key (repeatable)")
	cmd.Flags().StringVar(&opts.SubscriptionID, "subscription-id", "", "Filter by subscription ID")
	cmd.Flags().StringVar(&opts.SubscriptionType, "subscription-type", "", "Filter by subscription type: personal|team")
	cmd.Flags().StringVar(&opts.StartTime, "start-time", "", "Filter events after this time (RFC3339)")
	cmd.Flags().StringVar(&opts.EndTime, "end-time", "", "Filter events before this time (RFC3339)")
	cmd.Flags().StringVar(&opts.Search, "search", "", "Search event content")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status: success|pending|failed")
	cmd.Flags().IntVar(&opts.CurrentPage, "current-page", 0, "Pagination page index")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Pagination page size")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	return cmd
}

func (r *runners) InitNotificationEventRetry(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "retry EVENT_ID",
		Short: "Retry a notification event",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := r.kotsAPI.RetryNotificationEvent(args[0])
			if err != nil {
				return errors.Wrap(err, "retry notification event")
			}
			return print.NotificationRetry(outputFormat, r.w, resp)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	return cmd
}

func (r *runners) InitNotificationEventTypeList(parent *cobra.Command) *cobra.Command {
	var outputFormat, query string
	var limit int

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List notification event types",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := r.kotsAPI.ListNotificationEventTypes(query, limit)
			if err != nil {
				return errors.Wrap(err, "list notification event types")
			}
			return print.NotificationEventTypes(outputFormat, r.w, resp)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&query, "q", "", "Search query")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results to request from the API (0 means no limit)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	return cmd
}

func (r *runners) InitNotificationEmailResendVerification(parent *cobra.Command) *cobra.Command {
	var outputFormat, email string

	cmd := &cobra.Command{
		Use:   "resend-verification",
		Short: "Resend a verification email",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(email) == "" {
				return errors.New("the --email flag is required")
			}
			resp, err := r.kotsAPI.ResendNotificationVerification(kotsclient.ResendNotificationVerificationRequest{
				EmailAddress: email,
			})
			if err != nil {
				return errors.Wrap(err, "resend verification email")
			}
			return print.NotificationEmailAction(outputFormat, r.w, resp)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&email, "email", "", "Email address to verify")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func (r *runners) InitNotificationEmailVerify(parent *cobra.Command) *cobra.Command {
	var outputFormat, email, code string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify an email address for notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(email) == "" {
				return errors.New("the --email flag is required")
			}
			if strings.TrimSpace(code) == "" {
				return errors.New("the --code flag is required")
			}
			resp, err := r.kotsAPI.VerifyNotificationEmail(kotsclient.VerifyNotificationEmailRequest{
				EmailAddress:     email,
				VerificationCode: code,
			})
			if err != nil {
				return errors.Wrap(err, "verify notification email")
			}
			return print.NotificationEmailAction(outputFormat, r.w, resp)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&email, "email", "", "Email address to verify")
	cmd.Flags().StringVar(&code, "code", "", "Verification code")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("code")
	return cmd
}

func (r *runners) InitNotificationWebhookTest(parent *cobra.Command) *cobra.Command {
	var outputFormat, file string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Send a test webhook from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := readNotificationWebhookTestRequest(file)
			if err != nil {
				return err
			}
			resp, err := r.kotsAPI.TestNotificationWebhook(req)
			if err != nil {
				return errors.Wrap(err, "test notification webhook")
			}
			return print.NotificationWebhookTest(outputFormat, r.w, resp)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&file, "file", "", "Path to a JSON file containing the webhook test payload")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func readNotificationCreateRequest(path string) (kotsclient.CreateNotificationSubscriptionRequest, error) {
	var req kotsclient.CreateNotificationSubscriptionRequest
	if err := readJSONFile(path, &req); err != nil {
		return req, err
	}
	if len(req.EventConfigs) == 0 {
		return req, errors.New("eventConfigs must contain at least one event config")
	}
	if strings.TrimSpace(req.EmailAddress) == "" && strings.TrimSpace(req.WebhookURL) == "" {
		return req, errors.New("one of emailAddress or webhookUrl must be set")
	}
	return req, nil
}

func readNotificationUpdateRequest(path string) (kotsclient.UpdateNotificationSubscriptionRequest, error) {
	var req kotsclient.UpdateNotificationSubscriptionRequest
	if err := readJSONFile(path, &req); err != nil {
		return req, err
	}
	if req.Name == nil && len(req.EventConfigs) == 0 && req.IsEnabled == nil && req.EmailAddress == "" && req.WebhookURL == "" && req.WebhookSecret == "" && req.CustomHeaders == nil {
		return req, errors.New("subscription update file must include at least one field")
	}
	return req, nil
}

func readNotificationWebhookTestRequest(path string) (kotsclient.TestNotificationWebhookRequest, error) {
	var req kotsclient.TestNotificationWebhookRequest
	if err := readJSONFile(path, &req); err != nil {
		return req, err
	}
	if strings.TrimSpace(req.WebhookURL) == "" {
		return req, errors.New("webhookUrl is required")
	}
	if strings.TrimSpace(req.EventType) == "" {
		return req, errors.New("eventType is required")
	}
	return req, nil
}

func readJSONFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "read file")
	}
	if !json.Valid(data) {
		return errors.New("file is not valid JSON")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return errors.Wrap(err, "unmarshal json")
	}
	return nil
}
