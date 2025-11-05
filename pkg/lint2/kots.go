package lint2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// LintKots lints a KOTS Config file using the kots binary.
//
// The function checks for a local binary override via REPLICATED_KOTS_PATH
// environment variable (for development), falling back to the resolver for
// production use.
//
// The kots binary supports the following output:
//   - JSON format via --output json flag
//   - Exit code 0 for success, non-zero for validation failures
//   - Mixed text/JSON output (JSON extraction required)
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - configPath: Absolute path to the KOTS Config file
//   - kotsVersion: Version of kots binary to use (e.g., "latest", "1.128.3")
//
// Returns:
//   - LintResult with Success flag and list of LintMessages
//   - Error if linting cannot be performed (binary not found, file not found, etc.)
func LintKots(ctx context.Context, configPath string, kotsVersion string) (*LintResult, error) {
	// Defensive check: validate config path exists (before resolving binary)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("kots config path does not exist: %s", configPath)
		}
		return nil, fmt.Errorf("failed to access kots config path: %w", err)
	}

	// Check for local binary override (for development)
	// TODO: Remove REPLICATED_KOTS_PATH environment variable support
	// once kots linter is stable and published to production releases.
	kotsPath := os.Getenv("REPLICATED_KOTS_PATH")
	if kotsPath == "" {
		// Use resolver to get kots binary
		resolver := tools.NewResolver()
		var err error
		kotsPath, err = resolver.Resolve(ctx, tools.ToolKots, kotsVersion)
		if err != nil {
			return nil, fmt.Errorf("resolving kots: %w", err)
		}
	}

	// Build command arguments
	args := []string{"lint", "--output", "json", configPath}

	// Execute kots lint
	cmd := exec.CommandContext(ctx, kotsPath, args...)
	output, err := cmd.CombinedOutput()

	// kots lint returns non-zero exit code if there are validation errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := parseKotsOutput(outputStr)
	if parseErr != nil {
		// If we can't parse the output, return both the parse error and original error
		if err != nil {
			return nil, fmt.Errorf("kots lint failed and output parsing failed: %w\nParse error: %v\nOutput: %s", err, parseErr, outputStr)
		}
		return nil, fmt.Errorf("failed to parse kots lint output: %w\nOutput: %s", parseErr, outputStr)
	}

	// Determine success based on exit code
	// Exit code 0 = no errors, non-zero = validation errors
	success := err == nil

	return &LintResult{
		Success:  success,
		Messages: messages,
	}, nil
}

// KotsLintOutput represents the JSON output from kots lint
type KotsLintOutput struct {
	LintExpressions   []KotsLintExpression `json:"lintExpressions"`
	IsLintingComplete bool                 `json:"isLintingComplete"`
}

type KotsLintExpression struct {
	Rule      string             `json:"rule"`
	Type      string             `json:"type"` // "error", "warn", "info"
	Message   string             `json:"message"`
	Path      string             `json:"path"`
	Positions []KotsLintPosition `json:"positions"`
}

type KotsLintPosition struct {
	Start KotsLintLinePosition `json:"start"`
}

type KotsLintLinePosition struct {
	Line int64 `json:"line"`
}

// parseKotsOutput parses the JSON output from kots lint command.
// Handles mixed text/JSON output by extracting JSON first.
func parseKotsOutput(output string) ([]LintMessage, error) {
	if output == "" {
		return []LintMessage{}, nil
	}

	// Extract clean JSON from output that may contain error messages
	// (e.g., "ERROR: validation failed" text before/after JSON)
	jsonStr, err := extractJSONFromOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from output: %w", err)
	}

	// Decode the JSON into the kots result structure
	var kotsOutput KotsLintOutput
	if err := json.Unmarshal([]byte(jsonStr), &kotsOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON output: %w", err)
	}

	var messages []LintMessage

	// Process each lint expression
	for _, expr := range kotsOutput.LintExpressions {
		msg := formatKotsMessage(expr)
		messages = append(messages, LintMessage{
			Severity: normalizeKotsSeverity(expr.Type),
			Message:  msg,
			Path:     expr.Path,
		})
	}

	return messages, nil
}

// formatKotsMessage formats a KOTS lint expression into a readable message.
// Includes line number (if available) and rule name.
func formatKotsMessage(expr KotsLintExpression) string {
	msg := expr.Message

	// Add line number if positions available
	if len(expr.Positions) > 0 && expr.Positions[0].Start.Line > 0 {
		msg = fmt.Sprintf("line %d: %s", expr.Positions[0].Start.Line, msg)
	}

	// Add rule name if available
	if expr.Rule != "" {
		msg = fmt.Sprintf("[%s] %s", expr.Rule, msg)
	}

	return msg
}

// normalizeKotsSeverity converts KOTS severity types to standard severity levels.
// Returns uppercase values to match the CLI's internal representation and summary calculator.
// Performs case-insensitive matching to handle variations in casing from the binary output.
func normalizeKotsSeverity(kotsType string) string {
	switch strings.ToLower(kotsType) {
	case "error":
		return "ERROR"
	case "warn":
		return "WARNING"
	case "info":
		return "INFO"
	default:
		return "INFO" // Default fallback
	}
}
