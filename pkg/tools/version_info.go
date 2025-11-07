package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

// FetchRecommendedVersions fetches recommended tool versions from replicated.app/ping
// Returns a map of tool name to recommended version, or error if the API call fails.
// Tool names are normalized (e.g., "support-bundle" → "support_bundle", "embedded-cluster" → "embedded_cluster")
func FetchRecommendedVersions() (map[string]string, error) {
	url := "https://replicated.app/ping"

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching versions from replicated.app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("replicated.app/ping returned HTTP %d", resp.StatusCode)
	}

	var pingResp replicatedPingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
		return nil, fmt.Errorf("parsing ping response JSON: %w", err)
	}

	// Return the client_versions map directly
	// Note: Keys in the response use underscores (support_bundle, embedded_cluster)
	return pingResp.ClientVersions, nil
}

// CompareVersions compares a configured version against a recommended version.
// Returns true if the configured version is outdated by a MINOR or MAJOR version (ignores patch differences).
//
// Examples:
//   - "3.14.4" vs "3.14.5" → false (patch difference, no warning)
//   - "3.14.0" vs "3.15.0" → true (minor difference, warn)
//   - "3.0.0" vs "4.0.0" → true (major difference, warn)
//   - "3.16.0" vs "3.15.0" → false (ahead of recommended, no warning)
//
// IMPORTANT: Caller should check if configured is "latest" before calling this function.
// Returns false on parse errors (silent failure - config validation catches invalid semver).
func CompareVersions(configured, recommended string) bool {
	// Normalize versions by stripping 'v' prefix
	configured = strings.TrimPrefix(configured, "v")
	recommended = strings.TrimPrefix(recommended, "v")

	// Parse versions using semver
	configuredVer, err := semver.NewVersion(configured)
	if err != nil {
		// Silent failure - config validation should catch invalid semver
		return false
	}

	recommendedVer, err := semver.NewVersion(recommended)
	if err != nil {
		// Silent failure - if recommended version is invalid, skip check
		return false
	}

	// Return false if configured >= recommended (no warning needed)
	if !configuredVer.LessThan(recommendedVer) {
		return false
	}

	// At this point, configured < recommended
	// Warn only if major or minor version differs (ignore patch differences)

	if configuredVer.Major() < recommendedVer.Major() {
		// Major version behind
		return true
	}

	if configuredVer.Major() == recommendedVer.Major() && configuredVer.Minor() < recommendedVer.Minor() {
		// Minor version behind (same major)
		return true
	}

	// Patch-only difference (or prerelease/metadata) - don't warn
	return false
}
