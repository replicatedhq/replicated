package lint2

import (
	"context"
	"fmt"
	"os"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// resolveLinterBinary resolves the path to a linter binary, checking for
// environment variable override first, then falling back to the resolver.
//
// This function provides a unified way for all linters to resolve their
// binaries with consistent support for local development overrides.
//
// Parameters:
//   - ctx: Context for cancellation
//   - toolName: Tool constant (e.g., tools.ToolKots, tools.ToolHelm)
//   - version: Version to resolve (e.g., "latest", "1.128.3")
//   - envVarName: Environment variable name for local override (e.g., "REPLICATED_KOTS_PATH")
//
// Returns:
//   - Binary path (either from env var or resolver)
//   - Error if resolution fails
//
// Environment Variable Overrides (for local development):
//   - REPLICATED_HELM_PATH
//   - REPLICATED_PREFLIGHT_PATH
//   - REPLICATED_SUPPORT_BUNDLE_PATH
//   - REPLICATED_EMBEDDED_CLUSTER_PATH
//   - REPLICATED_KOTS_PATH
func resolveLinterBinary(
	ctx context.Context,
	toolName string,
	version string,
	envVarName string,
) (string, error) {
	// Check for local binary override (for development)
	if path := os.Getenv(envVarName); path != "" {
		return path, nil
	}

	// Use resolver to get binary from releases
	resolver := tools.NewResolver()
	path, err := resolver.Resolve(ctx, toolName, version)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", toolName, err)
	}

	return path, nil
}
