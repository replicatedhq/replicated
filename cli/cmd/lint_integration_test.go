//go:build integration
// +build integration

package cmd

import (
	"context"
	"os"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/tools"
)

// TestResolveAppContext_Integration tests app resolution with real API calls.
// Run with: go test -tags=integration -v ./cli/cmd -run TestResolveAppContext
//
// Prerequisites:
// - REPLICATED_API_TOKEN environment variable set
// - REPLICATED_API_ORIGIN (optional, defaults to https://api.replicated.com)
// - At least one app in the vendor account
func TestResolveAppContext_Integration(t *testing.T) {
	// Check for required credentials
	apiToken := os.Getenv("REPLICATED_API_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping integration test: REPLICATED_API_TOKEN not set")
	}

	apiOrigin := os.Getenv("REPLICATED_API_ORIGIN")
	if apiOrigin == "" {
		apiOrigin = "https://api.replicated.com"
	}

	// Set up real kotsAPI client
	httpClient := platformclient.NewHTTPClient(apiOrigin, apiToken)
	kotsAPI := kotsclient.VendorV3Client{HTTPClient: *httpClient}

	// Create runners instance with real API client
	r := &runners{
		kotsAPI: &kotsAPI,
		w:       tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', 0),
	}

	ctx := context.Background()

	t.Run("resolve with no app specified (API fetch)", func(t *testing.T) {
		config := &tools.Config{}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		// Should either:
		// - Return empty string (0 apps available)
		// - Return an app ID (1 app auto-selected)
		// - Return error (multiple apps, non-TTY)
		t.Logf("Resolved app ID: %q", appID)

		if appID != "" {
			// Verify it's a valid app ID by fetching it
			app, err := kotsAPI.GetApp(ctx, appID, true)
			if err != nil {
				t.Errorf("GetApp(%q) failed: %v", appID, err)
			}
			if app == nil || app.ID == "" {
				t.Errorf("GetApp returned invalid app: %+v", app)
			}
			t.Logf("Successfully resolved to app: %s (slug: %s)", app.Name, app.Slug)
		} else {
			t.Log("No app resolved (0 apps or API failure)")
		}
	})

	t.Run("resolve with explicit app ID in config", func(t *testing.T) {
		// First, list apps to get a valid app ID
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) == 0 {
			t.Skip("Skipping: no apps available in vendor account")
		}

		testAppID := apps[0].App.ID

		config := &tools.Config{
			AppId: testAppID,
		}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != testAppID {
			t.Errorf("Expected appID = %q, got %q", testAppID, appID)
		}
	})

	t.Run("resolve with explicit app slug in config", func(t *testing.T) {
		// First, list apps to get a valid app slug
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) == 0 {
			t.Skip("Skipping: no apps available in vendor account")
		}

		testAppSlug := apps[0].App.Slug
		expectedAppID := apps[0].App.ID

		config := &tools.Config{
			AppSlug: testAppSlug,
		}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != expectedAppID {
			t.Errorf("Expected appID = %q, got %q", expectedAppID, appID)
		}
	})

	t.Run("resolve with runners.appID set", func(t *testing.T) {
		// First, list apps to get a valid app ID
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) == 0 {
			t.Skip("Skipping: no apps available in vendor account")
		}

		testAppID := apps[0].App.ID

		// Set app ID in runners (simulates --app-id flag)
		r.appID = testAppID
		defer func() { r.appID = "" }() // cleanup

		config := &tools.Config{}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != testAppID {
			t.Errorf("Expected appID = %q, got %q", testAppID, appID)
		}
	})

	t.Run("resolve with invalid app ID should error", func(t *testing.T) {
		config := &tools.Config{
			AppId: "invalid-app-id-does-not-exist",
		}

		_, err := r.resolveAppContext(ctx, config)
		if err == nil {
			t.Error("Expected error when resolving invalid app ID, got nil")
		}
		t.Logf("Got expected error: %v", err)
	})

	t.Run("priority order: runners.appID > config.appId", func(t *testing.T) {
		// Get two different apps (if available)
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) < 1 {
			t.Skip("Skipping: need at least 1 app for priority test")
		}

		runnerAppID := apps[0].App.ID
		configAppID := "different-app-id" // doesn't matter, runner should win

		r.appID = runnerAppID
		defer func() { r.appID = "" }()

		config := &tools.Config{
			AppId: configAppID,
		}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != runnerAppID {
			t.Errorf("Priority test failed: expected runners.appID %q to win, got %q", runnerAppID, appID)
		}
	})
}
