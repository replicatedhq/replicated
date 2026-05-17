package platformclient

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestGetFeaturesRespectsCanceledContext(t *testing.T) {
	originalHTTPClient := httpClient
	defer func() {
		httpClient = originalHTTPClient
	}()

	httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			select {
			case <-r.Context().Done():
				return nil, r.Context().Err()
			case <-time.After(20 * time.Millisecond):
				return nil, errors.New("request context was not canceled")
			}
		}),
	}

	client := NewHTTPClient("https://example.test", "test-api-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetFeatures(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
