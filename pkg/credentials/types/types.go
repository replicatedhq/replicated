package types

type Credentials struct {
	// APIToken is the API token used to authenticate with the Replicated API
	APIToken string `json:"token"`

	IsEnv        bool   `json:"-"`
	IsConfigFile bool   `json:"-"`
	IsProfile    bool   `json:"-"`
	ProfileName  string `json:"-"` // Name of the profile that was loaded (if IsProfile is true)
}

// Profile represents a named authentication profile
type Profile struct {
	APIToken       string `json:"apiToken"`
	APIOrigin      string `json:"apiOrigin,omitempty"`
	RegistryOrigin string `json:"registryOrigin,omitempty"`
}

// ConfigFile represents the structure of ~/.replicated/config.yaml
type ConfigFile struct {
	// Legacy single token (for backward compatibility)
	Token string `json:"token,omitempty"`

	// New profile-based configuration
	Profiles       map[string]Profile `json:"profiles,omitempty"`
	DefaultProfile string             `json:"defaultProfile,omitempty"`
}
