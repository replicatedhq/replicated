package credentials

import (
	"fmt"
	"strings"

	"github.com/replicatedhq/replicated/pkg/credentials/types"
)

// OriginConfig holds the resolved origins for all services
type OriginConfig struct {
	VendorAPI      string
	VendorWeb      string
	Registry       string
	Linter         string
	KurlSH         string
	UsingNamespace bool
}

// ResolveOrigins resolves all service origins from a profile
// If the profile has a namespace, it generates okteto URLs
// Otherwise, it uses explicit origins or defaults
func ResolveOrigins(profile types.Profile) OriginConfig {
	config := OriginConfig{}

	// If namespace is provided, generate okteto URLs
	if profile.Namespace != "" {
		config.UsingNamespace = true
		config.VendorAPI = fmt.Sprintf("https://vendor-api-%s.okteto.repldev.com", profile.Namespace)
		config.VendorWeb = fmt.Sprintf("https://vendor-web-%s.okteto.repldev.com", profile.Namespace)
		config.Registry = fmt.Sprintf("vendor-registry-v2-%s.okteto.repldev.com", profile.Namespace)
		config.Linter = fmt.Sprintf("https://lint-%s.okteto.repldev.com", profile.Namespace)
		config.KurlSH = "https://kurl.sh" // KurlSH doesn't change for dev envs
		return config
	}

	// Otherwise, use explicit origins or defaults
	config.VendorAPI = getOrDefault(profile.APIOrigin, "https://api.replicated.com/vendor")
	config.VendorWeb = "https://vendor.replicated.com"
	config.Registry = getOrDefault(profile.RegistryOrigin, "registry.replicated.com")
	config.Linter = "https://lint.replicated.com"
	config.KurlSH = "https://kurl.sh"

	return config
}

func getOrDefault(value, defaultValue string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

// ResolveOriginsFromProfileName is a convenience function that loads a profile
// and resolves its origins
func ResolveOriginsFromProfileName(profileName string) (OriginConfig, error) {
	profile, err := GetProfile(profileName)
	if err != nil {
		return OriginConfig{}, err
	}
	return ResolveOrigins(*profile), nil
}
