package platformclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
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

func TestDoJSONWithoutUnmarshalRespectsCanceledContext(t *testing.T) {
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

	_, err := client.DoJSONWithoutUnmarshal(ctx, "GET", "/v3/apps", "")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
}

func TestDoJSONWithoutUnmarshalErrorHandlingAndHeaders(t *testing.T) {
	originalHTTPClient := httpClient
	defer func() {
		httpClient = originalHTTPClient
	}()

	tests := []struct {
		name           string
		statusCode     int
		body           string
		wantErrIs      error
		wantForbidden  bool
		wantAPIError   bool
		wantStatusCode int
	}{
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       `{"message":"missing"}`,
			wantErrIs:  ErrNotFound,
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			body:          `{"error":{"message":"no access"}}`,
			wantForbidden: true,
		},
		{
			name:           "other error",
			statusCode:     http.StatusBadRequest,
			body:           `{"message":"bad request"}`,
			wantAPIError:   true,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq *http.Request
			httpClient = &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					gotReq = r
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.body)),
						Header:     make(http.Header),
					}, nil
				}),
			}

			client := NewHTTPClient("https://example.test", "test-api-key")
			_, err := client.DoJSONWithoutUnmarshal(context.Background(), "GET", "/v3/apps", "")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if gotReq == nil {
				t.Fatal("request was not made")
			}
			if gotReq.Header.Get("Authorization") != "test-api-key" {
				t.Fatalf("expected Authorization header, got %q", gotReq.Header.Get("Authorization"))
			}
			if !strings.HasPrefix(gotReq.Header.Get("User-Agent"), "Replicated/") {
				t.Fatalf("expected User-Agent header, got %q", gotReq.Header.Get("User-Agent"))
			}

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
			}
			if tt.wantForbidden {
				var forbiddenErr ForbiddenError
				if !errors.As(err, &forbiddenErr) {
					t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
				}
				if !errors.Is(err, ErrForbidden) {
					t.Fatalf("expected ErrForbidden unwrap, got %v", err)
				}
			}
			if tt.wantAPIError {
				var apiErr APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantStatusCode {
					t.Fatalf("expected status %d, got %d", tt.wantStatusCode, apiErr.StatusCode)
				}
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
