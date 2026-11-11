package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/replicatedhq/replicated/pkg/platformclient"
)

const maxPreviewMetadataBytes = 96 * 1024

type previewMetadataPin struct {
	Name         string `json:"name"`
	SourceBranch string `json:"sourceBranch"`
}

type previewMetadataCustomer struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	Error          string   `json:"error,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	ChannelID      string   `json:"channelId,omitempty"`
	PinnedSequence *int64   `json:"pinnedSequence,omitempty"`
}

type previewMetadataResponse struct {
	Version       int                     `json:"version"`
	AppID         string                  `json:"appId"`
	Customer      previewMetadataCustomer `json:"customer"`
	HideUnmatched bool                    `json:"hideUnmatched"`
	Branches      []string                `json:"branches"`
	Pins          []previewMetadataPin    `json:"pins"`
}

type previewMetadataSnapshot struct {
	Version       int                       `json:"version"`
	CapturedAt    string                    `json:"capturedAt"`
	AppID         string                    `json:"appId"`
	HideUnmatched bool                      `json:"hideUnmatched"`
	Branches      []string                  `json:"branches"`
	Pins          []previewMetadataPin      `json:"pins"`
	Customers     []previewMetadataCustomer `json:"customers"`
}

type previewMetadataFetcher func(context.Context, string, string) (previewMetadataResponse, error)

func buildPreviewMetadataEnv(ctx context.Context, appID string, customerIDs []string, token string) (string, error) {
	client := platformclient.NewHTTPClient(platformOrigin, token)
	return collectPreviewMetadata(ctx, appID, customerIDs, func(ctx context.Context, appID, customerID string) (previewMetadataResponse, error) {
		path := fmt.Sprintf("/v3/app/%s/enterprise-portal/preview-metadata?customer_id=%s", url.PathEscape(appID), url.QueryEscape(customerID))
		var response previewMetadataResponse
		err := client.DoJSON(ctx, http.MethodGet, path, http.StatusOK, nil, &response)
		return response, err
	})
}

func collectPreviewMetadata(ctx context.Context, appID string, customerIDs []string, fetch previewMetadataFetcher) (string, error) {
	results := make([]previewMetadataResponse, len(customerIDs))
	lookupErrors := make([]error, len(customerIDs))
	limit := make(chan struct{}, 4)
	var pending sync.WaitGroup
	for i, customerID := range customerIDs {
		pending.Add(1)
		go func() {
			defer pending.Done()
			select {
			case limit <- struct{}{}:
				defer func() { <-limit }()
			case <-ctx.Done():
				lookupErrors[i] = ctx.Err()
				return
			}
			results[i], lookupErrors[i] = fetch(ctx, appID, customerID)
		}()
	}
	pending.Wait()
	if err := ctx.Err(); err != nil {
		return "", err
	}

	snapshot := previewMetadataSnapshot{
		Version:    1,
		CapturedAt: time.Now().UTC().Format(time.RFC3339Nano),
		AppID:      appID,
		Customers:  make([]previewMetadataCustomer, 0, len(customerIDs)),
	}
	sharedSet := false
	for i, customerID := range customerIDs {
		result := results[i]
		if lookupErrors[i] != nil || result.Version != 1 || result.AppID != appID || result.Customer.ID != customerID || result.Customer.Status != "ok" || result.Customer.Labels == nil {
			snapshot.Customers = append(snapshot.Customers, previewMetadataCustomer{ID: customerID, Status: "error", Error: "metadata unavailable"})
			continue
		}
		if sharedSet && (snapshot.HideUnmatched != result.HideUnmatched || !reflect.DeepEqual(snapshot.Branches, result.Branches) || !reflect.DeepEqual(snapshot.Pins, result.Pins)) {
			snapshot.Customers = append(snapshot.Customers, previewMetadataCustomer{ID: customerID, Status: "error", Error: "metadata changed during capture"})
			continue
		}
		if !sharedSet {
			snapshot.HideUnmatched = result.HideUnmatched
			snapshot.Branches = result.Branches
			snapshot.Pins = result.Pins
			sharedSet = true
		}
		snapshot.Customers = append(snapshot.Customers, result.Customer)
	}

	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("encode preview metadata: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	if len(encoded) > maxPreviewMetadataBytes {
		return "", fmt.Errorf("preview metadata exceeds the %d byte environment limit", maxPreviewMetadataBytes)
	}
	return encoded, nil
}
