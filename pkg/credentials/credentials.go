package credentials

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/replicatedhq/replicated/pkg/credentials/types"
)

var (
	ErrCredentialsNotFound = errors.New("credentials not found")
	ErrProfileNotFound     = errors.New("profile not found")
)

func SetCurrentCredentials(token string) error {
	configFileCredentials := &types.Credentials{
		APIToken: token,
	}

	b, err := json.Marshal(configFileCredentials)
	if err != nil {
		return err
	}

	configFile := configFilePath()
	if err := os.MkdirAll(path.Dir(configFile), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configFile, b, 0600); err != nil {
		return err
	}

	return nil
}

func RemoveCurrentCredentials() error {
	configFile := configFilePath()
	if err := os.Remove(configFile); err != nil {
		return err
	}
	return nil
}

func GetCurrentCredentials() (*types.Credentials, error) {
	return GetCredentialsWithProfile("")
}

// GetCredentialsWithProfile retrieves credentials with the following priority:
// 1. Environment variables (REPLICATED_API_TOKEN)
// 2. Named profile (if profileName is provided)
// 3. Default profile from config file (if profileName is empty)
// 4. Legacy single token from config file (backward compatibility)
func GetCredentialsWithProfile(profileName string) (*types.Credentials, error) {
	// Priority 1: Check environment variables first
	envCredentials, err := getEnvCredentials()
	if err != nil && err != ErrCredentialsNotFound {
		return nil, err
	}
	if err == nil {
		return envCredentials, nil
	}

	// Priority 2 & 3: Check profile-based credentials
	profileCredentials, err := getProfileCredentials(profileName)
	if err != nil && err != ErrCredentialsNotFound && err != ErrProfileNotFound {
		return nil, err
	}
	if err == nil {
		return profileCredentials, nil
	}

	// Priority 4: Fall back to legacy config file credentials
	configFileCredentials, err := getConfigFileCredentials()
	if err != nil && err != ErrCredentialsNotFound {
		return nil, err
	}
	if err == nil {
		return configFileCredentials, nil
	}

	return nil, ErrCredentialsNotFound
}

// getProfileCredentials retrieves credentials from a named profile
// If profileName is empty, uses the default profile
func getProfileCredentials(profileName string) (*types.Credentials, error) {
	config, err := readConfigFile()
	if err != nil {
		return nil, err
	}

	// If no profile name provided, use default
	if profileName == "" {
		profileName = config.DefaultProfile
	}

	// If still no profile name, return not found
	if profileName == "" {
		return nil, ErrProfileNotFound
	}

	// Get the profile
	profile, exists := config.Profiles[profileName]
	if !exists {
		return nil, ErrProfileNotFound
	}

	// Validate that profile has a token
	if profile.APIToken == "" {
		return nil, errors.New("profile has no API token")
	}

	return &types.Credentials{
		APIToken:    profile.APIToken,
		IsProfile:   true,
		ProfileName: profileName, // Store the actual profile name that was used
	}, nil
}

// GetProfileOrigins returns the API and registry origins for a given profile
// Returns empty strings if profile doesn't exist or doesn't specify origins
func GetProfileOrigins(profileName string) (apiOrigin, registryOrigin string, err error) {
	if profileName == "" {
		return "", "", nil
	}

	profile, err := GetProfile(profileName)
	if err != nil {
		return "", "", err
	}

	return profile.APIOrigin, profile.RegistryOrigin, nil
}

func getEnvCredentials() (*types.Credentials, error) {
	if os.Getenv("REPLICATED_API_TOKEN") != "" {
		return &types.Credentials{
			APIToken: os.Getenv("REPLICATED_API_TOKEN"),
			IsEnv:    true,
		}, nil
	}

	return nil, ErrCredentialsNotFound
}

func getConfigFileCredentials() (*types.Credentials, error) {
	configFile := configFilePath()
	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCredentialsNotFound
		}

		return nil, err
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	credentials := types.Credentials{}
	if err := json.Unmarshal(b, &credentials); err != nil {
		return nil, err
	}
	credentials.IsConfigFile = true

	return &credentials, nil
}

// Profile management functions

// readConfigFile reads the config file and returns the parsed ConfigFile struct
func readConfigFile() (*types.ConfigFile, error) {
	configFile := configFilePath()
	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &types.ConfigFile{
				Profiles: make(map[string]types.Profile),
			}, nil
		}
		return nil, err
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config types.ConfigFile
	if err := json.Unmarshal(b, &config); err != nil {
		// Try legacy format (just a Credentials struct)
		var legacyCreds types.Credentials
		if legacyErr := json.Unmarshal(b, &legacyCreds); legacyErr == nil && legacyCreds.APIToken != "" {
			// Convert legacy format to new format
			return &types.ConfigFile{
				Token:    legacyCreds.APIToken,
				Profiles: make(map[string]types.Profile),
			}, nil
		}
		return nil, err
	}

	// Initialize profiles map if nil
	if config.Profiles == nil {
		config.Profiles = make(map[string]types.Profile)
	}

	return &config, nil
}

// writeConfigFile writes the ConfigFile struct to the config file
func writeConfigFile(config *types.ConfigFile) error {
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configFile := configFilePath()
	if err := os.MkdirAll(path.Dir(configFile), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configFile, b, 0600); err != nil {
		return err
	}

	return nil
}

// AddProfile adds or updates a profile in the config file
func AddProfile(name string, profile types.Profile) error {
	if name == "" {
		return errors.New("profile name cannot be empty")
	}

	config, err := readConfigFile()
	if err != nil {
		return err
	}

	config.Profiles[name] = profile

	// Set as default if it's the first profile
	if config.DefaultProfile == "" {
		config.DefaultProfile = name
	}

	return writeConfigFile(config)
}

// RemoveProfile removes a profile from the config file
func RemoveProfile(name string) error {
	config, err := readConfigFile()
	if err != nil {
		return err
	}

	if _, exists := config.Profiles[name]; !exists {
		return ErrProfileNotFound
	}

	delete(config.Profiles, name)

	// Clear default if it was the removed profile
	if config.DefaultProfile == name {
		config.DefaultProfile = ""
		// Set to first available profile if any exist
		for profileName := range config.Profiles {
			config.DefaultProfile = profileName
			break
		}
	}

	return writeConfigFile(config)
}

// GetProfile retrieves a specific profile by name
func GetProfile(name string) (*types.Profile, error) {
	config, err := readConfigFile()
	if err != nil {
		return nil, err
	}

	profile, exists := config.Profiles[name]
	if !exists {
		return nil, ErrProfileNotFound
	}

	return &profile, nil
}

// ListProfiles returns all profiles and the default profile name
func ListProfiles() (map[string]types.Profile, string, error) {
	config, err := readConfigFile()
	if err != nil {
		return nil, "", err
	}

	return config.Profiles, config.DefaultProfile, nil
}

// SetDefaultProfile sets the default profile
func SetDefaultProfile(name string) error {
	config, err := readConfigFile()
	if err != nil {
		return err
	}

	if _, exists := config.Profiles[name]; !exists {
		return ErrProfileNotFound
	}

	config.DefaultProfile = name
	return writeConfigFile(config)
}

// GetDefaultProfile returns the name of the default profile
func GetDefaultProfile() (string, error) {
	config, err := readConfigFile()
	if err != nil {
		return "", err
	}

	return config.DefaultProfile, nil
}

func configFilePath() string {
	return filepath.Join(homeDir(), ".replicated", "config.yaml")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
