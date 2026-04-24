package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

const (
	legacyCacheFileName = "replicated.cache"
	cacheFileNameFormat = "replicated-%s.cache" // profile-specific cache
)

type Cache struct {
	DefaultApp string `json:"default_app"`

	LastAppRefresh *time.Time  `json:"last_app_refresh"`
	Apps           []types.App `json:"apps"`

	// profileName is the profile this cache is associated with
	// Empty string means legacy/no profile
	profileName string
}

// InitCache creates or loads a cache for the given profile name.
// If profileName is empty, uses legacy cache file.
func InitCache(profileName string) (*Cache, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache directory")
	}

	cacheFilePath := getCacheFilePath(cacheDir, profileName)

	// Try to load existing cache
	cache, err := loadCache(cacheFilePath)
	if err != nil {
		// If the file doesn't exist, create a new cache
		if os.IsNotExist(errors.Cause(err)) {
			cache = &Cache{
				Apps:           []types.App{},
				LastAppRefresh: nil,
				DefaultApp:     "",
				profileName:    profileName,
			}

			// save it
			err = cache.Save()
			if err != nil {
				return nil, errors.Wrap(err, "failed to save cache")
			}
		} else {
			return nil, errors.Wrap(err, "failed to load cache")
		}
	} else {
		// Set the profile name on loaded cache
		cache.profileName = profileName
	}

	return cache, nil
}

// getCacheFilePath returns the cache file path for the given profile
func getCacheFilePath(cacheDir, profileName string) string {
	if profileName == "" {
		return filepath.Join(cacheDir, legacyCacheFileName)
	}
	return filepath.Join(cacheDir, filepath.Clean(profileName)+".cache")
}

func (c *Cache) Save() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return errors.Wrap(err, "failed to get cache directory")
	}

	cacheFilePath := getCacheFilePath(cacheDir, c.profileName)

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
func (c *Cache) GetDefault(defaultType string) (string, error) {
	switch defaultType {
	case "app":
		return c.DefaultApp, nil
	default:
		return "", errors.Errorf("unknown default type: %s", defaultType)
	}
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

// ClearAll removes all cached data from the current cache
func (c *Cache) ClearAll() error {
	c.Apps = []types.App{}
	c.DefaultApp = ""
	c.LastAppRefresh = nil
	return c.Save()
}

// DeleteCacheFile deletes the cache file for a specific profile
func DeleteCacheFile(profileName string) error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return errors.Wrap(err, "failed to get cache directory")
	}

	cacheFilePath := getCacheFilePath(cacheDir, profileName)

	// It's not an error if the file doesn't exist
	if err := os.Remove(cacheFilePath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to delete cache file")
	}

	return nil
}

// DeleteAllCacheFiles removes all cache files (for logout)
func DeleteAllCacheFiles() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return errors.Wrap(err, "failed to get cache directory")
	}

	// Read all files in cache directory
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Cache directory doesn't exist, nothing to delete
		}
		return errors.Wrap(err, "failed to read cache directory")
	}

	// Delete all .cache files
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".cache" {
			filePath := filepath.Join(cacheDir, file.Name())
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				// Log but don't fail on individual file deletion errors
				continue
			}
		}
	}

	return nil
}
