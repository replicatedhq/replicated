package platformclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_SendTelemetryEvent(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		client := platformclient.NewHTTPClient(u, "replicated-cli-telemetry-token")

		// Create payload matching telemetry.EventPayload
		exitCode := 0
		durationMs := 1000
		payload := map[string]interface{}{
			"event_id":        "550e8400-e29b-41d4-a716-446655440000",
			"command":         "test command",
			"exit_code":       &exitCode,
			"duration_ms":     &durationMs,
			"has_config_file": false,
		}

		err = client.DoJSON(context.Background(), "POST", "/v3/cli/telemetry/event", 201, payload, nil)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Record CLI command execution event").
		UponReceiving("A request to record CLI telemetry event").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/cli/telemetry/event"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-telemetry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"event_id":        dsl.UUID(),
				"command":         dsl.String("test command"),
				"exit_code":       dsl.Like(0),
				"duration_ms":     dsl.Like(1000),
				"has_config_file": dsl.Like(false),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"event_id": dsl.UUID(),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_SendTelemetryError(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		client := platformclient.NewHTTPClient(u, "replicated-cli-telemetry-token")

		// Create payload matching telemetry.ErrorPayload
		payload := map[string]interface{}{
			"event_id":      "550e8400-e29b-41d4-a716-446655440000",
			"error_type":    "Error",
			"error_message": "command failed",
			"command":       "test command",
		}

		err = client.DoJSON(context.Background(), "POST", "/v3/cli/telemetry/error", 201, payload, nil)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Record CLI error details").
		UponReceiving("A request to record CLI error telemetry").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/cli/telemetry/error"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-telemetry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"event_id":      dsl.UUID(),
				"error_type":    dsl.String("Error"),
				"error_message": dsl.String("command failed"),
				"command":       dsl.String("test command"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"status": dsl.String("ok"),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_SendTelemetryStats(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		client := platformclient.NewHTTPClient(u, "replicated-cli-telemetry-token")

		// Create payload matching backend's StatsParameters
		payload := map[string]interface{}{
			"command":         "test command",
			"helm_charts":     3,
			"k8s_manifests":   45,
			"preflights":      2,
			"support_bundles": 1,
			"tool_versions": map[string]string{
				"helm":    "3.12.0",
				"kubectl": "1.28.0",
			},
		}

		err = client.DoJSON(context.Background(), "POST", "/v3/cli/telemetry/stats", 201, payload, nil)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Record CLI resource statistics").
		UponReceiving("A request to record CLI stats telemetry").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/cli/telemetry/stats"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-telemetry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"command":         dsl.String("test command"),
				"helm_charts":     dsl.Like(3),
				"k8s_manifests":   dsl.Like(45),
				"preflights":      dsl.Like(2),
				"support_bundles": dsl.Like(1),
				"tool_versions": dsl.Like(map[string]string{
					"helm":    "3.12.0",
					"kubectl": "1.28.0",
				}),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"status": dsl.String("ok"),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
