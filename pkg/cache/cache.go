package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

const cacheFileName = "replicated.cache"

type Cache struct {
	DefaultApp string `json:"default_app"`

	LastAppRefresh *time.Time  `json:"last_app_refresh"`
	Apps           []types.App `json:"apps"`
}

func InitCache() (*Cache, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache directory")
	}

	cacheFilePath := filepath.Join(cacheDir, cacheFileName)

	// Try to load existing cache
	cache, err := loadCache(cacheFilePath)
	if err != nil {
		// If the file doesn't exist, create a new cache
		if os.IsNotExist(errors.Cause(err)) {
			cache = &Cache{
				Apps:           []types.App{},
				LastAppRefresh: nil,
				DefaultApp:     "",
			}

			// save it
			err = cache.Save()
			if err != nil {
				return nil, errors.Wrap(err, "failed to save cache")
			}
		} else {
			return nil, errors.Wrap(err, "failed to load cache")
		}
	}

	return cache, nil
}

func (c *Cache) Save() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return errors.Wrap(err, "failed to get cache directory")
	}

	cacheFilePath := filepath.Join(cacheDir, cacheFileName)

	data, err := json.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "failed to marshal cache data")
	}

	err = os.MkdirAll(filepath.Dir(cacheFilePath), 0755)
	if err != nil {
		return errors.Wrap(err, "failed to create cache directory")
	}

	err = os.WriteFile(cacheFilePath, data, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write cache file")
	}

	return nil
}

func loadCache(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cache file")
	}

	var cache Cache
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal cache data")
	}

	return &cache, nil
}

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}

	return filepath.Join(homeDir, ".replicated", "cache"), nil
}

func (c *Cache) SetDefault(defaultType string, defaultValue string) error {
	switch defaultType {
	case "app":
		c.DefaultApp = defaultValue
	default:
		return errors.Errorf("unknown default type: %s", defaultType)
	}

	if err := c.Save(); err != nil {
		return errors.Wrap(err, "failed to save cache")
	}

	return nil
}

func (c *Cache) ClearDefault(defaultType string) error {
	switch defaultType {
	case "app":
		c.DefaultApp = ""
	default:
		return errors.Errorf("unknown default type: %s", defaultType)
	}

	if err := c.Save(); err != nil {
		return errors.Wrap(err, "failed to save cache")
	}

	return nil
}
