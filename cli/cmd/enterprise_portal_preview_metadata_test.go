package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func previewMetadataResult(customerID string) previewMetadataResponse {
	return previewMetadataResponse{
		Version: 1, AppID: "app", Branches: []string{"main"}, Pins: []previewMetadataPin{},
		Customer: previewMetadataCustomer{ID: customerID, Status: "ok", Labels: []string{"1.2.3"}, ChannelID: "stable"},
	}
}

func decodePreviewMetadataEnv(t *testing.T, encoded string) previewMetadataSnapshot {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	var snapshot previewMetadataSnapshot
	require.NoError(t, json.Unmarshal(raw, &snapshot))
	return snapshot
}

func TestCollectPreviewMetadataPreservesCustomerFailures(t *testing.T) {
	encoded, err := collectPreviewMetadata(context.Background(), "app", []string{"a", "b", "c"}, func(_ context.Context, appID, customerID string) (previewMetadataResponse, error) {
		if appID != "app" {
			return previewMetadataResponse{}, fmt.Errorf("unexpected app %q", appID)
		}
		if customerID == "b" {
			return previewMetadataResponse{}, errors.New("private API error")
		}
		return previewMetadataResult(customerID), nil
	})
	require.NoError(t, err)
	snapshot := decodePreviewMetadataEnv(t, encoded)
	require.Equal(t, 1, snapshot.Version)
	require.Equal(t, "app", snapshot.AppID)
	require.NotEmpty(t, snapshot.CapturedAt)
	require.Equal(t, []string{"main"}, snapshot.Branches)
	require.Equal(t, []string{"a", "b", "c"}, []string{snapshot.Customers[0].ID, snapshot.Customers[1].ID, snapshot.Customers[2].ID})
	require.Equal(t, "ok", snapshot.Customers[0].Status)
	require.Equal(t, "error", snapshot.Customers[1].Status)
	require.Nil(t, snapshot.Customers[1].Labels)
	require.Equal(t, "ok", snapshot.Customers[2].Status)
	raw, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "private API error")
}

func TestCollectPreviewMetadataRejectsInconsistentSharedData(t *testing.T) {
	encoded, err := collectPreviewMetadata(context.Background(), "app", []string{"a", "b"}, func(_ context.Context, _, customerID string) (previewMetadataResponse, error) {
		result := previewMetadataResult(customerID)
		if customerID == "b" {
			result.Branches = []string{"other"}
		}
		return result, nil
	})
	require.NoError(t, err)
	snapshot := decodePreviewMetadataEnv(t, encoded)
	require.Equal(t, []string{"main"}, snapshot.Branches)
	require.Equal(t, "ok", snapshot.Customers[0].Status)
	require.Equal(t, "error", snapshot.Customers[1].Status)
}

func TestCollectPreviewMetadataBoundsConcurrencyAndPayload(t *testing.T) {
	ids := make([]string, 12)
	for i := range ids {
		ids[i] = string(rune('a' + i))
	}
	var active, peak atomic.Int32
	encoded, err := collectPreviewMetadata(context.Background(), "app", ids, func(_ context.Context, _, customerID string) (previewMetadataResponse, error) {
		current := active.Add(1)
		for old := peak.Load(); current > old && !peak.CompareAndSwap(old, current); old = peak.Load() {
		}
		time.Sleep(5 * time.Millisecond)
		active.Add(-1)
		return previewMetadataResult(customerID), nil
	})
	require.NoError(t, err)
	require.Len(t, decodePreviewMetadataEnv(t, encoded).Customers, len(ids))
	require.LessOrEqual(t, peak.Load(), int32(4))

	_, err = collectPreviewMetadata(context.Background(), "app", []string{"a"}, func(context.Context, string, string) (previewMetadataResponse, error) {
		result := previewMetadataResult("a")
		result.Customer.Labels = []string{strings.Repeat("x", maxPreviewMetadataBytes)}
		return result, nil
	})
	require.ErrorContains(t, err, "environment limit")
}

func TestBuildPreviewMetadataEnvUsesVendorAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/app/app/enterprise-portal/preview-metadata" || r.URL.Query().Get("customer_id") != "customer/a" {
			t.Errorf("unexpected metadata request: %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(previewMetadataResult("customer/a")); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	oldOrigin := platformOrigin
	platformOrigin = server.URL
	defer func() { platformOrigin = oldOrigin }()

	encoded, err := buildPreviewMetadataEnv(context.Background(), "app", []string{"customer/a"}, "test-token")
	require.NoError(t, err)
	require.Equal(t, "customer/a", decodePreviewMetadataEnv(t, encoded).Customers[0].ID)
}
