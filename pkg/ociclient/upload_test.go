package ociclient

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestUploadBlobUsesCallerContext(t *testing.T) {
	tempFile, err := os.CreateTemp("", "ociclient-upload-*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte("content")); err != nil {
		t.Fatal(err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatal(err)
	}

	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("request"), "upload")

	oldClient := http.DefaultClient
	defer func() {
		http.DefaultClient = oldClient
	}()

	requests := 0
	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Context().Value(contextKey("request")); got != "upload" {
				t.Fatalf("request context value = %v, want upload", got)
			}

			requests++
			resp := &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     make(http.Header),
				Body:       io.NopCloser(http.NoBody),
			}

			switch requests {
			case 1:
				if req.Method != http.MethodPost {
					t.Fatalf("request %d method = %s, want POST", requests, req.Method)
				}
				resp.Header.Set("Location", "https://registry.example/upload/1")
			case 2:
				if req.Method != http.MethodPatch {
					t.Fatalf("request %d method = %s, want PATCH", requests, req.Method)
				}
				resp.Header.Set("Location", "https://registry.example/upload/1")
			case 3:
				if req.Method != http.MethodPut {
					t.Fatalf("request %d method = %s, want PUT", requests, req.Method)
				}
				resp.StatusCode = http.StatusCreated
			default:
				t.Fatalf("unexpected request %d", requests)
			}

			return resp, nil
		}),
	}

	if _, err := uploadBlob(ctx, tempFile.Name(), "https://registry.example/v2/repo", "token", false, ""); err != nil {
		t.Fatal(err)
	}

	if requests != 3 {
		t.Fatalf("requests = %d, want 3", requests)
	}
}

func TestUploadManifestUsesCallerContext(t *testing.T) {
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("request"), "manifest")

	oldClient := http.DefaultClient
	defer func() {
		http.DefaultClient = oldClient
	}()

	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Context().Value(contextKey("request")); got != "manifest" {
				t.Fatalf("request context value = %v, want manifest", got)
			}
			if req.Method != http.MethodPut {
				t.Fatalf("request method = %s, want PUT", req.Method)
			}

			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(http.NoBody),
			}, nil
		}),
	}

	blob := &Blob{
		Digest:       "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		Size:         7,
		RelativePath: "model.bin",
	}
	configBlob := &Blob{
		Digest: "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		Size:   2,
	}

	if err := uploadManifest(ctx, []*Blob{blob}, configBlob, "https://registry.example/v2/repo", "token", "tag", "."); err != nil {
		t.Fatal(err)
	}
}
