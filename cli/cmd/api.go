package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitAPICommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "api",
		Short:  "Make ad-hoc API calls to the Replicated API",
		Long:   ``,
		Hidden: false,
	}
	parent.AddCommand(cmd)

	return cmd
}

// normalizeAPIPath ensures the path starts with "/" and returns the version
// segment (v1/v2/v3) for validation.
func normalizeAPIPath(raw string) (path string, version string, err error) {
	path = strings.TrimSpace(raw)
	if path == "" {
		return "", "", errors.New("path is required")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	parts := strings.Split(path, "/")
	// Drop empty segments from leading/trailing slashes.
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	if len(cleaned) == 0 {
		return "", "", errors.New("path is required")
	}

	version = cleaned[0]
	switch version {
	case "v1", "v2", "v3":
		return path, version, nil
	default:
		return "", "", errors.Errorf("unsupported API version %q; path must start with /v1, /v2, or /v3 (got %q)", version, path)
	}
}

// doAPIRequest performs a raw JSON HTTP call against the vendor API origin and
// writes the response body to stdout. Supports v1, v2, and v3 paths.
func (r *runners) doAPIRequest(ctx context.Context, method, rawPath, body string) error {
	path, _, err := normalizeAPIPath(rawPath)
	if err != nil {
		return err
	}

	if r.platformAPI == nil {
		return errors.New("API client is not configured")
	}

	response, err := r.platformAPI.DoJSONWithoutUnmarshal(ctx, method, path, body)
	if err != nil {
		return err
	}

	fmt.Printf("%s", response)
	return nil
}
