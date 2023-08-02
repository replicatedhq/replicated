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

func GetCurrentCredentials() (*types.Credentials, error) {
	// priority order:
	// 1. env vars
	// 2. config file

	envCredentials, err := getEnvCredentials()
	if err != nil && err != ErrCredentialsNotFound {
		return nil, err
	}
	if err == nil {
		return envCredentials, nil
	}

	configFileCredentials, err := getConfigFileCredentials()
	if err != nil && err != ErrCredentialsNotFound {
		return nil, err
	}

	if err == nil {
		return configFileCredentials, nil
	}

	return nil, ErrCredentialsNotFound
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

func configFilePath() string {
	return filepath.Join(homeDir(), ".replicated", "config.yaml")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
