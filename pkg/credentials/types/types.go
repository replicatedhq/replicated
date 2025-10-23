package types

type Credentials struct {
	// APIToken is the API token used to authenticate with the Replicated API
	APIToken string `json:"token"`

	IsEnv        bool `json:"-"`
	IsConfigFile bool `json:"-"`
	IsProfile    bool `json:"-"`
}

// Profile represents a named authentication profile
type Profile struct {
	APIToken       string `json:"apiToken"`
	APIOrigin      string `json:"apiOrigin,omitempty"`
	RegistryOrigin string `json:"registryOrigin,omitempty"`
	// Namespace is used for okteto dev environments to auto-generate service URLs
	// e.g., namespace="noahecampbell" generates:
	//   - vendor-api-noahecampbell.okteto.repldev.com
	//   - vendor-web-noahecampbell.okteto.repldev.com
	//   - etc.
	Namespace string `json:"namespace,omitempty"`
}

// ConfigFile represents the structure of ~/.replicated/config.yaml
type ConfigFile struct {
	// Legacy single token (for backward compatibility)
	Token string `json:"token,omitempty"`

	// New profile-based configuration
	Profiles       map[string]Profile `json:"profiles,omitempty"`
	DefaultProfile string             `json:"defaultProfile,omitempty"`
}
